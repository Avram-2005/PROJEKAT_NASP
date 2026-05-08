package memtable

import (
	"fmt"
	"testing"
	"time"

	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

// helper funkcija za ostale testove
func makeAdapter(t *testing.T, structType string) *MemtableAdapter {
	t.Helper()
	conf := MemtableConfig{
		Type:              structType,
		MaxSizeEntries:    10,
		SkipListMaxHeight: 8,
		BPlusTreeDegree:   2,
	}
	adapt, err := NewMemtableAdapter(conf)
	if err != nil {
		t.Fatalf("[%s]NewMemtableAdapter failed: %v", structType, err)
	}
	return adapt
}

var allStructTypes = []string{"hashmap", "skip_list", "b_plus_tree"}

// konstruktor
func TestNewAdapterValidTypes(t *testing.T) {
	for _, typ := range allStructTypes {
		adapt := makeAdapter(t, typ)
		if adapt == nil {
			t.Fatalf("[%s] adapter is nil", typ)
		}
	}
}

func TestNewAdapterInvalidType(t *testing.T) {
	conf := MemtableConfig{Type: "invalid_type"}
	_, err := NewMemtableAdapter(conf)
	if err == nil {
		t.Fatal("Expected error for invalid type, got nil")
	}
}

// put i get
func TestPutGet(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			err := adapt.Put("kljuc1", []byte("vrednost1"))
			if err != nil {
				t.Fatalf("Put operation failed: %v", err)
			}
			val, found, err := adapt.Get("kljuc1")
			if err != nil {
				t.Fatalf("Get operation failed: %v", err)
			}
			if !found {
				t.Fatalf("Expected the bool to be true in found, but found false")
			}
			if string(val) != "vrednost1" {
				t.Fatalf("Expected 'vrednost1', got %s", val)
			}
		})
	}
}

func TestPutRecord(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			ts := time.Now().Add(-24 * time.Hour)
			rec := &record.Record{
				Key:       "test_key",
				Value:     []byte("test_value"),
				Tombstone: false,
				Timestamp: ts,
			}
			err := adapt.PutRecord(rec)
			if err != nil {
				t.Fatalf("PutRecord failed: %v", err)
			}
			val, found, err := adapt.Get("test_key")
			if !found {
				t.Fatal("Key not found after PutRecord")
			}
			if string(val) != "test_value" {
				t.Fatalf("Value corrupted: expected test_value, got %s", val)
			}
		})
	}
}
func TestGetNonExistentKey(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			_, found, err := adapt.Get("nonexisting_key")
			if err != nil {
				t.Fatalf("Get function error: %v", err)

			}
			if found {
				t.Fatal("Expected found to be false, but got true")
			}
		})
	}
}

func TestPutNewValExistingKey(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			adapt.Put("kljuc1", []byte("vrednost1"))
			adapt.Put("kljuc1", []byte("nova_vrednost"))
			val, found, _ := adapt.Get("kljuc1")
			if !found {
				t.Fatal("Key doesn't exist after update")
			}
			if string(val) != "nova_vrednost" {
				t.Fatalf("Expected 'nova_vrednost', but got %s", val)
			}
		})
	}
}

// size
func TestSizeAfterPut(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			if adapt.Size() != 0 {
				t.Fatalf("Expected size=0 at the start, but got: %d", adapt.Size())
			}
			adapt.Put("k1", []byte("v1"))
			adapt.Put("k2", []byte("v2"))
			adapt.Put("k3", []byte("v3"))
			if adapt.Size() != 3 {
				t.Fatalf("Expected size=3, but got %d", adapt.Size())
			}
		})
	}
}
func TestSizeAfterUpdate(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			adapt.Put("k1", []byte("val1"))
			adapt.Put("k1", []byte("new_val"))
			if adapt.Size() != 1 {
				t.Fatalf("Expected size=1 after update, but got %d", adapt.Size())
			}
		})
	}
}

// Delete
func TestDeleteExistingKey(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			adapt.Put("kljuc1", []byte("val1"))
			deleted, err := adapt.Delete("kljuc1")
			if err != nil {
				t.Fatalf("Delete operation error: %v", err)
			}
			if !deleted {
				t.Fatal("Expected to get true, got false instead")
			}
			_, found, _ := adapt.Get("kljuc1")
			if found {
				t.Fatal("Key still exists after delition")
			}
		})
	}
}

func TestDeleteNonExistentKey(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			deleted, err := adapt.Delete("nonexisting_key")
			if err != nil {
				t.Fatalf("Delete shouldn't cause a fatal error: %v", err)
			}
			if deleted {
				t.Fatal("Expected to get false in deleted, but god true instead")
			}
		})
	}
}

func TestSizeAfterDelete(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			adapt.Put("kljuc1", []byte("val1"))
			adapt.Put("kljuc2", []byte("val2"))
			adapt.Delete("kljuc1")
			if adapt.Size() != 1 {
				t.Fatalf("Expected size to be 1 after deletion,but got %d", adapt.Size())
			}
		})
	}
}

// IsEmpty i Clear funkcionalnosti
func TestIsEmpty(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			if !adapt.IsEmpty() {
				t.Fatal("Expected to be empty at the start")
			}
			adapt.Put("kljuc1", []byte("val1"))
			if adapt.IsEmpty() {
				t.Fatal("Expected to not be empty aftera Put operation")
			}
		})
	}
}

func TestClear(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			adapt.Put("kljuc1", []byte("val1"))
			adapt.Put("kljuc2", []byte("val2"))
			adapt.Clear()
			if adapt.Size() != 0 {
				t.Fatalf("Expected size to be 0 after calling Clear, but got %d instead", adapt.Size())
			}
			if adapt.TotalEntries() != 0 {
				t.Fatalf("Expected the number of total entries to be 0 after Clear(), got %d instead", adapt.TotalEntries())
			}
			if !adapt.IsEmpty() {
				t.Fatal("Expected the table to be empty after Clear()")
			}
		})
	}
}

// GetSortedEntries
func TestGetSortedEntries(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			adapt.Put("narandza", []byte("val1"))
			adapt.Put("ananas", []byte("val2"))
			adapt.Put("sarma", []byte("val3"))
			entries := adapt.GetSortedEntries()
			if len(entries) != 3 {
				t.Fatalf("Expected 3 entries, got %d", len(entries))
			}
			expected := []string{"ananas", "narandza", "sarma"}
			for i, e := range entries {
				if e.Key != expected[i] {
					t.Fatalf("On position %d expected '%s', got '%s' instead", i, expected[i], e.Key)
				}
			}
		})
	}
}

// RangeScan
func TestRangeScan(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			for i := 0; i < 5; i++ {
				adapt.Put(fmt.Sprintf("kljuc%d", i), []byte(fmt.Sprintf("val%d", i)))
			}
			results := adapt.RangeScan("kljuc1", "kljuc3")
			if len(results) != 3 {
				t.Fatalf("Expected 3 results, got %d", len(results))
			}
			if results[0].Key != "kljuc1" || results[2].Key != "kljuc3" {
				t.Fatalf("Wrong range selected %v", results)
			}
		})
	}
}

func TestRangeScanNoResults(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			adapt.Put("kljuc1", []byte("val1"))
			results := adapt.RangeScan("x", "z")
			if len(results) != 0 {
				t.Fatalf("Expected 0 results, but got %d instead", len(results))
			}
		})
	}
}

// Prefix Scan
func TestPrefixScan(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			adapt.Put("kljuc_1", []byte("val1"))
			adapt.Put("kljuc_2", []byte("val2"))
			adapt.Put("vred_kljuca_3", []byte("val3"))
			adapt.Put("kljuc_4", []byte("val4"))
			results := adapt.PrefixScan("kljuc")
			if len(results) != 3 {
				t.Fatalf("Expected 3 results, got %d instead", len(results))
			}
			for _, r := range results {
				if len(r.Key) < 3 || r.Key[:5] != "kljuc" {
					t.Fatalf("Key '%s' doesn't start with 'kljuc'", r.Key)
				}
			}
		})
	}
}

func TestPrefixScanNoResults(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			adapt.Put("kljuc1", []byte("val1"))
			results := adapt.PrefixScan("bab")
			if len(results) != 0 {
				t.Fatalf("Expected 0 results, got %d instead", len(results))
			}
		})
	}
}

// Iterator
func TestIterator(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			adapt.Put("banana", []byte("val1"))
			adapt.Put("ananas", []byte("val2"))
			adapt.Put("citurs", []byte("val3"))
			iter := adapt.Iterator()
			count := 0
			prevKey := ""
			for iter.Next() {
				if iter.Key() <= prevKey && prevKey != "" {
					t.Fatalf("Iterator is not sorted: '%s' after '%s'", iter.Key(), prevKey)
				}
				prevKey = iter.Key()
				count++
			}
			if count != 3 {
				t.Fatalf("Expected 3 iterations, got %d", count)
			}
		})
	}
}

func TestIteratorEmpty(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			iter := adapt.Iterator()
			if iter.Next() {
				t.Fatal("Empty adapter cnnot have true in variable next")
			}
		})
	}
}

// Flush i IsFull
func TestShouldFlush(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			conf := MemtableConfig{
				Type:              typ,
				MaxSizeEntries:    3,
				SkipListMaxHeight: 8,
				BPlusTreeDegree:   2,
			}
			adapt, _ := NewMemtableAdapter(conf)
			adapt.Put("kljuc1", []byte("val1"))
			adapt.Put("kljuc2", []byte("val2"))
			if adapt.ShouldFlush() {
				t.Fatal("Cannot flush with 2 out of 3 inputs")
			}
			adapt.Put("kljuc3", []byte("val3"))
			if !adapt.ShouldFlush() {
				t.Fatal("Should flush with a full table")
			}
		})
	}
}

func TestShouldFlushByBytes(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			conf := MemtableConfig{
				Type:              typ,
				MaxSizeBytes:      250,
				SkipListMaxHeight: 8,
				BPlusTreeDegree:   2,
			}
			adapt, _ := NewMemtableAdapter(conf)
			adapt.Put("a", []byte("1"))
			adapt.Put("b", []byte("2"))
			if adapt.ShouldFlush() {
				t.Fatal("Cannot flush with 200 out of 250 bytes")
			}
			adapt.Put("c", []byte("3"))
			if !adapt.ShouldFlush() {
				t.Fatal("Should have flushed since the bytes exeed 250 bytes")
			}

		})
	}
}

// tombstone
func TestTombstone(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			adapt := makeAdapter(t, typ)
			adapt.Put("k1", []byte("v1"))
			deleted, err := adapt.Delete("k1")
			if err != nil || !deleted {
				t.Fatalf("Delete failed: %v", err)
			}
			_, found, _ := adapt.Get("k1")
			if found {
				t.Fatal("Key still exists after logical delete")
			}
			entries := adapt.GetSortedEntries()
			for _, e := range entries {
				if e.Key == "k1" && !e.Tombstone {
					t.Fatal("Tombstone not persisted in GetSortedEntries")
				}
			}
		})
	}
}

func TestPutRecordInManager(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 10)
			ts := time.Now().Add(-24 * time.Hour)
			rec := &record.Record{
				Key:       "recovery_key",
				Value:     []byte("recovery_val"),
				Tombstone: false,
				Timestamp: ts,
			}
			err := m.PutRecord(rec)
			if err != nil {
				t.Fatalf("PutRecord failed: %v", err)
			}
			val, found, err := m.Get("recovery_key")
			if err != nil || !found {
				t.Fatalf("Get failed: found=%v, err=%v", found, err)
			}
			if string(val) != "recovery_val" {
				t.Fatalf("Wrong value: %s", val)
			}
		})
	}
}
