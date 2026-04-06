package memtable

import (
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
