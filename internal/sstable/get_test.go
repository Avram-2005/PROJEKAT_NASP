package sstable

import (
	"testing"
)

var flushed bool

func flush(t *testing.T) {
	SetupDirectory(t.TempDir())
	mem := manySmallKeyKVMemtable{}
	err := Flush(mem, 100, bm)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	flushed = true
}

func TestGetFirstKey(t *testing.T) {
	flush(t)

	val, err := Get("key000", 100, bm)
	if err != nil {
		t.Fatalf("Failed to get key 'key000': %v", err)
	}
	expectedValue := "value000"
	if string(val) != expectedValue {
		t.Fatalf("Expected value '%s' for key 'key000', but got %s", expectedValue, val)
	}
}

func TestGetLastKey(t *testing.T) {
	flush(t)

	val, err := Get("key999", 100, bm)
	if err != nil {
		t.Fatalf("Failed to get key 'key999': %v", err)
	}
	expectedValue := "value999"
	if string(val) != expectedValue {
		t.Fatalf("Expected value '%s' for key 'key999', but got %s", expectedValue, val)
	}
}

func TestGetNonExistentKey(t *testing.T) {
	flush(t)

	_, err := Get("nonexistent", 100, bm)
	if err == nil {
		t.Fatalf("Expected error when getting non-existent key, but got none")
	}
}

func TestGetSummaryIntervalKey(t *testing.T) {
	flush(t)

	val, err := Get("key500", 100, bm)
	if err != nil {
		t.Fatalf("Failed to get key 'key500': %v", err)
	}
	expectedValue := "value500"
	if string(val) != expectedValue {
		t.Fatalf("Expected value '%s' for key 'key500', but got %s", expectedValue, val)
	}
}
