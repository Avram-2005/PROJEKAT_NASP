package sstable

import (
	"fmt"
	"math/rand"
	"testing"
)

var flushCounter int
var testTempDirs = make(map[testHelper]string)

func getNextFlushNum() int {
	flushCounter++
	return flushCounter
}

type testHelper interface {
	Helper()
	Fatalf(format string, args ...any)
	TempDir() string
}

func flush(t testHelper, multFiles bool, mem Memtable) string {
	flushNum := getNextFlushNum()

	// Get or create temp directory for this test
	tempDir, exists := testTempDirs[t]
	if !exists {
		tempDir = t.TempDir()
		testTempDirs[t] = tempDir
	}

	m, err := SetupSSTableManager(tempDir, SSTableConfig{
		SummaryInterval: 10,
		MultipleFiles:   multFiles,
	}, bm)
	if err != nil {
		t.Fatalf("Failed to setup SSTable: %v", err)
	}
	sst, err := m.Flush(mem, flushNum)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	return sst.path
}

func flush1(t testHelper, multFiles bool) string {
	return flush(t, multFiles, &manySmallKeyKVMemtable{})
}

func flush2(t testHelper, multFiles bool) string {
	return flush(t, multFiles, &fewLargeKeyKVMemtable{})
}

func flush3(t testHelper, multFiles bool) string {
	return flush(t, multFiles, &manyLargeKeyKVMemtable{})
}

func testGetSpecific(t *testing.T, key string, expectedValue string, multFiles bool) {
	sstablePath := flush1(t, multFiles)

	val, err := GetSpecific(key, sstablePath, bm)
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
	sstablePath := flush1(t, true)

	value, err := GetSpecific("nonexistent", sstablePath, bm)
	if err != nil {
		t.Fatalf("Error when getting non-existent key: %v", err)
	}
	if value != nil {
		t.Fatalf("Expected nil value when getting non-existent key, but got %v", value)
	}
}

func TestGetSpecificNonExistentKeyOneFile(t *testing.T) {
	sstablePath := flush1(t, false)

	value, err := GetSpecific("nonexistent", sstablePath, bm)
	if err != nil {
		t.Fatalf("Error when getting non-existent key: %v", err)
	}
	if value != nil {
		t.Fatalf("Expected nil value when getting non-existent key, but got %v", value)
	}
}

func TestGet(t *testing.T) {
	flush1(t, true)
	flush2(t, false)
	flush3(t, false)

	m, err := SetupSSTableManager(testTempDirs[t], SSTableConfig{
		SummaryInterval: 10,
		MultipleFiles:   true,
	}, bm)
	if err != nil {
		t.Fatalf("Failed to setup SSTableManager: %v", err)
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
		val, err := m.Get(key, bm)
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
	sstablePath := flush1(b, multFiles)

	keys := genKeys(1000)
	for b.Loop() {
		key := keys[r.Intn(len(keys))]
		_, err := GetSpecific(key, sstablePath, bm)
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
