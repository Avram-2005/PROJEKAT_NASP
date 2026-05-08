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
