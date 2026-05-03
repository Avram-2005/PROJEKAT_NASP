package sstable

import (
	"fmt"
	"time"

	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

type smallSmallKeyKVMemtable struct {
}

func (m smallSmallKeyKVMemtable) GetSortedEntries() []Record {
	ts := time.Now()
	r1, _ := NewRecord("a", []byte("value1"), false, ts)
	r2, _ := NewRecord("b", []byte("value2"), false, ts)
	r3, _ := NewRecord("c", []byte("value3"), false, ts)
	return []Record{*r1, *r2, *r3}
}

type manySmallKeyKVMemtable struct {
}

func (m manySmallKeyKVMemtable) GetSortedEntries() []Record {
	ts := time.Now()
	entries := make([]Record, 1000)
	for i := range 1000 {
		rec, _ := NewRecord(
			fmt.Sprintf("key%03d", i),
			[]byte(fmt.Sprintf("value%03d", i)),
			false,
			ts,
		)
		entries[i] = *rec
	}
	return entries
}

type fewLargeKeyKVMemtable struct {
}

func (m fewLargeKeyKVMemtable) GetSortedEntries() []Record {
	ts := time.Now()
	largeValue := make([]byte, 10000)
	for i := range largeValue {
		largeValue[i] = 'A'
	}
	r1, _ := NewRecord("long1", largeValue, false, ts)
	r2, _ := NewRecord("long2", largeValue, false, ts)
	r3, _ := NewRecord("long3", largeValue, false, ts)
	return []Record{*r1, *r2, *r3}
}

type manyLargeKeyKVMemtable struct {
}

func (m manyLargeKeyKVMemtable) GetSortedEntries() []Record {
	ts := time.Now()
	largeValue := make([]byte, 10000)
	for i := range largeValue {
		largeValue[i] = 'B'
	}
	entries := make([]Record, 10000)
	for i := range 10000 {
		rec, _ := NewRecord(
			fmt.Sprintf("longkey%04d", i),
			largeValue,
			false,
			ts,
		)
		entries[i] = *rec
	}
	return entries
}

type tombstoneOlderMemtable struct {
}

func (m tombstoneOlderMemtable) GetSortedEntries() []Record {
	ts := time.Unix(100, 0)
	r1, _ := NewRecord("dead", []byte("alive-old"), false, ts)
	r2, _ := NewRecord("keep", []byte("keep-old"), false, ts)
	return []Record{*r1, *r2}
}

type tombstoneNewerMemtable struct {
}

func (m tombstoneNewerMemtable) GetSortedEntries() []Record {
	ts := time.Unix(200, 0)
	r1, _ := NewRecord("dead", []byte{}, true, ts)
	r2, _ := NewRecord("fresh", []byte("fresh-val"), false, ts)
	r3, _ := NewRecord("keep", []byte("keep-new"), false, ts)
	return []Record{*r1, *r2, *r3}
}
