package sstable

import (
	"testing"
	"time"

	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

type memtable struct {
}

func (m memtable) GetSortedEntries() []Record {
	ts := time.Now()
	r1, _ := NewRecord("aa", []byte("value3"), false, ts)
	r2, _ := NewRecord("bar", []byte("value3"), false, ts)
	r3, _ := NewRecord("v", []byte("value3"), false, ts)
	r4, _ := NewRecord("value1", []byte("value1"), false, ts)
	r5, _ := NewRecord("value2", []byte("value2"), false, ts)
	r6, _ := NewRecord("value3", []byte("value3"), false, ts)
	return []Record{*r1, *r2, *r3, *r4, *r5, *r6}
}

func TestPrefixIterator(t *testing.T) {
	mem := memtable{}

	m, sst, err := testFlush(t.TempDir(), mem, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	m2, sst2, err := testFlush(t.TempDir(), mem, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// prefiks v, treba da vrati value1, value2, value3 i v
	iter, err := m.NewPrefixIterator(sst, "v")
	if err != nil {
		t.Fatalf("Failed to create prefix iterator: %v", err)
	}
	defer iter.Close()

	var keys []string
	for {
		ok, err := iter.Next()
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
		if !ok {
			break
		}
		keys = append(keys, iter.iterator.Rec.Key)
	}

	expected := []string{"v", "value1", "value2", "value3"}
	if len(keys) != len(expected) {
		t.Fatalf("Expected %d keys, got %d: %v", len(expected), len(keys), keys)
	}
	for i := range expected {
		if keys[i] != expected[i] {
			t.Errorf("Expected key %s, got %s", expected[i], keys[i])
		}
	}

	// prefiks value, treba da vrati value1, value2 i value3
	iter2, err := m.NewPrefixIterator(sst, "value")
	if err != nil {
		t.Fatalf("Failed to create prefix iterator: %v", err)
	}
	defer iter2.Close()

	var keys2 []string
	for {
		ok, err := iter2.Next()
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
		if !ok {
			break
		}
		keys2 = append(keys2, iter2.iterator.Rec.Key)
	}

	expected2 := []string{"value1", "value2", "value3"}
	if len(keys2) != len(expected2) {
		t.Fatalf("Expected %d keys, got %d: %v", len(expected2), len(keys2), keys2)
	}
	for i := range expected2 {
		if keys2[i] != expected2[i] {
			t.Errorf("Expected key %s, got %s", expected2[i], keys2[i])
		}
	}

	// prefiks z, nema kljuceva
	iter3, err := m2.NewPrefixIterator(sst2, "z")
	if err != nil {
		t.Fatalf("Failed to create prefix iterator: %v", err)
	}
	defer iter3.Close()

	ok, err := iter3.Next()
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if ok {
		t.Fatal("Expected no results, but got one")
	}

	// prefiks v, ali se zaustaljamo posle 2. kljuca, znaci treba da vrati v i value1

	iter4, err := m.NewPrefixIterator(sst, "v")
	if err != nil {
		t.Fatalf("Failed to create prefix iterator: %v", err)
	}
	defer iter4.Close()

	var keys3 []string
	count := 0

	for {
		ok, err := iter4.Next()
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
		if !ok {
			break
		}
		keys3 = append(keys3, iter4.iterator.Rec.Key)
		count++

		if count == 2 {
			iter4.Stop()
		}
	}

	expected = []string{"v", "value1"}
	if len(keys3) != len(expected) {
		t.Fatalf("Expected %d keys, got %d: %v", len(expected), len(keys3), keys3)
	}
	for i := range expected {
		if keys3[i] != expected[i] {
			t.Errorf("Expected key %s, got %s", expected[i], keys3[i])
		}
	}
}
