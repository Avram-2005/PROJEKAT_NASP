package sstable

import (
	"fmt"
	"time"

	mt "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

type mockMemtableBase struct{}

func (m mockMemtableBase) PutRecord(rec *Record) error {
	return fmt.Errorf("mock memtable: PutRecord not supported")
}

func (m mockMemtableBase) Get(key string) ([]byte, bool, error) {
	return nil, false, nil
}

func (m mockMemtableBase) GetRecord(key string) (*Record, bool, error) {
	return nil, false, nil
}

func (m mockMemtableBase) Put(key string, value []byte) error {
	return fmt.Errorf("mock memtable: Put not supported")
}

func (m mockMemtableBase) Delete(key string) (bool, error) {
	return false, nil
}

func (m mockMemtableBase) Size() int {
	return len(m.GetSortedEntries())
}

func (m mockMemtableBase) TotalEntries() int {
	return len(m.GetSortedEntries())
}

func (m mockMemtableBase) IsEmpty() bool {
	return len(m.GetSortedEntries()) == 0
}

func (m mockMemtableBase) Clear() {}

func (m mockMemtableBase) GetSortedEntries() []*Record {
	return nil
}

func (m mockMemtableBase) RangeScan(startKey, endKey string) []*Record {
	return nil
}

func (m mockMemtableBase) PrefixScan(prefix string) []*Record {
	return nil
}

func (m mockMemtableBase) Iterator() mt.Iterator {
	return nil
}

func (m mockMemtableBase) ShouldFlush() bool {
	return false
}

func (m mockMemtableBase) IsFull() bool {
	return false
}

type smallSmallKeyKVMemtable struct {
	mockMemtableBase
}

func (m smallSmallKeyKVMemtable) GetSortedEntries() []*Record {
	ts := time.Now()
	r1, _ := NewRecord("a", []byte("value1"), false, ts)
	r2, _ := NewRecord("b", []byte("value2"), false, ts)
	r3, _ := NewRecord("c", []byte("value3"), false, ts)
	return []*Record{r1, r2, r3}
}

type manySmallKeyKVMemtable struct {
	mockMemtableBase
}

func (m manySmallKeyKVMemtable) GetSortedEntries() []*Record {
	ts := time.Now()
	entries := make([]*Record, 1000)
	for i := range 1000 {
		rec, _ := NewRecord(
			fmt.Sprintf("key%03d", i),
			[]byte(fmt.Sprintf("value%03d", i)),
			false,
			ts,
		)
		entries[i] = rec
	}
	return entries
}

type fewLargeKeyKVMemtable struct {
	mockMemtableBase
}

func (m fewLargeKeyKVMemtable) GetSortedEntries() []*Record {
	ts := time.Now()
	largeValue := make([]byte, 10000)
	for i := range largeValue {
		largeValue[i] = 'A'
	}
	r1, _ := NewRecord("long1", largeValue, false, ts)
	r2, _ := NewRecord("long2", largeValue, false, ts)
	r3, _ := NewRecord("long3", largeValue, false, ts)
	return []*Record{r1, r2, r3}
}

type manyLargeKeyKVMemtable struct {
	mockMemtableBase
}

func (m manyLargeKeyKVMemtable) GetSortedEntries() []*Record {
	ts := time.Now()
	largeValue := make([]byte, 10000)
	for i := range largeValue {
		largeValue[i] = 'B'
	}
	entries := make([]*Record, 10000)
	for i := range 10000 {
		rec, _ := NewRecord(
			fmt.Sprintf("longkey%04d", i),
			largeValue,
			false,
			ts,
		)
		entries[i] = rec
	}
	return entries
}

type tombstoneOlderMemtable struct {
	mockMemtableBase
}

func (m tombstoneOlderMemtable) GetSortedEntries() []*Record {
	ts := time.Unix(100, 0)
	r1, _ := NewRecord("dead", []byte("alive-old"), false, ts)
	r2, _ := NewRecord("keep", []byte("keep-old"), false, ts)
	return []*Record{r1, r2}
}

type tombstoneNewerMemtable struct {
	mockMemtableBase
}

func (m tombstoneNewerMemtable) GetSortedEntries() []*Record {
	ts := time.Unix(200, 0)
	r1, _ := NewRecord("dead", []byte{}, true, ts)
	r2, _ := NewRecord("fresh", []byte("fresh-val"), false, ts)
	r3, _ := NewRecord("keep", []byte("keep-new"), false, ts)
	return []*Record{r1, r2, r3}
}
