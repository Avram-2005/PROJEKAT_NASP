package sstable

import "errors"

// FIXME: DELETE AFTER Memtable MERGE /
// ////////////////////////////////////

type KeyValue struct {
	Key       string
	Value     []byte
	Tombstone bool //za brisanje, true ako je obrisan
}

type Memtable interface {
	GetSortedEntries() []KeyValue //povratna vred/ parovi kljuc-vred neophodni za sstable
}

//////////////////////////////////////

// TODO: Compression (1.3[DZ3])
type dataBlock struct {
	crc       uint32
	blockSize uint16
	timestamp uint32
	data      []byte
}

// TODO: Save to multiple files (Cassandra) or in one file (LevelDB) (1.3[DZ2])
type SSTable struct {
	index   index
	summary summary
}

func (sst *SSTable) Get(key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func Flush(mem *Memtable) (*SSTable, error) {
	return nil, errors.New("not implemented")
}
