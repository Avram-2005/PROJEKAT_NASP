package sstable

import "fmt"

type manySmallKeyKVMemtable struct {
}

func (m manySmallKeyKVMemtable) GetSortedEntries() []KeyValue {
	entries := make([]KeyValue, 1000)
	for i := range 1000 {
		entries[i] = KeyValue{
			Key:       fmt.Sprintf("key%03d", i),
			Value:     []byte(fmt.Sprintf("value%d", i)),
			Tombstone: false,
		}
	}
	return entries
}
