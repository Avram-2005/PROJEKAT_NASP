package SkipList

import (
	"errors"
	"fmt"
	"math/rand"
)

type node struct {
	key   string
	value []byte
	next  []*node //niz pokazivaca koji pokazuju na svaki nivo
}

type SkipList struct {
	maxHeight int   //maksimalna visina
	height    int   //vsiina na kojoj smo trenutno
	head      *node //pocetni cvor - minus beskonacno
	size      int   //br elemenata
}

// kreiranje nove SkipListe
func NewSkipList(maxHeight int) (*SkipList, error) {
	if maxHeight <= 0 {
		return nil, errors.New("Maximum height must be positive")
	}
	return &SkipList{
		maxHeight: maxHeight,
		height:    1,
		head: &node{
			key:  "",
			next: make([]*node, maxHeight),
		},
		size: 0,
	}, nil
}

// Novcic funkcija odredjuje 0 ili 1 za dodavanje elemenata skipliste
// (pomocna funkcija dobijena na vezbama)
func (s *SkipList) randomHeight() (int, error) {
	level := 1
	// moguce vrednosti koje vraca rand su 0 i 1
	// zaustavljamo se kad dobijemo 0
	for ; rand.Intn(2) == 1; level++ {
		if level >= s.maxHeight {
			return s.maxHeight, nil //vraca maksimalnu velicinu, a ne gresku
		}
	}
	if level > s.maxHeight {
		level = s.maxHeight
	}
	return level, nil
}

// Dodavanje novog elementa u skip listu
func (skipList *SkipList) Put(key string, value []byte) error {
	if key == "" {
		return errors.New("Key cannot be empty")
	}
	if value == nil {
		return errors.New("Value cannot be nil")
	}
	//priprema niza
	//  u njega pamtimo poslednji cvor pre kljuca na svakom nivou, pa znamo da tu ubacimo novi element
	prev := make([]*node, skipList.maxHeight)
	current := skipList.head //pretraga krece od minus beskonacno
	//trazimo najnizi nivo, kako bismo dodali element
	for i := skipList.height - 1; i >= 0; i-- {
		//idemo ka desno dok next nije nil ili kljuc sledeceg cvora je manji od naseg
		for current.next[i] != nil && current.next[i].key < key {
			current = current.next[i] //prelazak na sledeci cvor
		}
		prev[i] = current //poslednji cvor ciji je kljuc manji od naseg cuva se u niz
	}

	current = current.next[0] //nivo 0, u current je ili nas kljuc ili prvi cvor sa vecim
	//provera postojanja kljuca
	if current != nil && current.key == key {
		current.value = value
		return nil
	}
	//kljuc nije postojao, pa pravimo novi
	newHeight, err := skipList.randomHeight() //bacanje novcica
	if err != nil {
		return fmt.Errorf("Failed to generate random height: %w", err)
	}

	if newHeight > skipList.height {
		for i := skipList.height; i < newHeight; i++ {
			prev[i] = skipList.head
		}
		skipList.height = newHeight
	}
	newNode := &node{ //kreiranje novog cvora
		key:   key,
		value: value,
		next:  make([]*node, newHeight),
	}

	for i := 0; i < newHeight; i++ { //dodajemo novi cvor na sve nivoe do broja puta koliko je novcic pokazao na 1
		newNode.next[i] = prev[i].next[i] //pokazivac novog cvora pokazuje na ono sto je prev pokazivao
		prev[i].next[i] = newNode         //prev pokazuje na novi cvor
	}
	skipList.size++
	return nil
}

// Brisanje elementa iz skipliste
func (skipList *SkipList) Delete(key string) error {
	if key == "" {
		return errors.New("Key cannot be empty")
	}
	prev := make([]*node, skipList.maxHeight)   //u prev ce se naci poslednji cvor pre kljuca koji brisemo
	current := skipList.head                    //krecemo od minus beskonacno
	for i := skipList.height - 1; i >= 0; i-- { //trazimo cvor koji zelimo da obrisemo
		for current.next[i] != nil && current.next[i].key < key {
			current = current.next[i]
		}
		prev[i] = current //sacuvamo cvor pre kljuca koji smo trazili
	}
	current = current.next[0] //ili cvor koji brisemo ili prvi sa vecim kljucem
	if current == nil || current.key != key {
		return errors.New("Key not found") //element nije pronadjen
	}

	for i := 0; i < skipList.height; i++ { //brisanje elementa sa svih nivoa
		if prev[i].next[i] != current {
			break // izlazimo iz petlje, jer current nije na ovom nivou, pa ni na vsiim od njega
		}
		prev[i].next[i] = current.next[i] //prethodni clan ce pokazuje na ono sto je pokazivao obrisan element
	}

	//proveravamo da li imamo prazne nivoe
	for skipList.height > 1 && skipList.head.next[skipList.height-1] == nil {
		skipList.height-- //ako pokazivac minus beskonacno pokazuja na none (tj plus beskonacno)
		//znaci da je nivo prazan pa ga uklanjamo
	}
	skipList.size--
	return nil //element je uspesno obrisan
}

// Pretraga elemenata po kljucu
// Vraca par (vrednost, bool)
func (skipList *SkipList) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, errors.New("Key cannot be empty")
	}
	current := skipList.head
	for i := skipList.height - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].key < key { //krecemo se desno dok nenadjemo cvor sa vecim ili jednakim kljucem
			current = current.next[i]
		}
	}
	current = current.next[0]                 //nulti nivo, gledamo cvor na koji pokazuje current
	if current == nil || current.key != key { //proveravamo da li je pretraga uspela
		return nil, errors.New("Key not found") //element nije pronadjen
	}
	return current.value, nil //element je nadjen
}
