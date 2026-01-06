package skiplist

import (
	"fmt"
	"testing"
)

func TestSkipList(t *testing.T) {
	skipl := NewSkipList(4)
	fmt.Println("SkipLista je kreirana")
	skipl.Put("abc", []byte("prvaVrednost"))
	skipl.Put("def", []byte("drugaVrednost"))
	skipl.Put("ghi", []byte("trecaVrednost"))

	fmt.Println("Dodata tri elementa")

	if _, found := skipl.Get("abc"); found {
		fmt.Println("Element 'abc' je pronadjen")

	} else {
		t.Error("Element 'abc' nije pronadjen")
	}

	if _, found := skipl.Get("jkl"); found {
		fmt.Println("Element 'jkl' je pronadjen")

	} else {
		t.Error("Element 'jkl' nije pronadjen")
	}

	skipl.Put("abc", []byte("cetvrtaVrednost"))
	if value, found := skipl.Get("abc"); found {
		if string(value) == "cetvrtaVrednost" {
			fmt.Println("Vrednost kljuca 'abc' je azurirana")
		} else {
			fmt.Println("Vrednost kljuca 'abc' nije azurirana")
		}
	}

	if skipl.Delete("def") {
		fmt.Println("Kljuc 'def' je obrisan")
	} else {
		t.Error("Kljuc 'def' nije obrisan")
	}

	if _, found := skipl.Get("def"); found {
		fmt.Println("Element 'def' je pronadjen")

	} else {
		t.Error("Element 'def' nije pronadjen")
	}
}
