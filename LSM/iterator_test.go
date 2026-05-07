package sstable

import (
	"testing"
	"time"

	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func iteratorMemtableEntries() []*Record {
	ts := time.Now()
	r1, _ := NewRecord("a", []byte("value3"), false, ts)
	r2, _ := NewRecord("bar", []byte("value3"), false, ts)
	r7, _ := NewRecord("j", []byte("value3"), false, ts)
	r71, _ := NewRecord("j", []byte{}, true, ts)
	r3, _ := NewRecord("v", []byte("value3"), false, ts)
	r4, _ := NewRecord("value1", []byte("value1"), false, ts)
	r5, _ := NewRecord("value2", []byte("value2"), false, ts)
	r6, _ := NewRecord("value3", []byte("value3"), false, ts)
	return []*Record{r1, r2, r7, r71, r3, r4, r5, r6}
}

func TestPrefixIterator(t *testing.T) {
	entries := iteratorMemtableEntries()

	m, sst, err := testFlush(t.TempDir(), entries, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	m2, sst2, err := testFlush(t.TempDir(), entries, true)
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

	// prefiks j, treba da vrati 2 kljuca j
	iter5, err := m.NewPrefixIterator(sst, "j")
	if err != nil {
		t.Fatalf("Failed to create prefix iterator: %v", err)
	}
	defer iter5.Close()

	var keys4 []string
	for {
		ok, err := iter5.Next()
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
		if !ok {
			break
		}
		keys4 = append(keys4, iter5.iterator.Rec.Key)
	}

	expected = []string{"j", "j"}
	if len(keys4) != len(expected) {
		t.Fatalf("Expected %d keys, got %d: %v", len(expected), len(keys4), keys4)
	}
	for i := range expected {
		if keys4[i] != expected[i] {
			t.Errorf("Expected key %s, got %s", expected[i], keys4[i])
		}
	}
}

func TestRangeIterator(t *testing.T) {
	entries := iteratorMemtableEntries()

	m, sst, err := testFlush(t.TempDir(), entries, false)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	m2, sst2, err := testFlush(t.TempDir(), entries, true)
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// range od a do c, treba da vrati kljuceve a i bar
	iter, err := m.NewRangeIterator(sst, "a", "c")
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

	expected := []string{"a", "bar"}
	if len(keys) != len(expected) {
		t.Fatalf("Expected %d keys, got %d: %v", len(expected), len(keys), keys)
	}
	for i := range expected {
		if keys[i] != expected[i] {
			t.Errorf("Expected key %s, got %s", expected[i], keys[i])
		}
	}

	// range od w do z, nema kljuceva
	iter2, err := m2.NewRangeIterator(sst2, "w", "z")
	if err != nil {
		t.Fatalf("Failed to create prefix iterator: %v", err)
	}
	defer iter2.Close()

	ok, err := iter2.Next()
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if ok {
		t.Fatal("Expected no results, but got one")
	}

	// range od b do a, b je vece od a, nema kljuceva
	iter3, err := m2.NewRangeIterator(sst2, "b", "a")
	if err != nil {
		t.Fatalf("Failed to create prefix iterator: %v", err)
	}
	defer iter3.Close()

	ok, err = iter3.Next()
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if ok {
		t.Fatal("Expected no results, but got one")
	}

	// range od j do j, treba da vrati 2 kljuca j
	iter5, err := m.NewRangeIterator(sst, "j", "j")
	if err != nil {
		t.Fatalf("Failed to create prefix iterator: %v", err)
	}
	defer iter5.Close()

	var keys4 []string

	for {
		ok, err := iter5.Next()
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}
		if !ok {
			break
		}
		keys4 = append(keys4, iter5.iterator.Rec.Key)
	}

	expected = []string{"j", "j"}
	if len(keys4) != len(expected) {
		t.Fatalf("Expected %d keys, got %d: %v", len(expected), len(keys4), keys4)
	}
	for i := range expected {
		if keys4[i] != expected[i] {
			t.Errorf("Expected key %s, got %s", expected[i], keys4[i])
		}
	}
}
