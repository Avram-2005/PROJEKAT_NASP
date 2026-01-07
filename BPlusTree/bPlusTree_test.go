package BPlusTree

import (
	"fmt"
	"testing"
)

func TestBPlusTreeOperations(t *testing.T) {
	tree, _ := NewBPlusTree(3)
	err := tree.Insert("kljuc1", []byte("vrednost1"))
	if err != nil {
		t.Errorf("Insert action failed: %v", err)
	}
	if tree.Size() != 1 {
		t.Errorf("Expected size 1, got %d", tree.Size())
	}

	//Pretraga
	val, found := tree.Search("kljuc1")
	if !found {
		t.Error("Search action failed to find the given key")
	}
	if string(val) != "vrednost1" {
		t.Errorf("Expected 'vrednost1' but got %s", string(val))
	}

	//Update
	err = tree.Insert("kljuc1", []byte("novaVred"))
	if err != nil {
		t.Errorf("Update action failed: %v", err)
	}
	val, _ = tree.Search("kljuc1")
	if string(val) != "novaVred" {
		t.Errorf("Update was not succsesful, got: %s", string(val))
	}

	//Brisanje
	deleted := tree.Delete("kljuc1")
	if !deleted {
		t.Error("Delete action failed")
	}

	if tree.Size() != 0 {
		t.Errorf("Expected size 0 after deletion, got %d", tree.Size())
	}

	_, found = tree.Search("key1")
	if found {
		t.Error("Key should not have been found after deletion")
	}

}

func BPlusTreePrefixScanTest(t *testing.T) {
	tree, _ := NewBPlusTree(3)

	keys := []string{
		"pas", "macka", "test1", "papagaj", "test2", "tigar", "test3", "petao", "test4", "delfin",
	}
	for _, key := range keys {
		tree.Insert(key, []byte("vrednost_"+key))
	}

	results := tree.PrefixScan("test")
	fmt.Printf("Found %d keys with prefix 'test': \n", len(results))
	for _, r := range results {
		fmt.Printf(" %s - %s\n", r.Key, string(r.Value))
	}

	for _, r := range results {
		if !startsWithPrefix(r.Key, "test") {
			t.Errorf("Key %s does not start with prefix 'test'", r.Key)
		}
	}

	expected := []string{"test1", "test2", "test3", "test4"}
	if len(results) != len(expected) {
		t.Errorf("Expected %d results, got %d", len(expected), len(results))
	}
}

func BPlusTreeRangeScanTest(t *testing.T) {
	tree, _ := NewBPlusTree(3)
	keys := []string{"a", "b", "v", "g", "d", "dj", "e"}
	for _, key := range keys {
		tree.Insert(key, []byte("vrednost_"+key))
	}
	results := tree.RangeScan("b", "dj")
	for _, r := range results {
		fmt.Printf("%s\n", r.Key)
	}

	expected := []string{"b", "v", "g", "d", "dj"}
	if len(results) != len(expected) {
		t.Errorf("Expected %d results but got %d", len(expected), len(results))
	}
}
