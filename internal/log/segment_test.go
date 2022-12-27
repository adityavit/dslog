package log

import (
	api "github.com/adityavit/dslog/api/v1"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestSegment(t *testing.T) {
	dir, _ := os.MkdirTemp("", "segment_test")
	defer os.RemoveAll(dir)
	rec := &api.Record{Value: []byte("hello world")}
	c := Config{}
	c.Segment.MaxIndexBytes = entWidth * 3
	// store three records only
	c.Segment.MaxStoreBytes = 1024
	s, err := newSegment(dir, 16, c)
	assert.NoError(t, err, "error creating segment")
	assert.Equal(t, uint64(16), s.baseOffset, "base offset is not set correctly")
	assert.Equal(t, uint64(16), s.nextOffset, "next offset is not set correctly")
	assert.False(t, s.IsMaxed(), "No space in the segment")

	for i := 0; i < 3; i++ {
		off, err := s.Append(rec)
		assert.NoError(t, err, "error appending the record to the segment")
		assert.Equal(t, s.baseOffset+uint64(i), off, "record is not added in the correct offset")
		r, err := s.Read(off)
		assert.NoError(t, err, "Error reading the record")
		assert.Equal(t, rec.Value, r.Value, "Fetched record is different from the stored one")
	}
	_, err = s.Append(rec)
	assert.Equal(t, io.EOF, err, "No EOF error is returned")
	assert.True(t, s.IsMaxed(), "segment should be true for maxed store records")

	// Creating new configuration for with smaller Store Bytes
	c.Segment.MaxStoreBytes = uint64(len(rec.Value) * 3)
	c.Segment.MaxIndexBytes = 1024

	s, err = newSegment(dir, 16, c)
	assert.NoError(t, err, "received error when creating segment again with different configuration")
	assert.True(t, s.IsMaxed(), "There should not be space in the segment")
	err = s.Remove()
	assert.NoError(t, err, "error when removing segment")
	s, err = newSegment(dir, 16, c)
	assert.NoError(t, err, "Error when creating segment again after deletion")
	assert.False(t, s.IsMaxed(), "segment should be empty after creating again")
}
