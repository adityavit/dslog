package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8 // Size of 8 bytes 64 bits to store the length of the record data, before the actual byte data of record
)

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := fi.Size()
	return &store{
		File: f,
		size: uint64(size),
		buf:  bufio.NewWriter(f),
	}, nil
}

func (s *store) Append(b []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// pos the start of the record
	pos = s.size
	// Store the length of the record bytes in the butter first in enc byte order.
	if err = binary.Write(s.buf, enc, uint64(len(b))); err != nil {
		return 0, 0, err
	}
	// write the record bytes in buffer
	w, err := s.buf.Write(b)
	if err != nil {
		return 0, 0, err
	}
	w += lenWidth
	s.size += uint64(w)
	return uint64(w), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {

}
