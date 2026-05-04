package sstable

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	merkleTree "github.com/Avram-2005/PROJEKAT_NASP/MerkleTree"
	. "github.com/Avram-2005/PROJEKAT_NASP/utils"
)

func TestMetadataValidationOneFile(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	isValid, corruptedData, err := m.ValidateSSTable(sst)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !isValid {
		t.Fatalf("Merkle validation failed, corruption data count: %d", len(corruptedData))
	}
}

func TestMetadataCorruptionOneFile(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	file, err := os.Open(sst.path)
	if err != nil {
		t.Fatalf("failed to open SSTable file: %v", err)
	}
	defer file.Close()

	dataReader := newBlockReader(file, m.bm, 0)
	var dataHeaderBuf [DATA_HEADER_L]byte
	_, err = dataReader.Read(dataHeaderBuf[:])
	if err != nil {
		t.Fatalf("Failed to read data header: %v", err)
	}

	currByte := CRC_L + TIMESTAMP_L + TOMBSTONE_L
	keySize := binary.BigEndian.Uint32(dataHeaderBuf[currByte:])
	currByte += KEY_SIZE_L

	valueOffset := DATA_HEADER_L + uint64(keySize)

	f2, err := os.OpenFile(sst.path, os.O_RDWR, 0644)
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

	isValid, corruptedKeys, err := m.ValidateSSTable(sst)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}

	if isValid {
		t.Fatalf("Merkle failed to detect change")
	}
	if len(corruptedKeys) == 0 {
		t.Fatalf("Corrupted data not detected")
	}

	if corruptedKeys[0].Key != "a" {
		t.Fatalf("Expected corrupted key 'a', got '%s'", string(corruptedKeys[0].Key))
	}

	if corruptedKeys[0].Key == "a" {
		t.Logf("Corrupted data under key: %s, with data: %s, with timestamp: %s, with tombstone: %t", string(corruptedKeys[0].Key), string(corruptedKeys[0].Value), corruptedKeys[0].Timestamp.String(), corruptedKeys[0].Tombstone)
	}
}

func TestMetadataValidationMultipleFiles(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	isValid, corruptedData, err := m.ValidateSSTable(sst)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !isValid {
		t.Fatalf("Merkle validation failed, corruption data count: %d", len(corruptedData))
	}
}

func TestMetadataCorruptionMultipleFiles(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	sstablePath := m.sstableFilepath(0, 0)
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

	isValid, corruptedKeys, err := m.ValidateSSTable(sst)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}

	if isValid {
		t.Fatalf("Merkle failed to detect change")
	}
	if len(corruptedKeys) == 0 {
		t.Fatalf("Corrupted data not detected")
	}

	if corruptedKeys[0].Key != "a" {
		t.Fatalf("Expected corrupted key 'a', got '%s'", string(corruptedKeys[0].Key))
	}

	if corruptedKeys[0].Key == "a" {
		t.Logf("Corrupted data under key: %s, with data: %s, with timestamp: %s, with tombstone: %t", string(corruptedKeys[0].Key), string(corruptedKeys[0].Value), corruptedKeys[0].Timestamp.String(), corruptedKeys[0].Tombstone)
	}
}
