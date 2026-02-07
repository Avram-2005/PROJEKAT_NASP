package sstable

import (
	"fmt"
	"os"
	"testing"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

var bm *BlockManager.BlockManager

func TestMain(m *testing.M) {
	var err error
	bm, err = BlockManager.NewBlockManager(100, 4)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize BlockManager: %v", err))
	}

	m.Run()
}

func testFlush(tempDir string, mem Memtable, numFlush int) (int64, error) {
	SetupDirectory(tempDir)

	// f, _ := os.Create(filepath.Join(tempDir, fmt.Sprintf("cpu_profile_flush_%d.prof", numFlush)))
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	err := Flush(mem, numFlush, bm)
	if err != nil {
		return -1, fmt.Errorf("Flush failed: %v", err)
	}

	info, err := os.Stat(sstableFilename(numFlush, "Data"))
	if err != nil {
		return -1, fmt.Errorf("Failed to stat SSTable file: %v", err)
	}

	return info.Size(), nil
}

type emptyMemtable struct {
}

func (m emptyMemtable) GetSortedEntries() []KeyValue {
	return []KeyValue{}
}

func TestFlushEmptyMemtable(t *testing.T) {
	mem := emptyMemtable{}
	size, err := testFlush(t.TempDir(), mem, 0)
	if err == nil {
		t.Fatalf("Expected error when flushing empty memtable, but got none")
	}
	if size != -1 {
		t.Fatalf("Expected size -1 for failed flush, but got %d", size)
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
	size, err := testFlush(t.TempDir(), mem, 1)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	for i, key := range []string{"a", "b", "c"} {
		val, err := Get(key, 1, bm)
		if err != nil {
			t.Fatalf("Failed to get key '%s' after flush: %v", key, err)
		}
		expectedValue := fmt.Sprintf("value%d", i+1)
		if string(val) != expectedValue {
			t.Fatalf("Expected value '%s' for key '%s', but got %s", expectedValue, key, val)
		}
	}

	expectedSize := int64(3 * (DATA_HEADER_L + 1 + 6)) // 3 entries, each with 1 byte key and 6 byte value
	if size != expectedSize {
		t.Fatalf("Expected SSTable file size %d, but got %d", expectedSize, size)
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
	size, err := testFlush(t.TempDir(), mem, 2)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	expectedSize := int64(3 * (DATA_HEADER_L + 5 + 10000)) // 3 entries, each with 4 byte key and 10000 byte value
	if size != expectedSize {
		t.Fatalf("Expected SSTable file size %d, but got %d", expectedSize, size)
	}
}

func TestFlushManySmallKV(t *testing.T) {
	mem := manySmallKeyKVMemtable{}
	size, err := testFlush(t.TempDir(), mem, 3)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	file, err := os.Open(sstableFilename(3, "Index"))
	for i := 0; err != nil; i++ {
		var dataHeaderBuf [DATA_HEADER_L]byte
		n, err := file.Read(dataHeaderBuf[:])
		if n != 12 && err == nil {
			t.Fatalf("Expected to read 12 bytes for data header, but read %d", n)
		}
		if dataHeaderBuf[0] != 6 {
			t.Fatalf("Expected key length 6, but got %d", dataHeaderBuf[0])
		}
	}

	expectedSize := int64(1000 * (DATA_HEADER_L + 6 + 8)) // 1000 entries, each with 6 byte key and 8 byte value
	if size != expectedSize {
		t.Fatalf("Expected SSTable file size %d, but got %d", expectedSize, size)
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

func TestFlushManyLargeKV(t *testing.T) {
	mem := manyLargeKeyKVMemtable{}
	size, err := testFlush(t.TempDir(), mem, 4)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	largeValue := make([]byte, 10000)
	for i := range largeValue {
		largeValue[i] = 'B'
	}

	expectedSize := int64(10000 * (DATA_HEADER_L + 11 + 10000)) // 10000 entries, each with 11 byte key and 10000 byte value
	if size != expectedSize {
		t.Fatalf("Expected SSTable file size %d, but got %d", expectedSize, size)
	}
}
