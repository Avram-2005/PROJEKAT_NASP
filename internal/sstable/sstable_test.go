package sstable

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

func testFlush(t *testing.T, mem Memtable, numFlush int) error {
	bm, err := BlockManager.NewBlockManager(10e6, 4)
	if err != nil {
		t.Fatalf("Failed to create BlockManager: %v", err)
	}

	tempDir := t.TempDir()
	// tempDir := filepath.Join(".", "testdata", "sstable_flush_tests")
	SetupDirectory(tempDir)

	// f, _ := os.Create(filepath.Join(tempDir, fmt.Sprintf("cpu_profile_flush_%d.prof", numFlush)))
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	err = Flush(mem, numFlush, bm)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	filename := filepath.Join(tempDir, "tables", fmt.Sprintf("usertable-%d-Data.txt", numFlush))
	_, err = bm.Get(filename, 1)
	if err != nil {
		return err
	}
	return nil
}

type emptyMemtable struct {
}

func (m emptyMemtable) GetSortedEntries() []KeyValue {
	return []KeyValue{}
}

func TestFlushEmptyMemtable(t *testing.T) {
	mem := emptyMemtable{}
	err := testFlush(t, mem, 0)
	if err == nil {
		t.Fatalf("Expected error when flushing empty memtable, but got none")
	}
}

type smallSmallKeyKVMemtable struct {
}

func (m smallSmallKeyKVMemtable) GetSortedEntries() []KeyValue {
	return []KeyValue{
		{Key: "a", Value: []byte("value1"), Tombstone: false},
		{Key: "b", Value: []byte("value2"), Tombstone: false},
		{Key: "c", Value: []byte("value3"), Tombstone: false},
	}
}

func TestFlushFewSmallKV(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	err := testFlush(t, mem, 1)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
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

func TestFlushFewLargeKV(t *testing.T) {
	mem := fewLargeKeyKVMemtable{}
	err := testFlush(t, mem, 2)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
}

type manySmallKeyKVMemtable struct {
}

func (m manySmallKeyKVMemtable) GetSortedEntries() []KeyValue {
	entries := make([]KeyValue, 1000)
	for i := range 1000 {
		entries[i] = KeyValue{
			Key:       fmt.Sprintf("key%d", i),
			Value:     []byte(fmt.Sprintf("value%d", i)),
			Tombstone: false,
		}
	}
	return entries
}

func TestFlushManySmallKV(t *testing.T) {
	mem := manySmallKeyKVMemtable{}
	err := testFlush(t, mem, 3)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
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
			Key:       fmt.Sprintf("longkey%d", i),
			Value:     largeValue,
			Tombstone: false,
		}
	}
	return entries
}

func TestFlushManyLargeKV(t *testing.T) {
	mem := manyLargeKeyKVMemtable{}
	err := testFlush(t, mem, 4)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
}
