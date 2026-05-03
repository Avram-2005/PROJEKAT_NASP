package sstable

import (
	"testing"
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
