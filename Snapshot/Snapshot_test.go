package Snapshot

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

func TestBufferPool(t *testing.T) {
	sp, err := NewSnapshot()
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom incijalizacije snapshot-a")
		t.FailNow()
	}
	//Dodavanje svih vrednosti za testiranje
	value1 := make([]byte, 8)
	binary.BigEndian.PutUint16(value1, 16)
	value2 := make([]byte, 8)
	binary.BigEndian.PutUint16(value2, 12)
	value3 := make([]byte, 8)
	binary.BigEndian.PutUint16(value3, 56)
	value4 := make([]byte, 8)
	binary.BigEndian.PutUint16(value4, 64)
	fmt.Print(value1, "\n")
	fmt.Print(value2, "\n")
	fmt.Print(value3, "\n")
	fmt.Print(value4, "\n")
	sp.Add("key1", &value1)
	sp.Add("key1", &value2)
	sp.Add("key1", &value3)
	sp.Add("key2", &value4)
	// Provera broja verzija posle dodavanja
	number, err := sp.GetVersionCount("key1")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja broja verzija prvog kljuca")
		t.FailNow()
	}
	if number != 3 {
		fmt.Print(err)
		t.Errorf("pogresan broj verzija prvog kljuca")
		t.FailNow()
	}
	number, err = sp.GetVersionCount("key2")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja broja verzija drugog kljuca")
		t.FailNow()
	}
	if number != 1 {
		fmt.Print(err)
		t.Errorf("pogresan broj verzija drugog kljuca")
		t.FailNow()
	}
	//proveravanje metode getlatest
	compare1, err := sp.GetLatest("key1")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja poslednje verzije prvog kljuca")
		t.FailNow()
	}
	if !reflect.DeepEqual(value3, (*compare1)) {
		fmt.Print(value3, (*compare1))
		t.Errorf("pogresna poslednja verzija prvog kljuca")
		t.FailNow()
	}
	compare2, err := sp.GetLatest("key2")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja poslednje verzije drugog kljuca")
		t.FailNow()
	}
	if !reflect.DeepEqual(value4, (*compare2)) {
		fmt.Print(value4, (*compare2))
		t.Errorf("pogresna poslednja verzija drugog kljuca")
		t.FailNow()
	}
	//Proveravanje metode getfirst
	compare1, err = sp.GetFirst("key1")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja prve verzije prvog kljuca")
		t.FailNow()
	}
	if !reflect.DeepEqual(value1, (*compare1)) {
		fmt.Print(value1, (*compare1))
		t.Errorf("pogresna poslednja verzija prvog kljuca")
		t.FailNow()
	}
	compare2, err = sp.GetFirst("key2")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja prve verzije drugog kljuca")
		t.FailNow()
	}
	if !reflect.DeepEqual(value4, (*compare2)) {
		fmt.Print(value4, (*compare2))
		t.Errorf("pogresna prva verzija drugog kljuca")
		t.FailNow()
	}

	compare3, err := sp.Get("key1", 1)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja druge verzije prvog kljuca")
		t.FailNow()
	}
	if !reflect.DeepEqual(value2, (*compare3)) {
		fmt.Print(value2, (*compare3))
		t.Errorf("pogresna druga verzija prvog kljuca")
		t.FailNow()
	}

	err = sp.Free("key1")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom oslobadjanja kljuca")
		t.FailNow()
	}
	// posle oslobadjanja memorije za key1, ocekujemo error ako pokusamo da ga nadjemo
	_, err = sp.GetVersionCount("key1")
	if err == nil {
		fmt.Print(err)
		t.Errorf("treba da se prijavi greska, ali nije prijavljena")
		t.FailNow()
	}
}
