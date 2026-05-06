package sstable

import (
	"fmt"
	"testing"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	mt "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
)

var bm *BlockManager.BlockManager
var testSSTManagers = make(map[string]*SSTableManager)

func TestMain(m *testing.M) {
	var err error
	bm, err = BlockManager.NewBlockManager(100, 4)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize BlockManager: %v", err))
	}

	m.Run()
}

func testFlush(tempDir string, mem mt.Memtable, multFiles bool) (*SSTableManager, *SSTable, error) {
	m, exists := testSSTManagers[tempDir]
	if !exists {
		var err error
		m, err = SetupSSTableManager(tempDir, SSTableConfig{
			SummaryInterval: 10,
			MultipleFiles:   multFiles,
		}, bm)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to setup SSTable: %v", err)
		}
		testSSTManagers[tempDir] = m
	}
	m.config.MultipleFiles = multFiles

	sst, err := m.Flush(mem)
	if err != nil {
		return nil, nil, fmt.Errorf("Flush failed: %v", err)
	}

	return m, sst, nil
}

func TestFlushFewSmallKVMultipleFiles(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, true)
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

}

func TestFlushFewSmallKVOneFile(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, false)
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

}

func TestFlushFewLargeKVMultipleFiles(t *testing.T) {
	mem := fewLargeKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	val, err := m.Get("long2", sst)
	if err != nil {
		t.Fatalf("Failed to get key 'long2' after flush: %v", err)
	}
	if val == nil || len(val.Value) != 10000 {
		t.Fatalf("Expected 10000-byte value for key 'long2'")
	}
}

func TestFlushFewLargeKVOneFile(t *testing.T) {
	mem := fewLargeKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	val, err := m.Get("long1", sst)
	if err != nil {
		t.Fatalf("Failed to get key 'long1' after flush: %v", err)
	}
	if val == nil || len(val.Value) != 10000 {
		t.Fatalf("Expected 10000-byte value for key 'long1'")
	}
}

func TestFlushManySmallKVMultipleFiles(t *testing.T) {
	mem := manySmallKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	for _, key := range []string{"key000", "key500", "key999"} {
		val, err := m.Get(key, sst)
		if err != nil {
			t.Fatalf("Failed to get key '%s' after flush: %v", key, err)
		}
		if val == nil {
			t.Fatalf("Key '%s' not found in SSTable", key)
		}
	}
}

func TestFlushManySmallKVOneFile(t *testing.T) {
	mem := manySmallKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	for _, key := range []string{"key000", "key500", "key999"} {
		val, err := m.Get(key, sst)
		if err != nil {
			t.Fatalf("Failed to get key '%s' after flush: %v", key, err)
		}
		if val == nil {
			t.Fatalf("Key '%s' not found in SSTable", key)
		}
	}
}

func TestFlushManyLargeKVMultipleFIles(t *testing.T) {
	mem := manyLargeKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	val, err := m.Get("longkey9999", sst)
	if err != nil {
		t.Fatalf("Failed to get key 'longkey9999' after flush: %v", err)
	}
	if val == nil || len(val.Value) != 10000 {
		t.Fatalf("Expected 10000-byte value for key 'longkey9999'")
	}
}

func TestFlushManyLargeKVOneFile(t *testing.T) {
	mem := manyLargeKeyKVMemtable{}
	m, sst, err := testFlush(t.TempDir(), mem, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	val, err := m.Get("longkey0000", sst)
	if err != nil {
		t.Fatalf("Failed to get key 'longkey0000' after flush: %v", err)
	}
	if val == nil || len(val.Value) != 10000 {
		t.Fatalf("Expected 10000-byte value for key 'longkey0000'")
	}
}

func testMergeFiles(t *testing.T, ssts []*SSTable, config SSTableConfig, expectedKeys []string, expectedValues []string, unexpectedKeys []string) {
	m, err := SetupSSTableManager(t.TempDir(), config, bm)
	if err != nil {
		t.Fatalf("Failed to setup SSTableManager: %v", err)
	}

	mergedSST, err := m.Merge(ssts, 0)
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
	tempDir := t.TempDir()
	_, sst1, err := testFlush(tempDir, smallSmallKeyKVMemtable{}, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	_, sst2, err := testFlush(tempDir, fewLargeKeyKVMemtable{}, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	_, sst3, err := testFlush(tempDir, manySmallKeyKVMemtable{}, true)
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
	tempDir := t.TempDir()
	_, sst1, err := testFlush(tempDir, smallSmallKeyKVMemtable{}, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	_, sst2, err := testFlush(tempDir, fewLargeKeyKVMemtable{}, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	_, sst3, err := testFlush(tempDir, manySmallKeyKVMemtable{}, false)
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

func testMergeHonorsTombstones(t *testing.T, multipleFiles bool) {
	tempDir := t.TempDir()
	_, olderSST, err := testFlush(tempDir, tombstoneOlderMemtable{}, multipleFiles)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	_, newerSST, err := testFlush(tempDir, tombstoneNewerMemtable{}, multipleFiles)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	expectedKeys := []string{"fresh", "keep"}
	expectedValues := []string{"fresh-val", "keep-new"}
	unexpectedKeys := []string{"dead"}

	testMergeFiles(
		t,
		[]*SSTable{olderSST, newerSST},
		SSTableConfig{SummaryInterval: 10, MultipleFiles: multipleFiles},
		expectedKeys,
		expectedValues,
		unexpectedKeys,
	)
}

func TestMergeMultipleFilesHonorsTombstones(t *testing.T) {
	testMergeHonorsTombstones(t, true)
}

func TestMergeOneFileHonorsTombstones(t *testing.T) {
	testMergeHonorsTombstones(t, false)
}
