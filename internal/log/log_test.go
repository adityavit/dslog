package log

import (
	api "github.com/adityavit/dslog/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"io"
	"os"
	"testing"
)

func TestLog(t *testing.T) {
	testCases := map[string]func(t *testing.T, log *Log){
		"append and read a record success": testAppendRead,
		"offset out of range error":        testOutOfRangeErr,
		"init with existing segments":      testInitExisting,
		"reader":                           testReader,
		"truncate":                         testTruncate,
	}
	c := Config{}
	rec := &api.Record{
		Value: []byte("hello world"),
	}
	recBytes, err := proto.Marshal(rec)
	assert.NoError(t, err, "error create when marshaling rec")
	c.Segment.MaxStoreBytes = uint64(len(recBytes))
	for name, fn := range testCases {
		t.Run(name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "log_test")
			assert.NoError(t, err, "error creating dir")
			defer os.RemoveAll(dir)
			log, err := newLog(dir, c)
			assert.NoError(t, err, "error create new log")
			fn(t, log)
		})
	}
}

func testAppendRead(t *testing.T, log *Log) {
	rec := api.Record{
		Value: []byte("hello world"),
	}
	off, err := log.Append(&rec)
	assert.NoError(t, err, "Error when appending record")
	assert.Equal(t, uint64(0), off)

	actualRec, err := log.Read(off)
	assert.NoError(t, err, "Error when reading record")
	assert.Equal(t, rec.Value, actualRec.Value, "read record doesn't match stored record")
}

func testOutOfRangeErr(t *testing.T, log *Log) {
	read, err := log.Read(uint64(1))
	assert.Error(t, err, "error when reading the record")
	assert.Nil(t, read, "read has data")
}

func testInitExisting(t *testing.T, log *Log) {
	rec := api.Record{
		Value: []byte("hello world"),
	}
	for i := 0; i < 3; i++ {
		off, err := log.Append(&rec)
		assert.NoError(t, err, "Error when appending record")
		assert.Equal(t, uint64(i), off)
	}
	lOffset, err := log.LowestOffset()
	assert.Equal(t, uint64(0), lOffset, "lower offset doesn't match")
	hOffset, err := log.HighestOffset()
	assert.Equal(t, uint64(2), hOffset, "higher offset doesn't match")
	err = log.Close()
	assert.NoError(t, err, "Error when closing log")
	newLog, err := newLog(log.Dir, log.Config)
	assert.NoError(t, err, "Error when starting new log in same directory")
	readRec, err := newLog.Read(uint64(0))
	assert.NoError(t, err, "Error when reading record")
	assert.Equal(t, rec.Value, readRec.Value, "read record doesn't match the stored record")
	lOffset, err = newLog.LowestOffset()
	assert.Equal(t, uint64(0), lOffset, "lower offset doesn't match")
	hOffset, err = newLog.HighestOffset()
	assert.Equal(t, uint64(2), hOffset, "higher offset doesn't match")
}

func testReader(t *testing.T, log *Log) {
	rec := api.Record{
		Value: []byte("hello world"),
	}
	off, err := log.Append(&rec)
	assert.NoError(t, err, "Error when appending record")
	assert.Equal(t, uint64(0), off)
	reader := log.Reader()
	data, err := io.ReadAll(reader)
	assert.NoError(t, err, "Error when reading raw bytes from store")
	record := &api.Record{}
	err = proto.Unmarshal(data[lenWidth:], record)
	assert.NoError(t, err, "Error when unmarshalling record")
	assert.Equal(t, rec.Value, record.Value, "read record doesn't match stored record")
}

func testTruncate(t *testing.T, log *Log) {
	rec := api.Record{
		Value: []byte("hello world"),
	}
	for i := 0; i < 3; i++ {
		off, err := log.Append(&rec)
		assert.NoError(t, err, "Error when appending record")
		assert.Equal(t, uint64(i), off)
	}
	err := log.Truncate(1)
	assert.NoError(t, err, "Error when truncating log")
	_, err = log.Read(0)
	assert.Error(t, err, "No error reading 0 record after truncating as 1")
}
