package memtable

import (
	"fmt"
	"testing"
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
