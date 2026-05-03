package sstable

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
)

var testTempDirs = make(map[testHelper]string)
var testManagers = make(map[testHelper]*SSTableManager)

type testHelper interface {
	Helper()
	Fatalf(format string, args ...any)
	TempDir() string
}

func flush(t testHelper, multFiles bool, mem Memtable) (*SSTableManager, *SSTable) {
	// Get or create temp directory for this test
	tempDir, exists := testTempDirs[t]
	if !exists {
		tempDir = t.TempDir()
		testTempDirs[t] = tempDir
	}

	m, exists := testManagers[t]
	if !exists {
		var err error
		m, err = SetupSSTableManager(tempDir, SSTableConfig{
			SummaryInterval: 10,
			MultipleFiles:   multFiles,
		}, bm)
		if err != nil {
			t.Fatalf("Failed to setup SSTable: %v", err)
		}
		testManagers[t] = m
	}
	m.config.MultipleFiles = multFiles

	sst, err := m.Flush(mem)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	return m, sst
}

func flush1(t testHelper, multFiles bool) (*SSTableManager, *SSTable) {
	return flush(t, multFiles, &manySmallKeyKVMemtable{})
}

func flush2(t testHelper, multFiles bool) (*SSTableManager, *SSTable) {
	return flush(t, multFiles, &fewLargeKeyKVMemtable{})
}

func flush3(t testHelper, multFiles bool) (*SSTableManager, *SSTable) {
	return flush(t, multFiles, &manyLargeKeyKVMemtable{})
}

func testGetSpecific(t *testing.T, key string, expectedValue string, multFiles bool) {
	m, sst := flush1(t, multFiles)

	val, err := m.Get(key, sst)
	if err != nil {
		t.Fatalf("Failed to get key '%s': %v", key, err)
	}
	if string(val.Value) != expectedValue {
		t.Fatalf("Expected value '%s' for key '%s', but got %v", expectedValue, key, val)
	}
}

func TestGetSpecificFirstKeyMultipleFiles(t *testing.T) {
	testGetSpecific(t, "key000", "value000", true)
}

func TestGetSpecificFirstKeyOneFile(t *testing.T) {
	testGetSpecific(t, "key000", "value000", false)
}

func TestGetSpecificSecondKeyMultipleFiles(t *testing.T) {
	testGetSpecific(t, "key001", "value001", true)
}

func TestGetSpecificSecondKeyOneFile(t *testing.T) {
	testGetSpecific(t, "key001", "value001", false)
}

func TestGetSpecificLastKeyMultipleFiles(t *testing.T) {
	testGetSpecific(t, "key999", "value999", true)
}

func TestGetSpecificLastKeyOneFile(t *testing.T) {
	testGetSpecific(t, "key999", "value999", false)
}

func TestGetSpecificMiddleKeyMultipleFiles(t *testing.T) {
	testGetSpecific(t, "key500", "value500", true)
}

func TestGetSpecificMiddleKeyOneFile(t *testing.T) {
	testGetSpecific(t, "key500", "value500", false)
}

func TestGetSpecificNonExistentKeyMultipleFiles(t *testing.T) {
	m, sst := flush1(t, true)

	value, err := m.Get("nonexistent", sst)
	if err != nil {
		t.Fatalf("Error when getting non-existent key: %v", err)
	}
	if value != nil {
		t.Fatalf("Expected nil value when getting non-existent key, but got %v", value)
	}

	runtime.GC()
}

func TestGetSpecificNonExistentKeyOneFile(t *testing.T) {
	m, sst := flush1(t, false)

	value, err := m.Get("nonexistent", sst)
	if err != nil {
		t.Fatalf("Error when getting non-existent key: %v", err)
	}
	if value != nil {
		t.Fatalf("Expected nil value when getting non-existent key, but got %v", value)
	}
}

func testSSTableIterator(t *testing.T, multFiles bool) {
	m, sst := flush1(t, multFiles)

	iter, err := m.NewSSTableIterator(sst)
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}
	defer iter.Close()

	expectedKeyNum := 0
	for {
		rec := iter.Rec
		expectedKey := fmt.Sprintf("key%03d", expectedKeyNum)
		expectedValue := fmt.Sprintf("value%03d", expectedKeyNum)
		if rec.Key != expectedKey || string(rec.Value) != expectedValue {
			t.Fatalf("Expected key '%s' with value '%s', but got key '%s' with value '%s'", expectedKey, expectedValue, rec.Key, rec.Value)
		}
		expectedKeyNum++

		iterHasNext, err := iter.Next()
		if err != nil {
			t.Fatalf("Error when iterating: %v", err)
		}
		if !iterHasNext {
			break
		}
	}
	if expectedKeyNum != 1000 {
		t.Fatalf("Expected to iterate over 1000 keys, but iterated over %d", expectedKeyNum)
	}
}

func TestSSTableIteratorOneFile(t *testing.T) {
	testSSTableIterator(t, false)
}

func TestSSTableIteratorMultipleFiles(t *testing.T) {
	testSSTableIterator(t, true)
}

func TestGet(t *testing.T) {
	flush1(t, true)
	flush2(t, false)
	flush3(t, false)

	lsmCfg := LSMConfig{
		NumLevels:        4,
		CompactionFactor: 2,
	}
	sstCfg := SSTableConfig{
		SummaryInterval: 10,
		MultipleFiles:   true,
	}

	lsm, err := NewLSM(lsmCfg, testTempDirs[t], sstCfg, bm)
	if err != nil {
		t.Fatalf("Failed to create LSM: %v", err)
	}

	keys := []string{"key000", "key001", "key999", "long1", "long2", "long3", "longkey0000", "longkey9999"}
	longValueA := make([]byte, 10000)
	for i := range longValueA {
		longValueA[i] = 'A'
	}
	longValueB := make([]byte, 10000)
	for i := range longValueB {
		longValueB[i] = 'B'
	}
	expectedValues := []string{
		"value000", "value001", "value999",
		string(longValueA), string(longValueA), string(longValueA),
		string(longValueB), string(longValueB),
	}

	for i, key := range keys {
		val, err := lsm.Get(key)
		if err != nil {
			t.Fatalf("Failed to get key '%s': %v", key, err)
		}
		if string(val) != expectedValues[i] {
			t.Fatalf("Expected value '%s' for key '%s', but got %s", expectedValues[i], key, val)
		}
	}
}

func genKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("key%03d", i)
	}
	return keys
}

func benchmarkGetSpecific(b *testing.B, multFiles bool) {
	r := rand.New(rand.NewSource(42))
	m, sst := flush1(b, multFiles)

	keys := genKeys(1000)
	for b.Loop() {
		key := keys[r.Intn(len(keys))]
		_, err := m.Get(key, sst)
		if err != nil {
			b.Fatalf("Error when getting key: %v", err)
		}
	}
}

func BenchmarkGetSpecificMultipleFiles(b *testing.B) {
	benchmarkGetSpecific(b, true)
}

func BenchmarkGetSpecificOneFile(b *testing.B) {
	benchmarkGetSpecific(b, false)
}
