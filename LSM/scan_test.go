package sstable

import (
	"testing"
)

func TestPrefixScan(t *testing.T) {
	mem := memtable{}
	m, sst, err := testFlush(t.TempDir(), mem, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// prefiks v
	rec, err := m.PrefixScan(sst, "v")
	if err != nil {
		t.Fatalf("PrefixScan failed: %v", err)
	}

	expected := []string{"v", "value1", "value2", "value3"}
	if len(rec) != len(expected) {
		t.Fatalf("Expected %d keys, got %d", len(expected), len(rec))
	}
	for i := range expected {
		if rec[i].Key != expected[i] {
			t.Errorf("Expected key %s, got %s", expected[i], rec[i].Key)
		}
	}

	// prefiks x, nema kljuceva
	m2, sst2, _ := testFlush(t.TempDir(), mem, true)
	rec, err = m2.PrefixScan(sst2, "x")
	if err != nil {
		t.Fatalf("PrefixScan failed: %v", err)
	}
	if len(rec) != 0 {
		t.Fatalf("Expected no results, but got %d", len(rec))
	}
}

func TestRangeScan(t *testing.T) {
	mem := memtable{}
	m, sst, err := testFlush(t.TempDir(), mem, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// range od v do value2
	rec, err := m.RangeScan(sst, "v", "value2")
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}

	expected := []string{"v", "value1", "value2"}
	if len(rec) != len(expected) {
		t.Fatalf("Expected %d keys, got %d", len(expected), len(rec))
	}
	for i := range expected {
		if rec[i].Key != expected[i] {
			t.Errorf("Expected key %s, got %s", expected[i], rec[i].Key)
		}
	}

	// range od w do x, nema kljuceva
	m2, sst2, _ := testFlush(t.TempDir(), mem, true)
	rec, err = m2.RangeScan(sst2, "w", "x")
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}
	if len(rec) != 0 {
		t.Fatalf("Expected no results, but got %d", len(rec))
	}
}
