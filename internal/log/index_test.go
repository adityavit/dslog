package log

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestIndex(t *testing.T) {
	file, err := os.CreateTemp("", "index_test")
	assert.Nil(t, err, "error creating test file")
	c := Config{}
	c.Segment.MaxIndexBytes = 1024
	i, err := newIndex(file, c)
	assert.Nil(t, err, "error creating new index")
	assert.Equal(t, file.Name(), i.Name(), "index name is not same as file")

	// test reading from empty index
	_, _, err = i.Read(-1)
	assert.Error(t, err, "error not received when reading empty index")

	// test writing and reading the entries from the index
	entries := []struct {
		off uint32
		pos uint64
	}{
		{
			0, 0,
		},
		{
			1, 10,
		},
	}

	for _, test := range entries {
		errWrite := i.Write(test.off, test.pos)
		assert.Nil(t, errWrite, "error received while writing test")
		offset, pos, errRead := i.Read(int64(test.off))
		assert.Nil(t, errRead, "read from the index file throws an error")
		assert.Equal(t, test.off, offset, "offset read from the index doesn't match the expected")
		assert.Equal(t, test.pos, pos, "position read from the index doesn't match the expected")
	}

	// test reading beyond the index

	_, _, err = i.Read(int64(len(entries)))
	assert.Equal(t, io.EOF, err, "error is not end of file, when read outside the index")

	// Test closing of the index
	err = i.Close()
	assert.Nil(t, err, "error when closing index")

	// Create index again with the same file
	file, _ = os.OpenFile(file.Name(), os.O_RDWR, 0600)
	i, err = newIndex(file, c)
	off, pos, err := i.Read(-1)
	assert.Nil(t, err, "error received from reading from index")
	assert.Equal(t, entries[1].off, off, "last offset entry doesn't match")
	assert.Equal(t, entries[1].pos, pos, "last position entry doesn't match")
}
