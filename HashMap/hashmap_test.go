package HashMap

import (
	"fmt"
	"testing"
)

func TestHashMap(t *testing.T) {
	//kreiranje nove hashmape
	hm := NewHashMap()
	if hm == nil {
		t.Fatal("Failed to create a HashMap")
	}
	fmt.Println("HashMap was successfully created")

	//dodavanje elemenata
	err := hm.Put("kljuc1", []byte("vrednost1"))
	if err != nil {
		t.Errorf("Put action failed: %v", err)
	}
	err = hm.Put("kljuc2", []byte("vrednost2"))
	if err != nil {
		t.Errorf("Put action failed: %v", err)
	}
	fmt.Println("Two elements were added to the HashMap")

	//dobavljanje postojeceg elementa
	value, err := hm.Get("kljuc1")
	if err != nil {
		t.Errorf("Get action failed: %v", err)
	} else if string(value) != "vrednost1" {
		t.Errorf("Expected 'vrednost1', got '%s'", string(value))
	} else {
		fmt.Println("Element 'kljuc1' was found")
	}

	//dobavljanje nepostojeceg elementa
	_, err = hm.Get("kljuc3")
	if err == nil {
		t.Error("Expected error for a nonexistent key")
	} else {
		fmt.Printf("The nonexistent key raises an error: %v\n", err)
	}

	//dobavljanje sa praznim kljucem
	_, err = hm.Get("")
	if err == nil {
		t.Error("Expected an error for empty key")
	} else {
		fmt.Printf("An empty key returns the following error: %v\n", err)
	}

	//azuriranje postojeceg elementa
	err = hm.Put("kljuc1", []byte("novaVrednost"))
	if err != nil {
		t.Errorf("Put action failed: %v", err)
	}
	value, err = hm.Get("kljuc1")
	if err != nil {
		t.Errorf("Get action failed after update: %v", err)

	} else if string(value) != "novaVrednost" {
		t.Errorf("Expected 'novaVrednost' but got '%s'", string(value))
	} else {
		fmt.Println("Element 'kljuc1' got updated")
	}

	//dodavanje sa nil vrednoscu
	err = hm.Put("kljuc4", nil)
	if err == nil {
		t.Error("Expected error for a nil value")
	} else {
		fmt.Printf("Nil value returns an error: %v\n", err)
	}

	//brisanje elemenata
	err = hm.Delete("kljuc2")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	} else {
		fmt.Println("Element 'kljuc2' was deleted")
	}

	//brisanje nepostojeceg elemenata
	err = hm.Delete("kljuc7")
	if err == nil {
		t.Errorf("Expected an error for deleting a nonexistent key")
	} else {
		fmt.Printf("Deletion of a nonexistent element returns an error: %v\n", err)
	}

	//contains
	if !hm.Contains("kljuc1") {
		t.Error("Contains should return true for an existing key")
	} else {
		fmt.Println("Contains returns true for an existing key")
	}

	if hm.Contains("kljuc18") {
		t.Error("Contains should return false for a nonexistent key")
	} else {
		fmt.Println("Contains returns false for a nonexistent key")
	}
}

func TestSortingFunctions(t *testing.T) {
	hm := NewHashMap()
	hm.Put("zebra", []byte("z"))
	hm.Put("pas", []byte("p"))
	hm.Put("laptop", []byte("l"))
	hm.Put("mis", []byte("m"))

	sorted := hm.GetSortedEntries()
	fmt.Println("Sorted entries: ")
	for _, e := range sorted {
		fmt.Printf("%s\n", e.Key)
	}

	if sorted[0].Key != "laptop" || sorted[1].Key != "mis" || sorted[2].Key != "pas" || sorted[3].Key != "zebra" {
		t.Error("Not sorted correctly")
	}

	//RangeScan
	rangesc := hm.RangeScan("mis", "zebra")
	fmt.Println("RangeScan: ")
	for _, e := range rangesc {
		fmt.Printf("%s\n", e.Key)
	}
	if len(rangesc) != 3 || rangesc[0].Key != "mis" || rangesc[2].Key != "zebra" {
		t.Error("RangeScan is not working correctly")
	}
	//PrefixScan
	prefixsc := hm.PrefixScan("la")
	fmt.Printf("PrefixScan: ")
	for _, e := range prefixsc {
		fmt.Printf("%s\n", e.Key)
	}
	if len(prefixsc) != 1 || prefixsc[0].Key != "laptop" {
		t.Error("PrefixScan is not working correctly")
	}
}
