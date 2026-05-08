package sstable

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"

	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func testLSM(t *testing.T, flushBatches [][]*Record) (lsm *LSM) {
	sstCfg := SSTableConfig{
		SummaryInterval: 1,
		MultipleFiles:   true,
	}
	lsmCfg := LSMConfig{
		NumLevels:        4,
		CompactionFactor: 2,
	}
	lsm, err := NewLSM(lsmCfg, t.TempDir(), sstCfg, bm)
	if err != nil {
		t.Fatalf("Failed to create LSM: %v", err)
	}

	for i, records := range flushBatches {
		err := lsm.Flush(records)
		if err != nil {
			t.Fatalf("Flush failed for memtable %d: %v", i, err)
		}
	}
	return lsm
}

func TestLSMFlushNoCompaction(t *testing.T) {
	lsm := testLSM(t, [][]*Record{
		smallSmallKeyKVMemtableEntries(),
	})
	if len(lsm.levels[0].tables) != 1 {
		t.Fatalf("Expected 2 SSTables in level 0, got %d", len(lsm.levels[0].tables))
	}
}

func TestLSMFlushL0Compaction(t *testing.T) {
	lsm := testLSM(t, [][]*Record{
		smallSmallKeyKVMemtableEntries(),
		manyLargeKeyKVMemtableEntries(),
		smallSmallKeyKVMemtableEntries(),
	})
	if len(lsm.levels[0].tables) != 1 {
		t.Fatalf("Expected 0 SSTables in level 0 after compaction, got %d", len(lsm.levels[0].tables))
	}
	if len(lsm.levels[1].tables) != 1 {
		t.Fatalf("Expected 1 SSTable in level 1 after compaction, got %d", len(lsm.levels[1].tables))
	}
}

func TestLSMFlushL1Compaction(t *testing.T) {
	lsm := testLSM(t, [][]*Record{
		smallSmallKeyKVMemtableEntries(),
		manyLargeKeyKVMemtableEntries(),
		smallSmallKeyKVMemtableEntries(),
		manyLargeKeyKVMemtableEntries(),
	})
	if len(lsm.levels[0].tables) != 0 {
		t.Fatalf("Expected 0 SSTables in level 0 after compaction, got %d", len(lsm.levels[0].tables))
	}
	if len(lsm.levels[1].tables) != 0 {
		t.Fatalf("Expected 0 SSTables in level 1 after compaction, got %d", len(lsm.levels[1].tables))
	}
	if len(lsm.levels[2].tables) != 1 {
		t.Fatalf("Expected 1 SSTable in level 2 after compaction, got %d", len(lsm.levels[2].tables))
	}
}

func TestLSMGetWithUpdatedAndDeletedKeysAfterCompaction(t *testing.T) {
	lsm := testLSM(t, [][]*Record{
		tombstoneOlderMemtableEntries(),
		tombstoneNewerMemtableEntries(),
	})

	keepVal, err := lsm.Get("keep")
	if err != nil {
		t.Fatalf("Failed to get key 'keep': %v", err)
	}
	if string(keepVal) != "keep-new" {
		t.Fatalf("Expected updated value 'keep-new' for key 'keep', got '%s'", string(keepVal))
	}

	freshVal, err := lsm.Get("fresh")
	if err != nil {
		t.Fatalf("Failed to get key 'fresh': %v", err)
	}
	if string(freshVal) != "fresh-val" {
		t.Fatalf("Expected value 'fresh-val' for key 'fresh', got '%s'", string(freshVal))
	}

	if _, err := lsm.Get("dead"); err == nil {
		t.Fatalf("Expected key 'dead' to be deleted after compaction")
	}
}

func largeOverlapMemtableEntries(version int, count int) []*Record {
	ts := time.Unix(int64(1000+version), 0)
	entries := make([]*Record, count)
	for i := range count {
		rec, _ := NewRecord(
			fmt.Sprintf("bulk-%05d", i),
			[]byte(fmt.Sprintf("v%02d-%05d", version, i)),
			false,
			ts,
		)
		entries[i] = rec
	}
	return entries
}

func TestLSMGetLargeDatasetNewestWins(t *testing.T) {
	const (
		versions = 5
		keyCount = 5000
	)

	sstCfg := SSTableConfig{
		SummaryInterval: 16,
		MultipleFiles:   true,
	}
	lsmCfg := LSMConfig{
		NumLevels:        4,
		CompactionFactor: 10,
	}
	lsm, err := NewLSM(lsmCfg, t.TempDir(), sstCfg, bm)
	if err != nil {
		t.Fatalf("Failed to create LSM: %v", err)
	}

	for v := range versions {
		if err := lsm.Flush(largeOverlapMemtableEntries(v, keyCount)); err != nil {
			t.Fatalf("Flush failed for version %d: %v", v, err)
		}
	}

	for i := range keyCount {
		key := fmt.Sprintf("bulk-%05d", i)
		expected := fmt.Sprintf("v%02d-%05d", versions-1, i)
		val, err := lsm.Get(key)
		if err != nil {
			t.Fatalf("Failed to get key '%s': %v", key, err)
		}
		if string(val) != expected {
			t.Fatalf("Expected '%s' for key '%s', got '%s'", expected, key, string(val))
		}
	}
}

// helper
func setupTestLSM(t *testing.T) (*LSM, func()) {
	// privremeni direktorijum za test
	tmpDir, err := os.MkdirTemp("", "lsm_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	bm, err := BlockManager.NewBlockManager(10, 4)
	if err != nil {
		t.Fatalf("Failed to create block manager: %v", err)
	}

	sstConfig := SSTableConfig{
		SummaryInterval: 2,
		MultipleFiles:   false,
	}

	lsmConfig := LSMConfig{
		NumLevels:        3,
		CompactionFactor: 2,
	}

	lsm, err := NewLSM(lsmConfig, tmpDir, sstConfig, bm)
	if err != nil {
		t.Fatalf("Failed to create LSM: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return lsm, cleanup
}

// helper za kreiranje Record-a
func makeRecord(key string, value string) *Record {
	return &Record{
		Key:       key,
		Value:     []byte(value),
		Tombstone: false,
		Timestamp: time.Now(),
	}
}

func TestLSMPrefixScan(t *testing.T) {
	lsm, cleanup := setupTestLSM(t)
	defer cleanup()

	records1 := []*Record{
		makeRecord("apple", "fruit1"),
		makeRecord("apricot", "fruit2"),
		makeRecord("banana", "fruit3"),
	}
	if err := lsm.Flush(records1); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	records2 := []*Record{
		makeRecord("berry", "fruit4"),
		makeRecord("blueberry", "fruit5"),
		makeRecord("cherry", "fruit6"),
	}
	if err := lsm.Flush(records2); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	results, err := lsm.PrefixScan("ap")
	if err != nil {
		t.Fatalf("PrefixScan error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 results for 'ap', got %d", len(results))
	}

	results, err = lsm.PrefixScan("b")
	if err != nil {
		t.Fatalf("PrefixScan error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("Expected 3 results for 'b', got %d", len(results))
	}

	results, err = lsm.PrefixScan("xyz")
	if err != nil {
		t.Fatalf("PrefixScan error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("Expected 0 results for 'xyz', got %d", len(results))
	}
}

func TestLSMRangeScan(t *testing.T) {
	lsm, cleanup := setupTestLSM(t)
	defer cleanup()
	records := []*Record{
		makeRecord("a", "1"),
		makeRecord("b", "2"),
		makeRecord("c", "3"),
		makeRecord("d", "4"),
		makeRecord("e", "5"),
		makeRecord("f", "6"),
		makeRecord("g", "7"),
		makeRecord("h", "8"),
	}
	if err := lsm.Flush(records); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}
	results, err := lsm.RangeScan("b", "e")
	if err != nil {
		t.Fatalf("RangeScan error: %v", err)
	}
	if len(results) != 4 {
		t.Fatalf("Expected 4 results for 'b'-'e', got %d", len(results))
	}
	expected := []string{"b", "c", "d", "e"}
	for i, r := range results {
		if r.Key != expected[i] {
			t.Fatalf("Expected %s, got %s", expected[i], r.Key)
		}
	}
	results, err = lsm.RangeScan("x", "z")
	if err != nil {
		t.Fatalf("RangeScan error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("Expected 0 results for 'x'-'z', got %d", len(results))
	}
}

func TestLSMPrefixScanWithTombstone(t *testing.T) {
	lsm, cleanup := setupTestLSM(t)
	defer cleanup()

	records := []*Record{
		makeRecord("test_key1", "value1"),
		makeRecord("test_key2", "value2"),
	}
	if err := lsm.Flush(records); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}
	tombstoneRec := &Record{
		Key:       "test_key1",
		Value:     nil,
		Tombstone: true,
		Timestamp: time.Now(),
	}
	if err := lsm.Flush([]*Record{tombstoneRec}); err != nil {
		t.Fatalf("Failed to flush tombstone: %v", err)
	}

	results, err := lsm.PrefixScan("test")
	if err != nil {
		t.Fatalf("PrefixScan error: %v", err)
	}
	for _, r := range results {
		if r.Tombstone {
			t.Fatalf("Tombstone should not appear in results: %s", r.Key)
		}
	}
}
func TestLSMPrefixScanMultipleLevels(t *testing.T) {
	lsm, cleanup := setupTestLSM(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		records := []*Record{
			makeRecord("key_a", "value"),
			makeRecord("key_b", "value"),
		}
		if err := lsm.Flush(records); err != nil {
			t.Fatalf("Failed to flush: %v", err)
		}
	}

	results, err := lsm.PrefixScan("key")
	if err != nil {
		t.Fatalf("PrefixScan error: %v", err)
	}

	uniqueKeys := make(map[string]bool)
	for _, r := range results {
		uniqueKeys[r.Key] = true
	}

	if len(uniqueKeys) != 2 {
		t.Fatalf("Expected 2 unique keys, got %d", len(uniqueKeys))
	}
}
