package sstable

import (
	"fmt"
	"testing"
)

func TestLSMFlushNoCompaction(t *testing.T) {
	mem := smallSmallKeyKVMemtable{}
	config := LSMConfig{
		NumLevels:      4,
		NumFilesLevel0: 2,
	}
	lsm := NewLSM(bm, config)

	for range 2 {
		err := lsm.Flush(mem)
		if err != nil {
			t.Fatalf("Flush failed: %v", err)
		}
	}

	for i := range 2 {
		sstablePath := sstableFilepath(0, i)
		for i, key := range []string{"a", "b", "c"} {
			val, err := GetSpecific(key, sstablePath, bm)
			if err != nil {
				t.Fatalf("Failed to get key '%s' after flush: %v", key, err)
			}
			expectedValue := fmt.Sprintf("value%d", i+1)
			if string(val.Value) != expectedValue {
				t.Fatalf("Expected value '%s' for key '%s', but got %v", expectedValue, key, val)
			}
		}
	}
}
