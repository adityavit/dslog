package log

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	write = []byte("hello world")
	width = uint64(len(write)) + lenWidth
)

func TestStoreAppendRead(t *testing.T) {
	f, err := os.CreateTemp("", "store_append_read_test")
	assert.Nil(t, err)
	defer os.Remove(f.Name())
	s, err := newStore(f)
	assert.Nil(t, err)
	assert.NotNil(t, s, "Store should not be nil")
	testAppend(t, s)
	testRead(t, s)
	testReadAt(t, s)
}

func testAppend(t *testing.T, s *store) {
	t.Helper()
	n, pos, err := s.Append(write)
	assert.Nil(t, err)
	assert.Equal(t, pos, uint64(0), "doesn't start with first position")
	assert.Equal(t, n, width, "doesn't match the width")
}

func testRead(t *testing.T, s *store) {
	t.Helper()
	data, err := s.Read(0)
	assert.Nil(t, err)
	assert.Equal(t, data, write)
}

func testReadAt(t *testing.T, s *store) {
	t.Helper()
	// read the width first
	lenByte := make([]byte, lenWidth)
	n, err := s.ReadAt(lenByte, 0)
	assert.Nil(t, err)
	assert.Equal(t, n, lenWidth)
	assert.Equal(t, enc.Uint64(lenByte), uint64(len(write)))
	dataByte := make([]byte, len(write))
	n, err = s.ReadAt(dataByte, lenWidth)
	assert.Nil(t, err)
	assert.Equal(t, n, len(write))
	assert.Equal(t, dataByte, write)
}

func TestStoreClose(t *testing.T) {
	f, err := os.CreateTemp("", "store_close_store")
	assert.Nil(t, err, "Error creating temp file")
	defer os.Remove(f.Name())
	fName := f.Name()
	s, err := newStore(f)
	assert.Nil(t, err, "Error creating store")
	_, _, err = s.Append(write)
	assert.Nil(t, err, "Error appending to store")
	f, beforeSize, err := openFile(fName)
	assert.Nil(t, err, "Error opening file")
	// CLosing should flush the data
	err = s.Close()
	assert.Nil(t, err, "Error closing store")
	_, afterSize, err := openFile(fName)
	assert.Nil(t, err, "Error opening file")
	assert.Greater(t, afterSize, beforeSize, "After flushing size is not greater than before closing file")

}

func openFile(name string) (file *os.File, size int64, err error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, 0, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	return f, fi.Size(), nil

}
