package wal

import (
	"encoding/binary"
	"fmt"
	"time"
)

// Sve potrebno za WAL
// Nisu konacne velicine
// CRC 4B | Timestamp 8B | Tombstone 1B | KeySize 4B | ValueSize 8B | Key ... | Value ...
// Ukupno 29B
type Log struct {
	CRC       uint32
	Timestamp []byte
	Tombstone bool
	KeySize   uint64
	ValueSize uint64
	Key       string
	Value     []byte
}

const HEADERSIZE = 4 + 8 + 1 + 8 + 8

// Vraca binarni zapis
func (r *Log) ToBinary() ([]byte, error) {
	totalSize := HEADERSIZE + r.KeySize + r.ValueSize
	//Prostor za skladistenje celog loga
	buf := make([]byte, totalSize)
	offset := 0

	//CRC 4B
	binary.BigEndian.PutUint32(buf[offset:offset+4], r.CRC)
	offset += 4

	//Timestamp 8B
	copy(buf[offset:offset+8], r.Timestamp)
	offset += 8

	//Tombstone 1B
	if r.Tombstone {
		buf[offset] = 1
	} else {
		buf[offset] = 0
	}
	offset += 1

	//KeySize 4B
	binary.BigEndian.PutUint64(buf[offset:offset+8], r.KeySize)
	offset += 8

	//ValueSize 8B
	binary.BigEndian.PutUint64(buf[offset:offset+8], r.ValueSize)
	offset += 8

	//Key
	copy(buf[offset:offset+int(r.KeySize)], r.Key)
	offset += int(r.KeySize)

	//Value
	copy(buf[offset:offset+int(r.ValueSize)], r.Value)

	return buf, nil
}

// Konstruktor
func NewLog(key string, value []byte, tombstone bool, timestamp time.Time) (*Log, error) {
	//Pretvaranje time -> []byte
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(timestamp.UnixNano()))

	if tombstone && len(value) != 0 {
		return nil, fmt.Errorf("Ne moze tombstone da bude false i da value ima vrednost!")
	}

	l := &Log{
		CRC:       getCRC(),
		Timestamp: ts,
		Tombstone: tombstone,
		Key:       key,
		Value:     value,
		KeySize:   uint64(len(key)),
		ValueSize: uint64(len(value)),
	}

	return l, nil
}

// CRC?
func getCRC() uint32 {
	return 1
}
