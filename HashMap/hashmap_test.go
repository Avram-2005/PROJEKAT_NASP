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
