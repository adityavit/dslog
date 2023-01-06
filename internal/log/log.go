package log

import (
	"fmt"
	api "github.com/adityavit/dslog/api/v1"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Log struct {
	mu            sync.RWMutex
	Dir           string
	Config        Config
	activeSegment *segment
	segments      []*segment
}

func newLog(dir string, c Config) (*Log, error) {
	if c.Segment.MaxStoreBytes == 0 {
		c.Segment.MaxStoreBytes = 1024
	}
	if c.Segment.MaxIndexBytes == 0 {
		c.Segment.MaxIndexBytes = 1024
	}
	l := &Log{
		Dir:    dir,
		Config: c,
	}
	return l, l.Setup()
}

func (l *Log) Setup() error {
	entries, err := os.ReadDir(l.Dir)
	if err != nil {
		return err
	}
	var baseOffsets []uint64
	// Take the name of each of the file and then append it to the baseOffsets
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), indexExt) {
			offStr := strings.TrimSuffix(path.Base(entry.Name()), indexExt)
			off, _ := strconv.ParseUint(offStr, 10, 0)
			baseOffsets = append(baseOffsets, off)
		}
	}
	sort.Slice(baseOffsets, func(i, j int) bool {
		return baseOffsets[i] < baseOffsets[j]
	})
	for _, off := range baseOffsets {
		if err := l.newSegment(off); err != nil {
			return err
		}
	}
	if len(l.segments) == 0 {
		if err := l.newSegment(l.Config.Segment.InitialOffset); err != nil {
			return err
		}
	}
	return nil
}

func (l *Log) Append(record *api.Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	offset, err := l.activeSegment.Append(record)
	if err != nil {
		return 0, err
	}
	if l.activeSegment.IsMaxed() {
		err = l.newSegment(offset + 1)
	}
	return offset, err
}

func (l *Log) Read(offset uint64) (*api.Record, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var s *segment
	// find the segment in which the offset is present
	for _, seg := range l.segments {
		if seg.baseOffset <= offset && offset < seg.nextOffset {
			s = seg
			break
		}
	}
	// If there is no segment found with the offset within the segment
	if s == nil || s.nextOffset <= offset {
		return nil, fmt.Errorf("log offset not found")
	}
	return s.Read(offset)
}

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, seg := range l.segments {
		if err := seg.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (l *Log) Remove() error {
	if err := l.Close(); err != nil {
		return err
	}
	return os.RemoveAll(l.Dir)
}

func (l *Log) Reset() error {
	if err := l.Remove(); err != nil {
		return err
	}
	return l.Setup()
}

func (l *Log) LowestOffset() (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if len(l.segments) > 0 {
		return l.segments[0].baseOffset, nil
	}
	return l.Config.Segment.InitialOffset, nil
}

func (l *Log) HighestOffset() (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	off := uint64(0)
	if len(l.segments) > 0 {
		off = l.segments[len(l.segments)-1].nextOffset
	}
	if off == 0 {
		return 0, nil
	}
	return off - 1, nil
}

// Truncate removes all the segments from the log with last offset lower than the given offset
func (l *Log) Truncate(offset uint64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	var segments []*segment
	for _, seg := range l.segments {
		if seg.nextOffset <= offset+1 {
			if err := seg.Remove(); err != nil {
				return err
			}
			continue
		}
		segments = append(segments, seg)
	}
	l.segments = segments
	return nil
}

func (l *Log) newSegment(offset uint64) error {
	s, err := newSegment(l.Dir, offset, l.Config)
	if err != nil {
		return err
	}
	l.segments = append(l.segments, s)
	l.activeSegment = s
	return nil
}

func (l *Log) Reader() io.Reader {
	l.mu.RLock()
	defer l.mu.RUnlock()
	readers := make([]io.Reader, len(l.segments))
	for i, seg := range l.segments {
		readers[i] = &originReader{
			store: seg.store,
			off:   0,
		}
	}
	return io.MultiReader(readers...)
}

type originReader struct {
	*store
	off int64
}

func (r *originReader) Read(p []byte) (int, error) {
	n, err := r.ReadAt(p, r.off)
	r.off += int64(n)
	return n, err
}
