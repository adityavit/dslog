package log

import (
	"io"
	"os"
)
import "github.com/tysonmote/gommap"

var (
	offWidth uint64 = 4                   // 4 bytes
	posWidth uint64 = 8                   // 8 bytes
	entWidth        = offWidth + posWidth // 12 bytes
)

type index struct {
	file *os.File    // indexed file
	mmap gommap.MMap // Memory mapped file of the indexed file
	size uint64      // size of the index, where to add the next index
}

func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size())
	// Why the truncation of the file is done?
	// Growing the index file to max size before memory mapping the file
	// Increase the size of the file to the MaxIndexBytes, as once memory mapped, size cannot be changed.
	if err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}
	if idx.mmap, err = gommap.Map(idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED); err != nil {
		return nil, err
	}
	return idx, nil
}

func (i *index) Close() error {
	// Flush the memory map of the file; Flushing is done synchronously with MS_SYNC flag
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}
	// Flush the file from the memory
	if err := i.file.Sync(); err != nil {
		return err
	}
	// Truncate it back to the size of the content of the file.
	// Remove the extra empty bytes added at the end of the file.
	// The last entWidth bytes should be the last record in the index
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close()
}

// Read the index for the entry idx
func (i *index) Read(idx int64) (offset uint32, pos uint64, err error) {
	// Calculate position of the record first, if idx < 0 is passed get the last entry.
	// index are 0 indexed relative to the segment base for 32 bits. Getting 2^32 entries. i.e. 1 B entries per index.
	// Along with the index value there are 8 bytes for the position of the record data in the store.
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	if idx < 0 {
		offset = uint32(i.size/entWidth) - 1
	} else {
		offset = uint32(idx)
	}
	// pos is index of the entry times the size of each entry.
	pos = uint64(offset) * entWidth
	// Check if there are enough bytes for the entry.
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	// Get the bytes and decode the bytes to the offset and position.
	offset = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
	return offset, pos, nil
}

// Append the pos to the index at the end of the file.
func (i *index) Write(offset uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF
	}
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], offset)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
	i.size += entWidth
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
