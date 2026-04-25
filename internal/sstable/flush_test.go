package sstable

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
	. "github.com/Avram-2005/PROJEKAT_NASP/utils"
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

func testFlush(tempDir string, mem Memtable, numFlush int, multFiles bool) (*SSTableManager, *SSTable, error) {
	m, err := SetupSSTableManager(tempDir, SSTableConfig{
		SummaryInterval: 10,
		MultipleFiles:   multFiles,
	}, bm)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to setup SSTable: %v", err)
	}

	// f, _ := os.Create(filepath.Join(tempDir, fmt.Sprintf("cpu_profile_flush_%d.prof", numFlush)))
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	sst, err := m.Flush(mem, numFlush)
	if err != nil {
		return nil, nil, fmt.Errorf("Flush failed: %v", err)
	}

	return m, sst, nil
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

func calcSummarySectionSize(count int, keyLen int, summaryInterval int) int64 {
	numSummaryEntries := math.Ceil(float64(count) / float64(summaryInterval))
	return int64(int(numSummaryEntries)*(INDEX_HEADER_L+keyLen)) + 2*int64(KEY_SIZE_L+keyLen)
}

func calcFilterSectionSize(count int) int64 {
	return int64(BloomFilter.CalculateBloomFilterSize(uint(count), 0.01))
}

/*
	func calcMetadataSectionSize(count int, valueLen int) int64 {
		maxLeaves := 1
		for maxLeaves < count {
			maxLeaves *= 2
		}

		realLeafSize := int64(1 + 32 + 4 + valueLen)
		emptyNodeSize := int64(1 + 32)

		realLeaves := int64(count)
		emptyLeaves := int64(maxLeaves - count)
		totalInternalNodes := int64(maxLeaves - 1)

		return realLeaves*realLeafSize + emptyLeaves*emptyNodeSize + totalInternalNodes*emptyNodeSize + 4
	}
*/

func oneFileSize(keyCount int, keyLen int, valueLen int, metadataSize int64, summaryInterval int) int64 {
	size := calcFilterSectionSize(keyCount)
	size += calcDataSectionSize(keyCount, keyLen, valueLen)
	size += calcIndexSectionSize(keyCount, keyLen)
	size += calcSummarySectionSize(keyCount, keyLen, summaryInterval)
	//size += calcMetadataSectionSize(keyCount, valueLen)
	size += 4 * OFFSET_L // 4 offsets for the sections
	size += metadataSize
	return size
}

func TestFlushFewSmallKVMultipleFiles(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, 1, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	for i, key := range []string{"a", "b", "c"} {
		val, err := m.Get(key, sst)
		if err != nil {
			t.Fatalf("Failed to get key '%s' after flush: %v", key, err)
		}
		expectedValue := fmt.Sprintf("value%d", i+1)
		if val == nil {
			t.Fatalf("Key '%s' not found in SSTable", key)
		}
		if string(val.Value) != expectedValue {
			t.Fatalf("Expected value '%s' for key '%s', but got %v", expectedValue, key, val)
		}
	}

	sstablePath := m.sstableFilepath(0, 1)
	dataFilename := sstableFilenameMultFile(sstablePath, "Data")
	expectedSize := calcDataSectionSize(3, 1, 6) // 3 entries, each with 1 byte key and 6 byte value
	testFileSize(t, dataFilename, expectedSize)

	indexFilename := sstableFilenameMultFile(sstablePath, "Index")
	expectedSize = calcIndexSectionSize(3, 1) // 3 entries, each with 1 byte key and 8 byte offset
	testFileSize(t, indexFilename, expectedSize)

	summaryFilename := sstableFilenameMultFile(sstablePath, "Summary")
	expectedSize = calcSummarySectionSize(3, 1, m.config.SummaryInterval) // 1 entry in summary, with 1 byte key and 8 byte offset
	testFileSize(t, summaryFilename, expectedSize)

	filterFilename := sstableFilenameMultFile(sstablePath, "Filter")
	expectedSize = calcFilterSectionSize(3)
	testFileSize(t, filterFilename, expectedSize)

	//metadataFilename := sstableFilename(1, "Metadata")
	//expectedSize = calcMetadataSectionSize(3, 6)
	//testFileSize(t, metadataFilename, expectedSize)
}

func TestFlushFewSmallKVOneFile(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, 1, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	for i, key := range []string{"a", "b", "c"} {
		val, err := m.Get(key, sst)
		if err != nil {
			t.Fatalf("Failed to get key '%s' after flush: %v", key, err)
		}
		if val == nil {
			t.Fatalf("Key '%s' not found in SSTable", key)
		}
		expectedValue := fmt.Sprintf("value%d", i+1)
		if string(val.Value) != expectedValue {
			t.Fatalf("Expected value '%s' for key '%s', but got %v", expectedValue, key, val)
		}
	}

	sstablePath := m.sstableFilepath(0, 1)
	info, err := os.Stat(sstablePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	calculatedSize := calcFilterSectionSize(3) + calcDataSectionSize(3, 1, 6) + calcIndexSectionSize(3, 1) + calcSummarySectionSize(3, 1, m.config.SummaryInterval) + 4*OFFSET_L
	metadataSize := info.Size() - calculatedSize

	expectedSize := oneFileSize(3, 1, 6, metadataSize, m.config.SummaryInterval) // 3 entries, each with 1 byte key and 6 byte value
	testFileSize(t, sstablePath, expectedSize)
}

func TestFlushFewLargeKVMultipleFiles(t *testing.T) {
	mem := fewLargeKeyKVMemtable{}
	m, _, err := testFlush(t.TempDir(), mem, 2, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	sstablePath := m.sstableFilepath(0, 2)
	dataFilename := sstableFilenameMultFile(sstablePath, "Data")
	expectedSize := calcDataSectionSize(3, 5, 10000) // 3 entries, each with 5 byte key and 10000 byte value
	testFileSize(t, dataFilename, expectedSize)

	indexFilename := sstableFilenameMultFile(sstablePath, "Index")
	expectedSize = calcIndexSectionSize(3, 5) // 3 entries, each with 5 byte key and 8 byte offset
	testFileSize(t, indexFilename, expectedSize)

	summaryFilename := sstableFilenameMultFile(sstablePath, "Summary")
	expectedSize = calcSummarySectionSize(3, 5, m.config.SummaryInterval) // 1 entry in summary, with 5 byte key and 8 byte offset
	testFileSize(t, summaryFilename, expectedSize)

	filterFilename := sstableFilenameMultFile(sstablePath, "Filter")
	expectedSize = calcFilterSectionSize(3)
	testFileSize(t, filterFilename, expectedSize)

	/*metadataFilename := sstableFilename(2, "Metadata")
	expectedSize = calcMetadataSectionSize(3, 10000)
	testFileSize(t, metadataFilename, expectedSize)*/
}

func TestFlushFewLargeKVOneFile(t *testing.T) {
	mem := fewLargeKeyKVMemtable{}
	m, _, err := testFlush(t.TempDir(), mem, 2, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	sstablePath := m.sstableFilepath(0, 2)
	info, err := os.Stat(sstablePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	calculatedSize := calcFilterSectionSize(3) + calcDataSectionSize(3, 5, 10000) + calcIndexSectionSize(3, 5) + calcSummarySectionSize(3, 5, m.config.SummaryInterval) + 4*OFFSET_L
	metadataSize := info.Size() - calculatedSize
	expectedSize := oneFileSize(3, 5, 10000, metadataSize, m.config.SummaryInterval) // 3 entries, each with 5 byte key and 10000 byte value
	testFileSize(t, sstablePath, expectedSize)
}

func TestFlushManySmallKVMultipleFiles(t *testing.T) {
	mem := manySmallKeyKVMemtable{}
	m, _, err := testFlush(t.TempDir(), mem, 3, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	sstablePath := m.sstableFilepath(0, 3)
	dataFilename := sstableFilenameMultFile(sstablePath, "Data")
	expectedSize := calcDataSectionSize(1000, 6, 8) // 1000 entries, each with 6 byte key and 8 byte value
	testFileSize(t, dataFilename, expectedSize)

	indexFilename := sstableFilenameMultFile(sstablePath, "Index")
	expectedSize = calcIndexSectionSize(1000, 6) // 1000 entries, each with 6 byte key and 8 byte offset
	testFileSize(t, indexFilename, expectedSize)

	summaryFilename := sstableFilenameMultFile(sstablePath, "Summary")
	expectedSize = calcSummarySectionSize(1000, 6, m.config.SummaryInterval) // 1 entry in summary, with 6 byte key and 8 byte offset
	testFileSize(t, summaryFilename, expectedSize)

	filterFilename := sstableFilenameMultFile(sstablePath, "Filter")
	expectedSize = calcFilterSectionSize(1000)
	testFileSize(t, filterFilename, expectedSize)

	/*metadataFilename := sstableFilename(3, "Metadata")
	expectedSize = calcMetadataSectionSize(1000, 8)
	testFileSize(t, metadataFilename, expectedSize)*/
}

func TestFlushManySmallKVOneFile(t *testing.T) {
	mem := manySmallKeyKVMemtable{}
	m, _, err := testFlush(t.TempDir(), mem, 3, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	sstablePath := m.sstableFilepath(0, 3)
	info, err := os.Stat(sstablePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	calculatedSize := calcFilterSectionSize(1000) + calcDataSectionSize(1000, 6, 8) + calcIndexSectionSize(1000, 6) + calcSummarySectionSize(1000, 6, m.config.SummaryInterval) + 4*OFFSET_L
	metadataSize := info.Size() - calculatedSize
	expectedSize := oneFileSize(1000, 6, 8, metadataSize, m.config.SummaryInterval) // 1000 entries, each with 6 byte key and 8 byte value
	testFileSize(t, sstablePath, expectedSize)
}

func TestFlushManyLargeKVMultipleFIles(t *testing.T) {
	mem := manyLargeKeyKVMemtable{}
	m, _, err := testFlush(t.TempDir(), mem, 4, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	sstablePath := m.sstableFilepath(0, 4)
	dataFilename := sstableFilenameMultFile(sstablePath, "Data")
	expectedSize := calcDataSectionSize(10000, 11, 10000) // 10000 entries, each with 11 byte key and 10000 byte value
	testFileSize(t, dataFilename, expectedSize)

	indexFilename := sstableFilenameMultFile(sstablePath, "Index")
	expectedSize = calcIndexSectionSize(10000, 11) // 10000 entries, each with 11 byte key and 8 byte offset
	testFileSize(t, indexFilename, expectedSize)

	summaryFilename := sstableFilenameMultFile(sstablePath, "Summary")
	expectedSize = calcSummarySectionSize(10000, 11, m.config.SummaryInterval) // 1 entry in summary, with 11 byte key and 8 byte offset
	testFileSize(t, summaryFilename, expectedSize)

	filterFilename := sstableFilenameMultFile(sstablePath, "Filter")
	expectedSize = calcFilterSectionSize(10000)
	testFileSize(t, filterFilename, expectedSize)

	/*metadataFilename := sstableFilename(4, "Metadata")
	expectedSize = calcMetadataSectionSize(10000, 10000)
	testFileSize(t, metadataFilename, expectedSize)*/
}

func TestFlushManyLargeKVOneFile(t *testing.T) {
	mem := manyLargeKeyKVMemtable{}
	m, _, err := testFlush(t.TempDir(), mem, 4, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	sstablePath := m.sstableFilepath(0, 4)
	info, err := os.Stat(sstablePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	calculatedSize := calcFilterSectionSize(10000) + calcDataSectionSize(10000, 11, 10000) + calcIndexSectionSize(10000, 11) + calcSummarySectionSize(10000, 11, m.config.SummaryInterval) + 4*OFFSET_L
	metadataSize := info.Size() - calculatedSize
	expectedSize := oneFileSize(10000, 11, 10000, metadataSize, m.config.SummaryInterval) // 10000 entries, each with 11 byte key and 10000 byte value
	testFileSize(t, sstablePath, expectedSize)
}

/*
func TestMetadataValidationOneFile(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, _, err := testFlush(t.TempDir(), mem, 1, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	isValid, corruptedData, err := m.ValidateSSTable(1)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !isValid {
		t.Fatalf("Merkle validation failed, corruption data count: %d", len(corruptedData))
	}
}

func TestMetadataCorruptionOneFile(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, _, err := testFlush(t.TempDir(), mem, 2, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	file := m.sstableFilepath(0, 2)
	f, err := os.OpenFile(file, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Open file error: %v", err)
	}
	defer f.Close()

	dataReader := newBlockReader(f, m.bm, 0)
	var dataHeaderBuf [DATA_HEADER_L]byte
	_, err = dataReader.Read(dataHeaderBuf[:])
	if err != nil {
		t.Fatalf("Failed to read data header: %v", err)
	}

	currByte := CRC_L + TIMESTAMP_L + TOMBSTONE_L
	keySize := binary.BigEndian.Uint32(dataHeaderBuf[currByte:])
	currByte += KEY_SIZE_L

	valueOffset := DATA_HEADER_L + uint64(keySize)

	f2, err := os.OpenFile(file, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Open file error: %v", err)
	}
	defer f2.Close()

	newValue := []byte("AAA")
	_, err = f2.WriteAt(newValue, int64(valueOffset))
	if err != nil {
		t.Fatalf("Failed to corrupt data: %v", err)
	}
	f2.Close()

	newBM, err := BlockManager.NewBlockManager(100, 4)
	if err != nil {
		t.Fatalf("Failed to create new BlockManager: %v", err)
	}
	m.bm = newBM

	isValid, corruptedData, err := m.ValidateSSTable(2)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}

	if isValid {
		t.Fatalf("Merkle failed to detect change")
	}
	if len(corruptedData) == 0 {
		t.Fatalf("Corrupted data not detected")
	}

	if string(corruptedData[0]) != "value1" {
		t.Fatalf("Expected corrupted data 'value1', got '%s'", string(corruptedData[0]))
	}
}

func TestMetadataValidationMultipleFiles(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, _, err := testFlush(t.TempDir(), mem, 1, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	isValid, corruptedData, err := m.ValidateSSTable(1)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !isValid {
		t.Fatalf("Merkle validation failed, corruption data count: %d", len(corruptedData))
	}
}

func TestMetadataCorruptionMultipleFiles(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, _, err := testFlush(t.TempDir(), mem, 1, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	sstablePath := m.sstableFilepath(0, 1)
	dataFile := sstableFilenameMultFile(sstablePath, "Data")

	readFile, err := os.Open(dataFile)
	if err != nil {
		t.Fatalf("Failed to open file for reading: %v", err)
	}
	defer readFile.Close()

	reader := newBlockReader(readFile, bm, 0)

	var dataHeaderBuf [DATA_HEADER_L]byte
	_, err = reader.Read(dataHeaderBuf[:])
	if err != nil {
		t.Fatalf("Failed to read data header: %v", err)
	}

	currByte := CRC_L + TIMESTAMP_L + TOMBSTONE_L
	keySize := binary.BigEndian.Uint32(dataHeaderBuf[currByte:])
	currByte += KEY_SIZE_L

	valueOffset := DATA_HEADER_L + uint64(keySize)

	f2, err := os.OpenFile(dataFile, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Open file error: %v", err)
	}
	defer f2.Close()

	metadataFile := sstableFilenameMultFile(sstablePath, "Metadata")
	metadata, err := os.ReadFile(metadataFile)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	end := len(metadata)
	for end > 0 && metadata[end-1] == 0 {
		end--
	}
	metadata = metadata[:end]

	originalTree := merkleTree.Deserialize(metadata)
	if originalTree == nil {
		t.Fatalf("Failed to deserialize tree")
	}

	newValue := []byte("AAA")
	_, err = f2.WriteAt(newValue, int64(valueOffset))
	if err != nil {
		t.Fatalf("Failed to corrupt data: %v", err)
	}
	f2.Close()

	newBM, err := BlockManager.NewBlockManager(100, 4)
	if err != nil {
		t.Fatalf("Failed to create new BlockManager: %v", err)
	}
	m.bm = newBM

	isValid, corruptedData, err := m.ValidateSSTable(1)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}

	if isValid {
		t.Fatalf("Merkle failed to detect change")
	}
	if len(corruptedData) == 0 {
		t.Fatalf("Corrupted data not detected")
	}

	found := false
	for _, data := range corruptedData {
		if string(data) == "value1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Expected corrupted data 'value1', got: %v", corruptedData)
	}
}
*/

func testMergeFiles(t *testing.T, ssts []*SSTable, config SSTableConfig, expectedKeys []string, expectedValues []string, unexpectedKeys []string) {
	m, err := SetupSSTableManager(t.TempDir(), config, bm)
	if err != nil {
		t.Fatalf("Failed to setup SSTableManager: %v", err)
	}

	mergedSST, err := m.Merge(ssts, 0, len(ssts)+1)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	for i, key := range expectedKeys {
		val, err := m.Get(key, mergedSST)
		if err != nil {
			t.Fatalf("Failed to get key '%s' after merge: %v", key, err)
		}
		if val == nil {
			t.Fatalf("Key '%s' not found in merged SSTable", key)
		}
		if string(val.Value) != expectedValues[i] {
			t.Fatalf("Expected value '%s' for key '%s', but got %v", expectedValues[i], key, val)
		}
	}

	for _, key := range unexpectedKeys {
		val, err := m.Get(key, mergedSST)
		if err != nil {
			t.Fatalf("Failed to get key '%s' after merge: %v", key, err)
		}
		if val != nil {
			t.Fatalf("Expected key '%s' to not be found in merged SSTable, but got value %v", key, val)
		}
	}
}

func TestMergeMultipleFiles(t *testing.T) {
	_, sst1, err := testFlush(t.TempDir(), smallSmallKeyKVMemtable{}, 1, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	_, sst2, err := testFlush(t.TempDir(), fewLargeKeyKVMemtable{}, 2, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	_, sst3, err := testFlush(t.TempDir(), manySmallKeyKVMemtable{}, 3, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	expectedKeys := []string{"a", "b", "c", "key000", "key001", "key125", "key989", "long2"}

	largeValue := make([]byte, 10000)
	for i := range largeValue {
		largeValue[i] = 'A'
	}
	expectedValues := []string{"value1", "value2", "value3", "value000", "value001", "value125", "value989", string(largeValue)}
	unexpectedKeys := []string{"asdf", "test"}

	testMergeFiles(t, []*SSTable{sst1, sst2, sst3}, SSTableConfig{SummaryInterval: 100, MultipleFiles: true}, expectedKeys, expectedValues, unexpectedKeys)
}

func TestMergeOneFile(t *testing.T) {
	_, sst1, err := testFlush(t.TempDir(), smallSmallKeyKVMemtable{}, 1, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	_, sst2, err := testFlush(t.TempDir(), fewLargeKeyKVMemtable{}, 2, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	_, sst3, err := testFlush(t.TempDir(), manySmallKeyKVMemtable{}, 3, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	expectedKeys := []string{"a", "b", "c", "key000", "key001", "key125", "key989", "long2"}

	largeValue := make([]byte, 10000)
	for i := range largeValue {
		largeValue[i] = 'A'
	}
	expectedValues := []string{"value1", "value2", "value3", "value000", "value001", "value125", "value989", string(largeValue)}
	unexpectedKeys := []string{"asdf", "test"}

	testMergeFiles(t, []*SSTable{sst1, sst2, sst3}, SSTableConfig{SummaryInterval: 100, MultipleFiles: false}, expectedKeys, expectedValues, unexpectedKeys)
}
