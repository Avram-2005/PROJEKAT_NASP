package sstable

import (
	"fmt"
	"time"

	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func smallSmallKeyKVMemtableEntries() []*Record {
	ts := time.Now()
	r1, _ := NewRecord("a", []byte("value1"), false, ts)
	r2, _ := NewRecord("b", []byte("value2"), false, ts)
	r3, _ := NewRecord("c", []byte("value3"), false, ts)
	return []*Record{r1, r2, r3}
}

func manySmallKeyKVMemtableEntries() []*Record {
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

func fewLargeKeyKVMemtableEntries() []*Record {
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

func manyLargeKeyKVMemtableEntries() []*Record {
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

func tombstoneOlderMemtableEntries() []*Record {
	ts := time.Unix(100, 0)
	r1, _ := NewRecord("dead", []byte("alive-old"), false, ts)
	r2, _ := NewRecord("keep", []byte("keep-old"), false, ts)
	return []*Record{r1, r2}
}

func tombstoneNewerMemtableEntries() []*Record {
	ts := time.Unix(200, 0)
	r1, _ := NewRecord("dead", []byte{}, true, ts)
	r2, _ := NewRecord("fresh", []byte("fresh-val"), false, ts)
	r3, _ := NewRecord("keep", []byte("keep-new"), false, ts)
	return []*Record{r1, r2, r3}
}
