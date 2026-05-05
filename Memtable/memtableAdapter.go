package memtable

import (
	"fmt"
	"sort"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BPlusTree"
	"github.com/Avram-2005/PROJEKAT_NASP/HashMap"
	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
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
		data, err := hm.Get(key)
		if err != nil {
			return nil, false, nil
		}
		rec, _, err := record.DeserializeRecord(data)
		if err != nil || rec.Tombstone {
			return nil, false, nil
		}
		return rec.Value, true, nil
	}
	adapt.putFunc = func(key string, value []byte) error {
		return hm.Put(key, value)
	}
	adapt.deleteFunc = func(key string) (bool, error) {
		_, found, _ := adapt.getFunc(key)
		if !found {
			return false, nil
		}
		rec := &record.Record{
			Key:       key,
			Value:     nil,
			Tombstone: true,
			Timestamp: time.Now(),
		}
		data := rec.Serialize()
		err := hm.Put(key, data)
		if err != nil {
			return false, err
		}
		adapt.size--
		return true, nil
	}
	adapt.clearFunc = func() {
		newHm := HashMap.NewHashMap()
		adapt.dataStructure = newHm
		adapt.size = 0
		adapt.total = 0
	}

}

// implementacija skipliste
func (adapt *MemtableAdapter) initSkipList(sl *SkipList.SkipList) {
	adapt.getFunc = func(key string) ([]byte, bool, error) {
		data, found, err := sl.Get(key)
		if err != nil || !found {
			return nil, false, err
		}
		rec, _, err := record.DeserializeRecord(data)
		if err != nil || rec.Tombstone {
			return nil, false, nil
		}
		return rec.Value, true, nil
	}
	adapt.putFunc = func(key string, value []byte) error {
		return sl.Put(key, value)
	}
	adapt.deleteFunc = func(key string) (bool, error) {
		_, found, _ := adapt.getFunc(key)
		if !found {
			return false, nil
		}
		rec := &record.Record{
			Key:       key,
			Value:     nil,
			Tombstone: true,
			Timestamp: time.Now(),
		}
		data := rec.Serialize()
		err := sl.Put(key, data)
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
		data, found := bpt.Search(key)
		if !found {
			return nil, false, nil
		}
		rec, _, err := record.DeserializeRecord(data)
		if err != nil || rec.Tombstone {
			return nil, false, nil
		}
		return rec.Value, true, nil
	}

	adapt.putFunc = func(key string, value []byte) error {
		return bpt.Insert(key, value)
	}

	adapt.deleteFunc = func(key string) (bool, error) {
		_, found := bpt.Search(key)
		if !found {
			return false, nil
		}
		rec := &record.Record{
			Key:       key,
			Value:     nil,
			Tombstone: true,
			Timestamp: time.Now(),
		}
		data := rec.Serialize()
		err := bpt.Insert(key, data)
		if err != nil {
			return false, err
		}
		adapt.size--
		return true, nil
	}

	adapt.clearFunc = func() {
		newBpt, _ := BPlusTree.NewBPlusTree(adapt.config.BPlusTreeDegree)
		adapt.dataStructure = newBpt
		adapt.size = 0
		adapt.total = 0

	}
}

// Implementacija rangeScan
func (adapt *MemtableAdapter) RangeScan(startKey, endKey string) []*record.Record {
	switch adapt.structureType {
	case "hashmap":
		hm := adapt.dataStructure.(*HashMap.HashMap)
		return adapt.keyValueToRecWithoutTombstone(hm.RangeScan(startKey, endKey))
	case "skip_list":
		sl := adapt.dataStructure.(*SkipList.SkipList)
		return adapt.keyValueToRecWithoutTombstone(sl.RangeScan(startKey, endKey))
	case "b_plus_tree":
		bpt := adapt.dataStructure.(*BPlusTree.BPlusTree)
		return adapt.keyValueToRecWithoutTombstone(bpt.RangeScan(startKey, endKey))
	default:
		return []*record.Record{}
	}
}

// Implementacija PrefixScan
func (adapt *MemtableAdapter) PrefixScan(prefix string) []*record.Record {
	switch adapt.structureType {
	case "hashmap":
		hm := adapt.dataStructure.(*HashMap.HashMap)
		return adapt.keyValueToRecWithoutTombstone(hm.PrefixScan(prefix))
	case "skip_list":
		sl := adapt.dataStructure.(*SkipList.SkipList)
		return adapt.keyValueToRecWithoutTombstone(sl.PrefixScan(prefix))
	case "b_plus_tree":
		bpt := adapt.dataStructure.(*BPlusTree.BPlusTree)
		return adapt.keyValueToRecWithoutTombstone(bpt.PrefixScan(prefix))
	default:
		return []*record.Record{}
	}
}

// Implementacija dobavljanja sortiranih podataka
func (adapt *MemtableAdapter) GetSortedEntries() []*record.Record {
	switch adapt.structureType {
	case "hashmap":
		hm := adapt.dataStructure.(*HashMap.HashMap)
		return adapt.keyValueToRecords(hm.GetSortedEntries())
	case "skip_list": //vec je sortirana
		sl := adapt.dataStructure.(*SkipList.SkipList)
		return adapt.keyValueToRecords(sl.RangeScan("", "\U0010FFFF"))
	case "b_plus_tree": //vec je sortirano
		bpt := adapt.dataStructure.(*BPlusTree.BPlusTree)
		return adapt.keyValueToRecords(bpt.RangeScan("", "\U0010FFFF"))
	default:
		return []*record.Record{}
	}
}

func (adapt *MemtableAdapter) Put(key string, value []byte) error {
	_, exists, _ := adapt.Get(key)
	rec := &record.Record{
		Key:       key,
		Value:     value,
		Tombstone: false,
		Timestamp: time.Now(),
	}
	data := rec.Serialize()
	err := adapt.putFunc(key, data)
	if err != nil {
		return err
	}
	if !exists {
		adapt.size++
	}
	adapt.total++
	return nil
}

func (adapt *MemtableAdapter) PutRecord(rec *record.Record) error {
	if rec == nil {
		return fmt.Errorf("record ne moze biti nil")
	}
	_, exists, _ := adapt.Get(rec.Key)
	data := rec.Serialize()
	err := adapt.putFunc(rec.Key, data)
	if err != nil {
		return err
	}
	if !exists && !rec.Tombstone {
		adapt.size++
	}
	adapt.total++
	return nil
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

func (adapt *MemtableAdapter) keyValueToRecords(keyValEntries []struct {
	Key   string
	Value []byte
}) []*record.Record {
	result := make([]*record.Record, 0, len(keyValEntries))
	for _, e := range keyValEntries {
		rec, _, err := record.DeserializeRecord(e.Value)
		if err != nil {
			continue
		}
		result = append(result, rec)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

func (adapt *MemtableAdapter) keyValueToRecWithoutTombstone(keyValEntries []struct {
	Key   string
	Value []byte
}) []*record.Record {
	result := make([]*record.Record, 0, len(keyValEntries))
	for _, e := range keyValEntries {
		rec, _, err := record.DeserializeRecord(e.Value)
		if err != nil || rec.Tombstone {
			continue
		}
		result = append(result, rec)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

func (adapt *MemtableAdapter) Iterator() Iterator {
	entries := adapt.GetSortedEntries()
	active := make([]*record.Record, 0, len(entries))
	for _, e := range entries {
		if !e.Tombstone {
			active = append(active, e)
		}
	}
	return NewBaseIterator(active)
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
