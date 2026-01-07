package Memtable

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
	Put(key string, value []byte) error
	Get(key string) ([]byte, bool, error) //povratna vred-vrednost, uspesnost dodavanja, error
	Delete(key string) (bool, error)      //povratna vred- uspesnost brisanja, error

	Size() int         //broj postojecih elem (bez tombstone elem)
	TotalEntries() int //ukupan broj unosa
	IsEmpty() bool

	GetSortedEntries() []KeyValue //povratna vred/ parovi kljuc-vred neophodni za sstable
	RangeScan(startKey, endKey string) []KeyValue
	PrefixScan(prefix string) []KeyValue

	Iterator() Iterator
	CheckIfShouldFlush() bool //provera da li je vreme za flush u SSTable
	IsFull() bool
}

type MemtableStruct struct {
	Type              string //neka od tri strukture: hashmapa, skiplista ili b+ stablo
	MaxSizeBytes      int    //max velicina u bajtovima
	MaxSizeEntries    int    //max broj elemenata koji moze da primi
	BPlusTreeDegree   int    //max stepen stabla
	SkipListMaxHeight int    //max visina skipliste

}
