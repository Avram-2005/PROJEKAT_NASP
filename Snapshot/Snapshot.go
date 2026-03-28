package Snapshot

import (
	"container/list"
	"fmt"
)

// Snapshot je mapa koja na svakom kljucu sadrži listu verzija vrednosti na nekom ključu
// ovaj struct zapravo predstavlja mapu koja skladisti snapshot-ove raznih kljuceva baze
type Snapshot struct {
	snapshotMap map[string]*list.List
}

// Konstruktor ne zahteva ikakve parametre-snapshot objekat stvaramo pri pokretanju,
// pa ga punimo sa kljucevima koje korisnik zeli da prati
// TODO: proveriti da li mozda trebaju neki parametri
func NewSnapshot() (*Snapshot, error) {
	newMap := make(map[string]*list.List)
	return &Snapshot{
		snapshotMap: newMap,
	}, nil
}

// Funkcija koja ili stvara novi snapshot za odredjen kljuc,
// ili dodaje novu vrednost u postojeci
// key-kljuc pod koji dodajemo
// value-niz bajtova, najnovija vrednost
func (sp *Snapshot) Add(key string, value *[]byte) error {
	foundList, ok := sp.snapshotMap[key]
	// ako kljuc nije pronadjen, stvaramo nov snapshot
	if !ok {
		newList := list.New()
		newList.PushFront(value)
		sp.snapshotMap[key] = newList
		return nil
	}
	// ako kljuc vec postoji, dodajemo novu vrednost na kraj njegovog snapshot-a
	foundList.PushFront(value)
	return nil
}

// Funkcija koja dobavlja odredjenu verziju naseg podatka
// key-kljuc koji se trazi
// version-verzija kljuca koju trazimo, gde je prva dodata vrednost nulta verzija,
// i svaka naredna je veca za jedan
func (sp *Snapshot) Get(key string, version int) (*[]byte, error) {
	foundList, ok := sp.snapshotMap[key]
	//error ako kljuc koji trazimo ne postoji
	if !ok || foundList.Len() == 0 {
		return nil, fmt.Errorf("key not found in snapshot")
	}
	counter := 0
	for elem := foundList.Back(); elem != nil; elem = elem.Prev() {
		// da bi pronasli odgovarajucu verziju koristimo brojac i petlju
		if counter == version {
			return elem.Value.(*[]byte), nil
		}
		counter += 1
	}
	return nil, fmt.Errorf("version number not found for the specified key")
}

// Funkcija koja dobavlja nultu verziju nase vrednosti
// key-kljuc pod kojim se trazi
func (sp *Snapshot) GetFirst(key string) (*[]byte, error) {
	foundList, ok := sp.snapshotMap[key]
	if !ok || foundList.Len() == 0 {
		return nil, fmt.Errorf("key not found in snapshot")
	}
	return foundList.Back().Value.(*[]byte), nil
}

// Funkcija dobavlja poslednju verziju nase vrednosti
// key-kljuc pod kojim trazimo
func (sp *Snapshot) GetLatest(key string) (*[]byte, error) {
	foundList, ok := sp.snapshotMap[key]
	if !ok || foundList.Len() == 0 {
		return nil, fmt.Errorf("key not found in snapshot")
	}
	return foundList.Front().Value.(*[]byte), nil
}

// Funkcija dobavlja broj razlicitih verzija koje se skladiste za odredjen kljuc
func (sp *Snapshot) GetVersionCount(key string) (int, error) {
	foundList, ok := sp.snapshotMap[key]
	if !ok {
		return 0, fmt.Errorf("key is not tracked in this snapshot")
	}
	return foundList.Len(), nil
}

// Izbacuje snapshot za specifican kljuc iz radne memorije
func (sp *Snapshot) Free(key string) error {
	_, ok := sp.snapshotMap[key]
	if !ok {
		return fmt.Errorf("cannot free the memory of a nonexistent snapshot")
	}
	delete(sp.snapshotMap, key)
	return nil
}
