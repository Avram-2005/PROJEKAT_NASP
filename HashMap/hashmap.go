package HashMap

import (
	"errors"
	"sort"
)

type HashMap struct {
	data map[string][]byte //parovi kljuc-vrednost
	size int               //broj elemenata mape
}

// pravljenje nove prazne hashmape
func NewHashMap() *HashMap {
	return &HashMap{
		data: make(map[string][]byte),
		size: 0,
	}
}

// dodavanje novog elementa
func (hashmap *HashMap) Put(key string, value []byte) error {
	if key == "" {
		return errors.New("Key cannot be empty")
	}
	if value == nil {
		return errors.New("Value cannot be nil")
	}
	_, exists := hashmap.data[key] //provera da li kljuc vec postoji
	hashmap.data[key] = value      //dodamo vrednost
	if !exists {
		hashmap.size++ //brojac povecavamo samo ukoliko dodajemo element, ne i prilikom azuriranja
	}
	return nil
}

// pretraga elemenata po kljucu
// povratna vrednost je par vrednost,bool
func (hashmap *HashMap) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, errors.New("Key cannot be empty")
	}
	value, found := hashmap.data[key] //trazimo vrednost u mapi, ako smo nasli bice true,ako ne onda false
	if !found {
		return nil, errors.New("Key not found") //nismo nasli element
	}
	return value, nil
}

// brisanje elementa iz mape
func (hashmap *HashMap) Delete(key string) error {
	if key == "" {
		return errors.New("Key cannot be empty")
	}
	_, exists := hashmap.data[key] //trazimo element
	if exists {
		delete(hashmap.data, key)
		hashmap.size--
		return nil //element je uspesno pronadjen i obrisan
	}
	return errors.New("Key not found") //element nije pronadjen
}

// Povratna vrednost - broj elemenata mape
func (hashmap *HashMap) Size() int {
	return hashmap.size
}

// proverava da li je mapa prazna
// Povratna vrednost-boolean
func (hashmap *HashMap) IsEmpty() bool {
	return hashmap.size == 0
}

// Proverava postojanje kljuca u mapi
// Povratna vrednost boolean, true ako postoji
func (hashmap *HashMap) Contains(key string) bool {
	_, exists := hashmap.data[key] //trazimo element
	return exists                  //vraca true ako element postoji, a false ako ne postoji
}

// Vraca listu svih kljuceva u mapi
func (hashmap *HashMap) Keys() []string {
	keys := make([]string, 0, hashmap.size) //inicijalizacija liste u koju cemo da smestamo kljuceve,za pocetak postavili na 0
	for key := range hashmap.data {
		keys = append(keys, key)
	}
	return keys
}

// Vraca listu svih vrednosti u mapi
func (hashmap *HashMap) Values() [][]byte {
	values := make([][]byte, 0, hashmap.size) //inicijalizacija liste u koju smestamo sve vrednosti,za pocetak postavili na 0
	for _, value := range hashmap.data {
		values = append(values, value)
	}
	return values
}

// Vraca sve parove
func (hashmap *HashMap) Items() []struct {
	Key   string
	Value []byte
} {
	pairs := make([]struct {
		Key   string
		Value []byte
	}, 0, hashmap.size) //iniciijalizacija liste u koju smestamo parove, pocetna vrednost 0
	for key, value := range hashmap.data { //iteriramo kroz mapu
		pairs = append(pairs, struct { //za svaki par pravimo novu strukturu
			Key   string
			Value []byte
		}{Key: key, Value: value})
	}
	return pairs
}

// Vraca sortirane kljuceve za flush na disk
func (hashmap *HashMap) GetSortedEntries() []struct {
	Key   string
	Value []byte
} {
	keys := make([]string, 0, hashmap.size)
	for key := range hashmap.data {
		keys = append(keys, key)
	}
	//sortiranje kljuceva
	sort.Strings(keys)

	//pravljenje sortiranu listu parova
	pairs := make([]struct {
		Key   string
		Value []byte
	}, 0, hashmap.size)

	for _, key := range keys {
		pairs = append(pairs, struct {
			Key   string
			Value []byte
		}{
			Key:   key,
			Value: hashmap.data[key],
		})
	}
	return pairs
}

// vraca sortirane unose u opsegu
// ulazni parametri - pocetni i krajnji kljjuc (string)
func (hashmap *HashMap) RangeScan(startKey, endKey string) []struct {
	Key   string
	Value []byte
} {
	entries := hashmap.GetSortedEntries()
	result := make([]struct {
		Key   string
		Value []byte
	}, 0)

	for _, entry := range entries {
		if entry.Key >= startKey && entry.Key <= endKey {
			result = append(result, entry)
		} else if entry.Key > endKey {
			break //kljucevi su sortirani
		}
	}
	return result
}

// vraca sortirane unose po prefiksu
// ulazni parametar je prefix (string)
func (hashmap *HashMap) PrefixScan(prefix string) []struct {
	Key   string
	Value []byte
} {
	entries := hashmap.GetSortedEntries()
	result := make([]struct {
		Key   string
		Value []byte
	}, 0)

	for _, entry := range entries {
		if len(entry.Key) >= len(prefix) && entry.Key[:len(prefix)] == prefix {
			result = append(result, entry)
		} else if entry.Key > prefix && !startsWith(entry.Key, prefix) {
			break //stigli do kraja
		}
	}
	return result
}

func startsWith(str, prefix string) bool {
	if len(str) < len(prefix) {
		return false
	}
	return str[:len(prefix)] == prefix
}
