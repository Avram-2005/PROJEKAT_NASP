package memtable

import (
	"fmt"
	"testing"
)

// helper funckija
func makeManager(t *testing.T, structType string, maxCount int, maxEntries int) *MemtableManager {
	t.Helper()
	conf := MemtableConfig{
		Type:              structType,
		MaxSizeEntries:    maxEntries,
		SkipListMaxHeight: 8,
		BPlusTreeDegree:   2,
	}
	m, err := NewMemtableManager(maxCount, conf, nil)
	if err != nil {
		t.Fatalf("[%s] Failed to load NewMemtableManager %v", structType, err)
	}
	return m
}

// konstruktor
func TestNewManagerInvalidMaxCount(t *testing.T) {
	conf := MemtableConfig{
		Type:           "hashmap",
		MaxSizeEntries: 5,
	}
	_, err := NewMemtableManager(0, conf, nil)
	if err == nil {
		t.Fatal("Expected error when maxCount is one")
	}
}

func TestNewManagerOneInstance(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 5)
			if m.InstanceCount() != 1 {
				t.Fatalf("Expected one instance at the start, got %d isntead", m.InstanceCount())
			}
		})
	}
}

// Put i Get
func TestPutAndGet(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 10)
			m.Put("kljuc1", []byte("val1"))
			m.Put("kljuc2", []byte("val2"))
			val, found, err := m.Get("kljuc1")
			if err != nil || !found || string(val) != "val1" {
				t.Fatalf("Get was unsuccessful, got found-%v, val-%s,err-%v", found, val, err)
			}
		})
	}
}

func TestGetNonExistentManager(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 10)
			_, found, err := m.Get("nonexistent_key")
			if err != nil || found {
				t.Fatalf("Expected found variable to be false, but got found-%v,err %v instead", found, err)
			}
		})
	}
}

// Total Size
func TestTotalSize(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 2)
			m.Put("k1", []byte("val1"))
			m.Put("k2", []byte("val2"))
			m.Put("k3", []byte("val3"))
			if m.TotalSize() < 2 {
				t.Fatalf("Expected Total size to be greater than 2, got %d instead", m.TotalSize())
			}
		})
	}
}

// rotacija
func TestRotationAndNewInstances(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 2) //nakon 2 upisa treba da se kreira nova instanca
			m.Put("k1", []byte("val1"))
			m.Put("k2", []byte("val2"))
			if m.InstanceCount() != 2 {
				t.Fatalf("Expeted 2 instances after rotation, got %d instead", m.InstanceCount())
			}
		})
	}
}

func TestMaxInstancesNotExceeded(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			flushed := 0
			conf := MemtableConfig{
				Type:              typ,
				MaxSizeEntries:    2,
				SkipListMaxHeight: 8,
				BPlusTreeDegree:   2,
			}
			m, _ := NewMemtableManager(2, conf, func(entries []KeyValue) error {
				flushed++
				return nil
			})
			for i := 0; i < 4; i++ {
				m.Put(fmt.Sprintf("kljuc%d", i), []byte(fmt.Sprintf("val%d", i)))

			}
			if m.InstanceCount() > 2 {
				t.Fatalf("Number of instances must not exceed 2, got %d", m.InstanceCount())
			}
			if flushed == 0 {
				t.Fatal("Expected flush, but it didn't happen")
			}
		})
	}
}
