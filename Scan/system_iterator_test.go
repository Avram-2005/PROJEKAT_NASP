package scan

import (
	"testing"
	"time"

	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func TestSystemPrefixIterator(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()

	sstRecords := []*record.Record{
		makeRecord("apple", "sst_apple"),
		makeRecord("apricot", "sst_apricot"),
		makeRecord("cherry", "sst_cherry"),
	}

	time.Sleep(1 * time.Millisecond)
	time.Sleep(1 * time.Millisecond)

	scanner.memtable.Put("apple", []byte("mem_apple"))
	scanner.memtable.Put("apricot", []byte("mem_apricot"))
	scanner.memtable.Put("banana", []byte("mem_banana"))

	if err := scanner.lsm.Flush(sstRecords); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	iter, err := scanner.NewSystemPrefixIterator("ap")
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}
	defer iter.Stop()

	expectedKeys := []string{"apple", "apricot"}
	expectedValues := []string{"mem_apple", "mem_apricot"}

	count := 0
	for iter.Next() {
		if iter.Key() != expectedKeys[count] {
			t.Fatalf("Expected key %s, got %s", expectedKeys[count], iter.Key())
		}
		if string(iter.Value()) != expectedValues[count] {
			t.Fatalf("Expected value %s, got %s", expectedValues[count], iter.Value())
		}
		count++
	}
	if count != 2 {
		t.Fatalf("Expected 2 records, got %d", count)
	}
}

func TestSystemRangeIterator(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()

	scanner.memtable.Put("b", []byte("mem_b"))
	scanner.memtable.Put("d", []byte("mem_d"))

	time.Sleep(1 * time.Millisecond)
	sstRecords := []*record.Record{
		makeRecord("a", "sst_a"),
		makeRecord("c", "sst_c"),
		makeRecord("e", "sst_e"),
	}
	if err := scanner.lsm.Flush(sstRecords); err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	iter, err := scanner.NewSystemRangeIterator("b", "d")
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}
	defer iter.Stop()

	expectedKeys := []string{"b", "c", "d"}
	expectedValues := []string{"mem_b", "sst_c", "mem_d"}

	count := 0
	for iter.Next() {
		if iter.Key() != expectedKeys[count] {
			t.Fatalf("Expected key %s, got %s", expectedKeys[count], iter.Key())
		}
		if string(iter.Value()) != expectedValues[count] {
			t.Fatalf("Expected value %s, got %s", expectedValues[count], iter.Value())
		}
		count++
	}
	if count != 3 {
		t.Fatalf("Expected 3 records, got %d", count)
	}
}

func TestSystemIteratorEmpty(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()

	iter, err := scanner.NewSystemPrefixIterator("xyz")
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}
	defer iter.Stop()

	if iter.Next() {
		t.Fatal("Iterator should be empty")
	}
}

func TestSystemIteratorStop(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()

	scanner.memtable.Put("apple", []byte("fruit"))

	iter, err := scanner.NewSystemPrefixIterator("ap")
	if err != nil {
		t.Fatalf("Failed to create iterator: %v", err)
	}

	iter.Stop()

	if iter.Next() {
		t.Fatal("Next should return false after Stop")
	}
	if iter.Key() != "" {
		t.Fatal("Key should be empty after Stop")
	}
}
