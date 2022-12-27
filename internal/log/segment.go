package log

import (
	"fmt"
	api "github.com/adityavit/dslog/api/v1"
	"google.golang.org/protobuf/proto"
	"os"
	"path"
)

type segment struct {
	store                  *store
	index                  *index
	baseOffset, nextOffset uint64
	config                 Config
}

const (
	storeExt = ".store"
	indexExt = ".index"
)

func newSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}
	storeFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, storeExt)),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}
	if s.store, err = newStore(storeFile); err != nil {
		return nil, err
	}
	indexFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, indexExt)),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}
	if s.index, err = newIndex(indexFile, c); err != nil {
		return nil, err
	}
	if off, _, err := s.index.Read(-1); err != nil {
		s.nextOffset = baseOffset
	} else {
		s.nextOffset = s.baseOffset + uint64(off) + 1
	}
	return s, nil
}

// Append
// writes the record at the current offset defined by nextOffset - baseOffset
// returns the offset at which recrd indes is added
func (s *segment) Append(r *api.Record) (offset uint64, err error) {
	currOffset := s.nextOffset
	r.Offset = currOffset
	recBytes, err := proto.Marshal(r)
	if err != nil {
		return 0, err
	}
	// Append the record to the store
	// return thw position of the record
	_, pos, err := s.store.Append(recBytes)
	if err != nil {
		return 0, err
	}
	// writes the store position at the offset
	if err = s.index.Write(uint32(currOffset-s.baseOffset), pos); err != nil {
		return 0, err
	}
	s.nextOffset++
	return currOffset, nil
}

// Read
// Gets the position of the record from the index stored at the offset
// Use the position to read the record bytes yfrom the store
// unmarshal the record and return
func (s *segment) Read(offset uint64) (*api.Record, error) {
	_, pos, err := s.index.Read(int64(offset - s.baseOffset))
	if err != nil {
		return nil, err
	}
	recByte, err := s.store.Read(pos)
	if err != nil {
		return nil, err
	}
	rec := &api.Record{}
	err = proto.Unmarshal(recByte, rec)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// IsMaxed
// Checks if the store or the index size is greater than the Store or Index max bytes
func (s *segment) IsMaxed() bool {
	return s.store.size >= s.config.Segment.MaxStoreBytes || s.index.size >= s.config.Segment.MaxIndexBytes
}

func (s *segment) Remove() error {
	if err := s.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.store.Name()); err != nil {
		return err
	}
	if err := os.Remove(s.index.Name()); err != nil {
		return err
	}
	return nil
}

func (s *segment) Close() error {
	if err := s.store.Close(); err != nil {
		return err
	}
	if err := s.index.Close(); err != nil {
		return err
	}
	return nil
}

// nearestMultiple
// lower nearest Multiple of k closet to j i.e. nearestMultiple(9, 4) ==> 8 as 4*2 = 8 which is closest to 9.
// lower multiple is taken to save disk capacity
func nearestMultiple(j, k uint64) uint64 {
	if j >= 0 {
		// this will always be true as j and k are unsigned integers so are always positive.
		return (j / k) * k
	}
	// Will never reach here.
	return ((j - k + 1) / k) * k
}
