package sstable

import "fmt"

type smallSmallKeyKVMemtable struct {
}

func (m smallSmallKeyKVMemtable) GetSortedEntries() []KeyValue {
	return []KeyValue{
		{Key: "a", Value: []byte("value1"), Tombstone: false},
		{Key: "b", Value: []byte("value2"), Tombstone: false},
		{Key: "c", Value: []byte("value3"), Tombstone: false},
	}
}

type manySmallKeyKVMemtable struct {
}

func (m manySmallKeyKVMemtable) GetSortedEntries() []KeyValue {
	entries := make([]KeyValue, 1000)
	for i := range 1000 {
		entries[i] = KeyValue{
			Key:       fmt.Sprintf("key%03d", i),
			Value:     []byte(fmt.Sprintf("value%03d", i)),
			Tombstone: false,
		}
	}
	return entries
}

type fewLargeKeyKVMemtable struct {
}

func (m fewLargeKeyKVMemtable) GetSortedEntries() []KeyValue {
	largeValue := make([]byte, 10000)
	for i := range largeValue {
		largeValue[i] = 'A'
	}
	return []KeyValue{
		{Key: "long1", Value: largeValue, Tombstone: false},
		{Key: "long2", Value: largeValue, Tombstone: false},
		{Key: "long3", Value: largeValue, Tombstone: false},
	}
}

type manyLargeKeyKVMemtable struct {
}

func (m manyLargeKeyKVMemtable) GetSortedEntries() []KeyValue {
	largeValue := make([]byte, 10000)
	for i := range largeValue {
		largeValue[i] = 'B'
	}
	entries := make([]KeyValue, 10000)
	for i := range 10000 {
		entries[i] = KeyValue{
			Key:       fmt.Sprintf("longkey%04d", i),
			Value:     largeValue,
			Tombstone: false,
		}
	}
	return entries
}
