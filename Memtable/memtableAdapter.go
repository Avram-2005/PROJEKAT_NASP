package memtable

import (
	"fmt"

	"github.com/Avram-2005/PROJEKAT_NASP/BPlusTree"
	"github.com/Avram-2005/PROJEKAT_NASP/HashMap"
	"github.com/Avram-2005/PROJEKAT_NASP/SkipList"
)

type MemtableAdapter struct {
	config MemtableConfig
	size   int
	total  int

	dataStructure interface{}
	structureType string

	getFunc    func(key string) ([]byte, bool, error)
	putFunc    func(key string, value []byte) error
	deleteFunc func(key string) (bool, error)
	clearFunc  func()
}

func NewMemtableAdapter(config MemtableConfig) (*MemtableAdapter, error) {
	adapter := &MemtableAdapter{
		config: config,
		size:   0,
		total:  0,
	}
	//inicijalizacija na osnovu tipa
	switch config.Type {
	case "hashmap":
		hm := HashMap.NewHashMap()
		adapter.dataStructure = hm
		adapter.structureType = "hashmap"
		adapter.initHashMap(hm)
	case "skip_list":
		sl, err := SkipList.NewSkipList(config.SkipListMaxHeight)
		if err != nil {
			return nil, err
		}
		adapter.dataStructure = sl
		adapter.structureType = "skip_list"
		adapter.initSkipList(sl)
	case "b_plus_tree":
		bpt, err := BPlusTree.NewBPlusTree(config.BPlusTreeDegree)
		if err != nil {
			return nil, err
		}
		adapter.dataStructure = bpt
		adapter.structureType = "b_plus_tree"
		adapter.initBPlusTree(bpt)
	default:
		return nil, fmt.Errorf("Memtable type: %s, was not recognized", config.Type)
	}
	return adapter, nil
}

// Implementacija HashMape
func (adapt *MemtableAdapter) initHashMap(hm *HashMap.HashMap) {
	adapt.getFunc = func(key string) ([]byte, bool, error) {
		value, err := hm.Get(key)
		if err != nil {
			return nil, false, nil
		}
		return value, true, nil
	}
	adapt.putFunc = func(key string, value []byte) error {
		_, err := hm.Get(key) //provera da li kljuc postoji
		exists := err == nil
		err = hm.Put(key, value)
		if err != nil {
			return err
		}
		if !exists {
			adapt.size++
		}
		adapt.total++
		return nil
	}
	adapt.deleteFunc = func(key string) (bool, error) {
		_, err := hm.Get(key)
		if err != nil {
			return false, nil
		}
		err = hm.Delete(key)
		if err != nil {
			return false, err
		}
		adapt.size--
		return true, nil
	}
	adapt.clearFunc = func() {
		hm = HashMap.NewHashMap()
		adapt.size = 0
		adapt.total = 0
	}

}

// implementacija skipliste
func (adapt *MemtableAdapter) initSkipList(sl *SkipList.SkipList) {
	adapt.getFunc = func(key string) ([]byte, bool, error) {
		return sl.Get(key)
	}
	adapt.putFunc = func(key string, value []byte) error {
		_, found, _ := sl.Get(key)
		err := sl.Put(key, value)
		if err != nil {
			return err
		}
		if !found {
			adapt.size++
		}
		adapt.total++
		return nil
	}
	adapt.deleteFunc = func(key string) (bool, error) {
		_, found, _ := sl.Get(key)
		if !found {
			return false, nil
		}
		err := sl.Delete(key)
		if err != nil {
			return false, err
		}
		adapt.size--
		return true, nil
	}
	adapt.clearFunc = func() {
		sl.Clear()
		adapt.size = 0
		adapt.total = 0
	}
}

// implementacija b+ stabla
func (adapt *MemtableAdapter) initBPlusTree(bpt *BPlusTree.BPlusTree) {
	adapt.getFunc = func(key string) ([]byte, bool, error) {
		value, found := bpt.Search(key)
		return value, found, nil
	}
	adapt.putFunc = func(key string, value []byte) error {
		_, found := bpt.Search(key)
		err := bpt.Insert(key, value)
		if err != nil {
			return err
		}
		if !found {
			adapt.size++
		}
		adapt.total++
		return nil
	}
	adapt.deleteFunc = func(key string) (bool, error) {
		deleted := bpt.Delete(key)
		if deleted {
			adapt.size--
		}
		return deleted, nil
	}
	adapt.clearFunc = func() {
		bpt, _ = BPlusTree.NewBPlusTree(adapt.config.BPlusTreeDegree)
		adapt.size = 0
		adapt.total = 0

	}
}

// Implementacija rangeScan
func (adapt *MemtableAdapter) RangeScan(startKey, endKey string) []KeyValue {
	switch adapt.structureType {
	case "hashmap":
		hm := adapt.dataStructure.(*HashMap.HashMap)
		hmEntries := hm.RangeScan(startKey, endKey)
		return convertToKeyValue(hmEntries)
	case "skip_list":
		sl := adapt.dataStructure.(*SkipList.SkipList)
		slEntries := sl.RangeScan(startKey, endKey)
		return convertToKeyValue(slEntries)
	case "b_plus_tree":
		bpt := adapt.dataStructure.(*BPlusTree.BPlusTree)
		bptEntries := bpt.RangeScan(startKey, endKey)
		return convertToKeyValue(bptEntries)
	default:
		return []KeyValue{}
	}
}

// Implementacija PrefixScan
func (adapt *MemtableAdapter) PrefixScan(prefix string) []KeyValue {
	switch adapt.structureType {
	case "hashmap":
		hm := adapt.dataStructure.(*HashMap.HashMap)
		hmEntries := hm.PrefixScan(prefix)
		return convertToKeyValue(hmEntries)
	case "skip_list":
		sl := adapt.dataStructure.(*SkipList.SkipList)
		slEntries := sl.PrefixScan(prefix)
		return convertToKeyValue(slEntries)
	case "b_plus_tree":
		bpt := adapt.dataStructure.(*BPlusTree.BPlusTree)
		bptEntries := bpt.PrefixScan(prefix)
		return convertToKeyValue(bptEntries)
	default:
		return []KeyValue{}
	}
}

// Implementacija dobavljanja sortiranih podataka
func (adapt *MemtableAdapter) GetSortedEntries() []KeyValue {
	switch adapt.structureType {
	case "hashmap":
		hm := adapt.dataStructure.(*HashMap.HashMap)
		hmEntries := hm.GetSortedEntries()
		return convertToKeyValue(hmEntries)
	case "skip_list": //vec je sortirana
		sl := adapt.dataStructure.(*SkipList.SkipList)
		slEntries := sl.RangeScan("", "zzzzzzzzz")
		return convertToKeyValue(slEntries)
	case "b_plus_tree": //vec je sortirano
		bpt := adapt.dataStructure.(*BPlusTree.BPlusTree)
		bptEntries := bpt.RangeScan("", "zzzzzzzzz")
		return convertToKeyValue(bptEntries)
	default:
		return []KeyValue{}
	}
}

func convertToKeyValue(entries []struct {
	Key   string
	Value []byte
}) []KeyValue {
	result := make([]KeyValue, len(entries))
	for i, e := range entries {
		result[i] = KeyValue{
			Key:       e.Key,
			Value:     e.Value,
			Tombstone: e.Value == nil,
		}
	}
	return result
}

func (adapt *MemtableAdapter) Put(key string, value []byte) error {
	return adapt.putFunc(key, value)
}
func (adapt *MemtableAdapter) Get(key string) ([]byte, bool, error) {
	return adapt.getFunc(key)
}
func (adapt *MemtableAdapter) Delete(key string) (bool, error) {
	return adapt.deleteFunc(key)
}
func (adapt *MemtableAdapter) Size() int {
	return adapt.size
}
func (adapt *MemtableAdapter) TotalEntries() int {
	return adapt.total
}
func (adapt *MemtableAdapter) IsEmpty() bool {
	return adapt.size == 0
}
func (adapt *MemtableAdapter) Clear() {
	adapt.clearFunc()
}
func (adapt *MemtableAdapter) Iterator() Iterator {
	entries := adapt.GetSortedEntries()
	return NewBaseIterator(entries)
}
func (adapt *MemtableAdapter) ShouldFlush() bool {
	if adapt.config.MaxSizeBytes > 0 {
		estimatedSize := adapt.total * 100
		return estimatedSize >= adapt.config.MaxSizeBytes
	}
	if adapt.config.MaxSizeEntries > 0 {
		return adapt.total >= adapt.config.MaxSizeEntries
	}
	return false
}

func (adapt *MemtableAdapter) IsFull() bool {
	return adapt.ShouldFlush()
}
