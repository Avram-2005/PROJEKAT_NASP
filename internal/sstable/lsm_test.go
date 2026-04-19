package sstable

import (
	"testing"
)

func testLSM(t *testing.T, memtables []Memtable) {
	sstCfg := SSTableConfig{
		SummaryInterval: 1,
		MultipleFiles:   true,
	}
	lsmCfg := LSMConfig{
		NumLevels:      4,
		NumFilesLevel0: 2,
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
}

func TestLSMFlushNoCompaction(t *testing.T) {
	testLSM(t, []Memtable{
		smallSmallKeyKVMemtable{}, manyLargeKeyKVMemtable{},
	})
}

func TestLSMFlushL0Compaction(t *testing.T) {
	testLSM(t, []Memtable{
		smallSmallKeyKVMemtable{},
		manyLargeKeyKVMemtable{},
		smallSmallKeyKVMemtable{},
	})
}
