package wal

import "time"

// Sve potrebno za WAL
// Nisu konacne velicine
type Log struct {
	CRC       uint32
	Tombstone bool
	Timestamp []byte
	KeySize   uint32
	ValueSize uint64
	Key       string
	Value     []byte
}

// Vraca binarni zapis
func (r *Log) ToBinary() []byte {
	return nil
}

// Konstruktor
func NewLog(key string, value []byte, tombstone bool, timestamp time.Time) *Log {
	return nil
}

// CRC?
func getCRC() uint32 {
	return 1
}
