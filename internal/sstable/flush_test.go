package sstable

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
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

func testFlush(tempDir string, mem Memtable, numFlush int, multFiles bool) error {
	SetupSSTable(tempDir, 100, multFiles)

	// f, _ := os.Create(filepath.Join(tempDir, fmt.Sprintf("cpu_profile_flush_%d.prof", numFlush)))
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	err := Flush(mem, numFlush, bm)
	if err != nil {
		return fmt.Errorf("Flush failed: %v", err)
	}

	return nil
}

func testFileSize(t *testing.T, filename string, expectedSize int64) {
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", filename, err)
	}
	if info.Size() != expectedSize {
		t.Fatalf("Expected file size %d for file %s, but got %d", expectedSize, filename, info.Size())
	}
}

func calcDataSectionSize(count int, keyLen int, valueLen int) int64 {
	return int64(count * (DATA_HEADER_L + keyLen + valueLen))
}

func calcIndexSectionSize(count int, keyLen int) int64 {
	return int64(count * (INDEX_HEADER_L + keyLen))
}

func calcSummarySectionSize(count int, keyLen int) int64 {
	numSummaryEntries := math.Ceil(float64(count) / float64(summaryInterval))
	return int64(int(numSummaryEntries)*(INDEX_HEADER_L+keyLen)) + 2*int64(KEY_SIZE_L+keyLen)
}

func calcFilterSectionSize(count int) int64 {
	return int64(BloomFilter.CalculateBloomFilterSize(uint(count), 0.01))
}

func oneFileSize(keyCount int, keyLen int, valueLen int) int64 {
	size := calcFilterSectionSize(keyCount)
	size += calcDataSectionSize(keyCount, keyLen, valueLen)
	size += calcIndexSectionSize(keyCount, keyLen)
	size += calcSummarySectionSize(keyCount, keyLen)
	size += 3 * OFFSET_L // 3 offsets for the sections
	return size
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

func TestFlushFewSmallKVMultipleFiles(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	err := testFlush(t.TempDir(), mem, 1, true)
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

	dataFilename := sstableFilename(1, "Data")
	expectedSize := calcDataSectionSize(3, 1, 6) // 3 entries, each with 1 byte key and 6 byte value
	testFileSize(t, dataFilename, expectedSize)

	indexFilename := sstableFilename(1, "Index")
	expectedSize = calcIndexSectionSize(3, 1) // 3 entries, each with 1 byte key and 8 byte offset
	testFileSize(t, indexFilename, expectedSize)

	summaryFilename := sstableFilename(1, "Summary")
	expectedSize = calcSummarySectionSize(3, 1) // 1 entry in summary, with 1 byte key and 8 byte offset
	testFileSize(t, summaryFilename, expectedSize)

	filterFilename := sstableFilename(1, "Filter")
	expectedSize = calcFilterSectionSize(3)
	testFileSize(t, filterFilename, expectedSize)
}

func TestFlushFewSmallKVOneFile(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	err := testFlush(t.TempDir(), mem, 1, false)
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

	filename := sstableFilenameOneFile(1)
	expectedSize := oneFileSize(3, 1, 6) // 3 entries, each with 1 byte key and 6 byte value
	testFileSize(t, filename, expectedSize)
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

func TestFlushFewLargeKVMultipleFiles(t *testing.T) {
	mem := fewLargeKeyKVMemtable{}
	err := testFlush(t.TempDir(), mem, 2, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	dataFilename := sstableFilename(2, "Data")
	expectedSize := calcDataSectionSize(3, 5, 10000) // 3 entries, each with 5 byte key and 10000 byte value
	testFileSize(t, dataFilename, expectedSize)

	indexFilename := sstableFilename(2, "Index")
	expectedSize = calcIndexSectionSize(3, 5) // 3 entries, each with 5 byte key and 8 byte offset
	testFileSize(t, indexFilename, expectedSize)

	summaryFilename := sstableFilename(2, "Summary")
	expectedSize = calcSummarySectionSize(3, 5) // 1 entry in summary, with 5 byte key and 8 byte offset
	testFileSize(t, summaryFilename, expectedSize)

	filterFilename := sstableFilename(2, "Filter")
	expectedSize = calcFilterSectionSize(3)
	testFileSize(t, filterFilename, expectedSize)
}

func TestFlushFewLargeKVOneFile(t *testing.T) {
	mem := fewLargeKeyKVMemtable{}
	err := testFlush(t.TempDir(), mem, 2, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	filename := sstableFilenameOneFile(2)
	expectedSize := oneFileSize(3, 5, 10000) // 3 entries, each with 5 byte key and 10000 byte value
	testFileSize(t, filename, expectedSize)
}

func TestFlushManySmallKVMultipleFiles(t *testing.T) {
	mem := manySmallKeyKVMemtable{}
	err := testFlush(t.TempDir(), mem, 3, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	dataFilename := sstableFilename(3, "Data")
	expectedSize := calcDataSectionSize(1000, 6, 8) // 1000 entries, each with 6 byte key and 8 byte value
	testFileSize(t, dataFilename, expectedSize)

	indexFilename := sstableFilename(3, "Index")
	expectedSize = calcIndexSectionSize(1000, 6) // 1000 entries, each with 6 byte key and 8 byte offset
	testFileSize(t, indexFilename, expectedSize)

	summaryFilename := sstableFilename(3, "Summary")
	expectedSize = calcSummarySectionSize(1000, 6) // 1 entry in summary, with 6 byte key and 8 byte offset
	testFileSize(t, summaryFilename, expectedSize)

	filterFilename := sstableFilename(3, "Filter")
	expectedSize = calcFilterSectionSize(1000)
	testFileSize(t, filterFilename, expectedSize)
}

func TestFlushManySmallKVOneFile(t *testing.T) {
	mem := manySmallKeyKVMemtable{}
	err := testFlush(t.TempDir(), mem, 3, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	filename := sstableFilenameOneFile(3)
	expectedSize := oneFileSize(1000, 6, 8) // 1000 entries, each with 6 byte key and 8 byte value
	testFileSize(t, filename, expectedSize)
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

func TestFlushManyLargeKVMultipleFIles(t *testing.T) {
	mem := manyLargeKeyKVMemtable{}
	err := testFlush(t.TempDir(), mem, 4, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	dataFilename := sstableFilename(4, "Data")
	expectedSize := calcDataSectionSize(10000, 11, 10000) // 10000 entries, each with 11 byte key and 10000 byte value
	testFileSize(t, dataFilename, expectedSize)

	indexFilename := sstableFilename(4, "Index")
	expectedSize = calcIndexSectionSize(10000, 11) // 10000 entries, each with 11 byte key and 8 byte offset
	testFileSize(t, indexFilename, expectedSize)

	summaryFilename := sstableFilename(4, "Summary")
	expectedSize = calcSummarySectionSize(10000, 11) // 1 entry in summary, with 11 byte key and 8 byte offset
	testFileSize(t, summaryFilename, expectedSize)

	filterFilename := sstableFilename(4, "Filter")
	expectedSize = calcFilterSectionSize(10000)
	testFileSize(t, filterFilename, expectedSize)
}

func TestFlushManyLargeKVOneFile(t *testing.T) {
	mem := manyLargeKeyKVMemtable{}
	err := testFlush(t.TempDir(), mem, 4, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	filename := sstableFilenameOneFile(4)
	expectedSize := oneFileSize(10000, 11, 10000) // 10000 entries, each with 11 byte key and 10000 byte value
	testFileSize(t, filename, expectedSize)
}
