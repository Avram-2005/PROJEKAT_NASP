package sstable

import (
	"fmt"
	"testing"
	"time"

	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func testLSM(t *testing.T, memtables []Memtable) (lsm *LSM) {
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

	for i, mem := range memtables {
		err := lsm.Flush(mem)
		if err != nil {
			t.Fatalf("Flush failed for memtable %d: %v", i, err)
		}
	}
	return lsm
}

func TestLSMFlushNoCompaction(t *testing.T) {
	lsm := testLSM(t, []Memtable{
		smallSmallKeyKVMemtable{},
	})
	if len(lsm.levels[0].tables) != 1 {
		t.Fatalf("Expected 2 SSTables in level 0, got %d", len(lsm.levels[0].tables))
	}
}

func TestLSMFlushL0Compaction(t *testing.T) {
	lsm := testLSM(t, []Memtable{
		smallSmallKeyKVMemtable{},
		manyLargeKeyKVMemtable{},
		smallSmallKeyKVMemtable{},
	})
	if len(lsm.levels[0].tables) != 1 {
		t.Fatalf("Expected 0 SSTables in level 0 after compaction, got %d", len(lsm.levels[0].tables))
	}
	if len(lsm.levels[1].tables) != 1 {
		t.Fatalf("Expected 1 SSTable in level 1 after compaction, got %d", len(lsm.levels[1].tables))
	}
}

func TestLSMFlushL1Compaction(t *testing.T) {
	lsm := testLSM(t, []Memtable{
		smallSmallKeyKVMemtable{},
		manyLargeKeyKVMemtable{},
		smallSmallKeyKVMemtable{},
		manyLargeKeyKVMemtable{},
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
	lsm := testLSM(t, []Memtable{
		tombstoneOlderMemtable{},
		tombstoneNewerMemtable{},
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

type largeOverlapMemtable struct {
	version int
	count   int
}

func (m largeOverlapMemtable) GetSortedEntries() []Record {
	ts := time.Unix(int64(1000+m.version), 0)
	entries := make([]Record, m.count)
	for i := range m.count {
		rec, _ := NewRecord(
			fmt.Sprintf("bulk-%05d", i),
			[]byte(fmt.Sprintf("v%02d-%05d", m.version, i)),
			false,
			ts,
		)
		entries[i] = *rec
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
		if err := lsm.Flush(largeOverlapMemtable{
			version: v,
			count:   keyCount,
		}); err != nil {
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
