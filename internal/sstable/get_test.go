package sstable

import (
	"testing"
)

func flush(t *testing.T, multFiles bool) {
	SetupSSTable(t.TempDir(), 100, multFiles)
	mem := manySmallKeyKVMemtable{}
	err := Flush(mem, 100, bm)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
}

func testGet(t *testing.T, key string, expectedValue string, multFiles bool) {
	flush(t, multFiles)

	val, err := Get(key, 100, bm)
	if err != nil {
		t.Fatalf("Failed to get key '%s': %v", key, err)
	}
	if string(val) != expectedValue {
		t.Fatalf("Expected value '%s' for key '%s', but got %s", expectedValue, key, val)
	}
}

func TestGetFirstKeyMultipleFiles(t *testing.T) {
	testGet(t, "key000", "value000", true)
}

func TestGetFirstKeyOneFile(t *testing.T) {
	testGet(t, "key000", "value000", false)
}

func TestGetLastKeyMultipleFiles(t *testing.T) {
	testGet(t, "key999", "value999", true)
}

func TestGetLastKeyOneFile(t *testing.T) {
	testGet(t, "key999", "value999", false)
}

func TestGetMiddleKeyMultipleFiles(t *testing.T) {
	testGet(t, "key500", "value500", true)
}

func TestGetMiddleKeyOneFile(t *testing.T) {
	testGet(t, "key500", "value500", false)
}

func TestGetNonExistentKeyMultipleFiles(t *testing.T) {
	flush(t, true)

	value, err := Get("nonexistent", 100, bm)
	if err != nil {
		t.Fatalf("Error when getting non-existent key: %v", err)
	}
	if value != nil {
		t.Fatalf("Expected nil value when getting non-existent key, but got %s", value)
	}
}

func TestGetNonExistentKeyOneFile(t *testing.T) {
	flush(t, false)

	value, err := Get("nonexistent", 100, bm)
	if err != nil {
		t.Fatalf("Error when getting non-existent key: %v", err)
	}
	if value != nil {
		t.Fatalf("Expected nil value when getting non-existent key, but got %s", value)
	}
}
