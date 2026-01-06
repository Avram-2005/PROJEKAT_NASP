package SkipList

import (
	"fmt"
	"testing"
)

func TestSkipList(t *testing.T) {
	//pravljenje skipliste
	skipl, err := NewSkipList(4)
	if err != nil {
		t.Fatalf("Failed to create a SkipList: %v", err)
	}
	fmt.Println("SkipList was successfully created")

	//dodavanje elemenata
	err = skipl.Put("abc", []byte("prvaVrednost"))
	if err != nil {
		t.Errorf("Failed to put 'abc' into the skiplist: %v", err)
	}
	err = skipl.Put("def", []byte("drugaVrednost"))
	if err != nil {
		t.Errorf("Failed to put 'def' into the skiplist: %v", err)
	}
	err = skipl.Put("ghi", []byte("trecaVrednost"))
	if err != nil {
		t.Errorf("Failed to put 'ghi' into the skiplist: %v", err)
	}
	fmt.Println("Three elements added")

	//dobavljanje postojeceg elementa
	value, err := skipl.Get("abc")
	if err != nil {
		t.Errorf("Element 'abc' was not found: %v", err)

	} else {
		fmt.Printf("Element 'abc' was found: %s\n", string(value))
	}

	//dobavljanje nepostojeceg elementa
	_, err = skipl.Get("jkl")
	if err != nil {
		fmt.Printf("Element 'jkl' was not found: %v\n", err)

	} else {
		t.Error("Element 'jkl' was found")
	}

	//azuriranje postojeceg elementa
	err = skipl.Put("abc", []byte("cetvrtaVrednost"))
	if err != nil {
		t.Errorf("Failed to update element 'abc' : %v", err)
	}
	value, err = skipl.Get("abc")
	if err != nil {
		t.Errorf("Failed to update 'abc': %v", err)
	} else if string(value) == "cetvrtaVrednost" {
		fmt.Println("Element 'abc' was successfully updated.")
	}

	//brisanje elementa
	err = skipl.Delete("def")
	if err != nil {
		t.Error("Failed to delete 'def'")
	} else {
		fmt.Println("Element 'def' was deelted")
	}

	//provera uspesnosti brisanja
	_, err = skipl.Get("def")
	if err != nil {
		fmt.Printf("Element 'def' was not found after deletion: %v\n", err)
	} else {
		t.Error("Element 'def' was found even though it was deleted")
	}

	//brisanje nepostojeceg elementa
	err = skipl.Delete("nepostojec")
	if err != nil {
		fmt.Printf("An attempt at deleting a nonexistent element returns an error: %v/n", err)

	} else {
		t.Error("Deletion of a nonexistent element should have returned an error")
	}

	//dodavanje praznog kljuca
	err = skipl.Put("", []byte("vrednost"))
	if err != nil {
		fmt.Printf("Eveerything works : %v\n", err)
	} else {
		t.Error("An error message should have been returned")
	}
	//dodavanje nil vrednosti
	err = skipl.Put("test", nil)
	if err != nil {
		fmt.Printf("Everything works: %v\n", err)
	} else {
		t.Error("Nil value should have triggered an error")
	}
}
