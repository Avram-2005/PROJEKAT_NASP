package memtable

import (
	"fmt"
	"testing"
	"time"

	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
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
			m, _ := NewMemtableManager(2, conf, func(entries []*record.Record) error {
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

func TestManagerDelete(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 10)
			m.Put("delete_key", []byte("delete_value"))
			val, found, _ := m.Get("delete_key")
			if !found || string(val) != "delete_value" {
				t.Fatal("Key should exist before delete")
			}
			err := m.Delete("delete_key", time.Now())
			if err != nil {
				t.Fatalf("Delete failed: %v", err)
			}
			_, found, _ = m.Get("delete_key")
			if found {
				t.Fatal("Key still exists after logical delete")
			}
		})
	}
}

func TestManagerDeleteNonExistent(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 10)
			err := m.Delete("nonexistent_key_12345", time.Now())
			if err != nil {
				t.Fatalf("Delete of non-existent key should not return error, got %v", err)
			}
		})
	}
}

func TestManagerGetAfterDeleteAndRotation(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 2, 2)
			m.Put("key1", []byte("value1"))
			m.Put("key2", []byte("value2"))
			m.Delete("key1", time.Now())
			val, found, _ := m.Get("key2")
			if !found || string(val) != "value2" {
				t.Fatal("key2 should exist before rotation")
			}
			m.Put("key3", []byte("value3"))
			_, found, _ = m.Get("key2")
			if found {
				t.Fatal("key2 still in memory,should have been flushed(no SSTable integration)")
			}
			_, found, _ = m.Get("key1")
			if found {
				t.Fatal("key1 should still be deleted after rotation")
			}
			val, found, _ = m.Get("key3")
			if !found || string(val) != "value3" {
				t.Fatal("key3 should exist and have correct value")
			}
		})
	}
}

func TestGetRecord(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 10)
			m.Put("test_key", []byte("test_value"))
			rec, found, err := m.GetRecord("test_key")
			if err != nil {
				t.Fatalf("GetRecord error: %v", err)
			}
			if !found {
				t.Fatal("Key not found")
			}
			if rec.Key != "test_key" {
				t.Fatalf("Wrong key, expected test_key, got %s", rec.Key)
			}
			if string(rec.Value) != "test_value" {
				t.Fatalf("Wrong value, expected test_value, got %s", rec.Value)
			}
			if rec.Tombstone {
				t.Fatal("Tombstone should be false for existing key")
			}
			if rec.Timestamp.IsZero() {
				t.Fatal("Timestamp is zero")
			}
		})
	}
}

func TestGetRecordAfterDelete(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 10)
			m.Put("delete_key", []byte("delete_value"))
			m.Delete("delete_key", time.Now())
			rec, found, err := m.GetRecord("delete_key")
			if err != nil {
				t.Fatalf("GetRecord error: %v", err)
			}
			if !found {
				t.Fatal(" Key should exist as tombstone")
			}
			if !rec.Tombstone {
				t.Fatal("Tombstone should be true after delete")
			}
			if rec.Value == nil {
				t.Fatal("Value should not be nil for tombstone")
			}
		})
	}
}

func TestGetRecordNonExistent(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 10)

			rec, found, err := m.GetRecord("nonexistent_key_12345")
			if err != nil {
				t.Fatalf("GetRecord error: %v", err)
			}
			if found {
				t.Fatal("Should return found=false for non-existent key")
			}
			if rec != nil {
				t.Fatal("Should return nil record for non-existent key")
			}
		})
	}
}

func TestGetRecordAfterRotation(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 2, 2)
			m.Put("key1", []byte("value1"))
			m.Put("key2", []byte("value2"))
			m.Delete("key1", time.Now())
			m.Put("key3", []byte("value3"))
			rec, found, err := m.GetRecord("key1")
			if err != nil {
				t.Fatalf("GetRecord error: %v", err)
			}
			if !found {
				t.Fatal("Key1 should exist as tombstone")
			}
			if !rec.Tombstone {
				t.Fatal("Key1 should be tombstone")
			}
			rec2, found2, err := m.GetRecord("key2")
			if err != nil {
				t.Fatalf("GetRecord error for key2: %v", err)
			}
			if found2 && rec2.Tombstone {
				t.Fatal("Key2 should not be tombstone if found")
			}
			rec3, found3, err := m.GetRecord("key3")
			if err != nil {
				t.Fatalf("GetRecord error for key3: %v", err)
			}
			if !found3 {
				t.Fatal("Key3 not found")
			}
			if string(rec3.Value) != "value3" {
				t.Fatalf("Key3 wrong value, got %s", rec3.Value)
			}
		})
	}
}

func TestMemtablePrefixScan(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 10)
			m.Put("apple", []byte("fruit1"))
			m.Put("apricot", []byte("fruit2"))
			m.Put("banana", []byte("fruit3"))

			m.Put("berry", []byte("fruit4"))
			m.Put("blueberry", []byte("fruit5"))

			results := m.PrefixScan("ap")
			if len(results) != 2 {
				t.Fatalf("Expected 2 results for prefix 'ap', got %d", len(results))
			}
			if results[0].Key != "apple" || results[1].Key != "apricot" {
				t.Fatalf("Wrong results for prefix 'ap': %v", results)
			}

			results = m.PrefixScan("b")
			if len(results) != 3 {
				t.Fatalf("Expected 3 results for prefix 'b', got %d", len(results))
			}

			// Prefix scan za nepostojeci prefix
			results = m.PrefixScan("xyz")
			if len(results) != 0 {
				t.Fatalf("Expected 0 results for prefix 'xyz', got %d", len(results))
			}
		})
	}
}

func TestMemtablePrefixScanAfterDelete(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 10)

			m.Put("test_key1", []byte("value1"))
			m.Put("test_key2", []byte("value2"))
			m.Delete("test_key1", time.Now())

			results := m.PrefixScan("test")
			if len(results) != 1 {
				t.Fatalf("Expected 1 result after delete, got %d", len(results))
			}
			if results[0].Key != "test_key2" {
				t.Fatalf("Expected test_key2, got %s", results[0].Key)
			}
		})
	}
}
func TestMemtableRangeScan(t *testing.T) {
	for _, typ := range allStructTypes {
		t.Run(typ, func(t *testing.T) {
			m := makeManager(t, typ, 3, 10)
			m.Put("a", []byte("1"))
			m.Put("b", []byte("2"))
			m.Put("c", []byte("3"))
			m.Put("d", []byte("4"))
			m.Put("e", []byte("5"))

			results := m.RangeScan("b", "d")
			if len(results) != 3 {
				t.Fatalf("Expected 3 results for range 'b'-'d', got %d", len(results))
			}
			expected := []string{"b", "c", "d"}
			for i, r := range results {
				if r.Key != expected[i] {
					t.Fatalf("Expected %s, got %s", expected[i], r.Key)
				}
			}

			results = m.RangeScan("a", "b")
			if len(results) != 2 {
				t.Fatalf("Expected 2 results for range ''-'b', got %d", len(results))
			}

			results = m.RangeScan("x", "z")
			if len(results) != 0 {
				t.Fatalf("Expected 0 results for range 'x'-'z', got %d", len(results))
			}
		})
	}
}
