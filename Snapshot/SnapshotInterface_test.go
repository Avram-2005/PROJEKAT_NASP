package snapshot

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"

	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
)

func TestSnapshotMemtable(t *testing.T) {
	conf := memtable.MemtableConfig{
		Type:              "hashmap",
		MaxSizeEntries:    10,
		SkipListMaxHeight: 8,
		BPlusTreeDegree:   2,
	}
	memtableInstance, err := memtable.NewMemtableAdapter(conf)
	if err != nil {
		fmt.Print("Error initializing memtable instance")
		t.FailNow()
	}
	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 120)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 80)
	binary.BigEndian.PutUint16(data3, 67)
	memtableInstance.Put("key1", data1)
	memtableInstance.Put("key2", data2)
	memtableInstance.Put("key3", data3)
	secondMemtableInstance, err := memtable.NewMemtableAdapter(conf)
	if err != nil {
		fmt.Print("Error initializing second memtable instance")
		t.FailNow()
	}
	data4 := make([]byte, 90)
	secondMemtableInstance.Put("key3", data4)
	snapshot1, err := NewSnapshotMemtable("key1", memtableInstance)
	if err != nil {
		fmt.Print(err)
		fmt.Print("Error creating snapshot1")
		t.FailNow()
	}
	compare1, err := snapshot1.GetValue()
	if !reflect.DeepEqual(data1, (*compare1)) {
		fmt.Print(err)
		fmt.Print("Snapshot1 different between read and write")
		t.FailNow()
	}
	snapshot2, err := NewSnapshotMemtable("key2", memtableInstance)
	if err != nil {
		fmt.Print(err)
		fmt.Print("Error creating snapshot2")
		t.FailNow()
	}
	compare2, err := snapshot2.GetValue()
	if !reflect.DeepEqual(data2, (*compare2)) {
		fmt.Print(err)
		fmt.Print("Snapshot2 different between read and write")
		t.FailNow()
	}
	snapshot3, err := NewSnapshotMemtable("key3", memtableInstance)
	if err != nil {
		fmt.Print(err)
		fmt.Print("Error creating snapshot3")
		t.FailNow()
	}
	compare3, err := snapshot3.GetValue()
	if !reflect.DeepEqual(data3, (*compare3)) {
		fmt.Print(err)
		fmt.Print("Snapshot3 different between read and write")
		t.FailNow()
	}
	snapshot4, err := NewSnapshotMemtable("key3", secondMemtableInstance)
	if err != nil {
		fmt.Print(err)
		fmt.Print("Error creating snapshot3")
		t.FailNow()
	}
	compare4, err := snapshot4.GetValue()
	if !reflect.DeepEqual(data4, (*compare4)) {
		fmt.Print(err)
		fmt.Print("Snapshot4 different between read and write")
		t.FailNow()
	}
}
