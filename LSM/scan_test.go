package sstable

import (
	"testing"
	"time"

	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func scanMemtableEntries() []*Record {
	ts := time.Now()
	r1, _ := NewRecord("a", []byte("value3"), false, ts)
	r2, _ := NewRecord("bar", []byte("value3"), false, ts)
	r7, _ := NewRecord("j", []byte("value3"), false, ts)
	r71, _ := NewRecord("j", []byte{}, true, ts)
	r3, _ := NewRecord("v", []byte("value3"), false, ts)
	r4, _ := NewRecord("value1", []byte("value1"), false, ts)
	r5, _ := NewRecord("value2", []byte("value2"), false, ts)
	r6, _ := NewRecord("value3", []byte("value3"), false, ts)
	return []*Record{r1, r2, r7, r71, r3, r4, r5, r6}
}

func TestPrefixScan(t *testing.T) {
	mem := scanMemtableEntries()
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
	mem := scanMemtableEntries()
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
