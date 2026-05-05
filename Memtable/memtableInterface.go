package memtable

import (
	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

type KeyValue struct {
	Key       string
	Value     []byte
	Tombstone bool //za brisanje, true ako je obrisan
}

// iterator za iteriranje kroz Memtable
type Iterator interface {
	Next() bool
	Key() string
	Value() []byte
	Tombstone() bool
	Reset()
	Error() error
}
type Memtable interface {
	PutRecord(rec *record.Record) error
	Get(key string) ([]byte, bool, error) //povratna vred-vrednost, uspesnost dodavanja, error
	Put(key string, value []byte) error
	Delete(key string) (bool, error)
	Size() int         //broj postojecih elem (bez tombstone elem)
	TotalEntries() int //ukupan broj unosa
	IsEmpty() bool
	Clear()

	GetSortedEntries() []*record.Record //povratna vrednost je record, neophodna za sstable
	RangeScan(startKey, endKey string) []*record.Record
	PrefixScan(prefix string) []*record.Record

	Iterator() Iterator
	ShouldFlush() bool //provera da li je vreme za flush u SSTable
	IsFull() bool
}

type MemtableConfig struct {
	Type              string //neka od tri strukture: hashmapa, skiplista ili b+ stablo
	MaxSizeBytes      int    //max velicina u bajtovima
	MaxSizeEntries    int    //max broj elemenata koji moze da primi
	BPlusTreeDegree   int    //max stepen stabla
	SkipListMaxHeight int    //max visina skipliste

}
