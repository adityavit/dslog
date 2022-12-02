package server

import (
	"errors"
	"sync"
)

var ErrOffsetNotFound = errors.New("offset not found")

type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

type Log struct {
	m       sync.Mutex
	records []Record
}

func NewLog() *Log {
	return &Log{}
}

func (l *Log) Append(record Record) (uint64, error) {
	l.m.Lock()
	defer l.m.Unlock()
	record.Offset = uint64(len(l.records))
	l.records = append(l.records, record)
	return record.Offset, nil
}

func (l *Log) Read(offset uint64) (Record, error) {
	l.m.Lock()
	defer l.m.Unlock()
	if offset >= uint64(len(l.records)) {
		return Record{}, ErrOffsetNotFound
	}
	return l.records[offset], nil
}
