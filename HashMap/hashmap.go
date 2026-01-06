package HashMap

import "errors"

type hashMap struct {
	data map[string][]byte //parovi kljuc-vrednost
	size int               //broj elemenata mape
}

// pravljenje nove prazne hashmape
func NewHashMap() *hashMap {
	return &hashMap{
		data: make(map[string][]byte),
		size: 0,
	}
}

// dodavanje novog elementa
func (hashmap *hashMap) Put(key string, value []byte) error {
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
func (hashmap *hashMap) Get(key string) ([]byte, error) {
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
func (hashmap *hashMap) Delete(key string) error {
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
func (hashmap *hashMap) Size() int {
	return hashmap.size
}

//proverava da li je mapa prazna
//Povratna vrednost-boolean
func (hashmap *hashMap) IsEmpty() bool {
	return hashmap.size == 0
}

// Proverava postojanje kljuca u mapi
// Povratna vrednost boolean, true ako postoji
func (hashmap *hashMap) Contains(key string) bool {
	_, exists := hashmap.data[key] //trazimo element
	return exists                  //vraca true ako element postoji, a false ako ne postoji
}

// Vraca listu svih kljuceva u mapi
func (hashmap *hashMap) Keys() []string {
	keys := make([]string, 0, hashmap.size) //inicijalizacija liste u koju cemo da smestamo kljuceve,za pocetak postavili na 0
	for key := range hashmap.data {
		keys = append(keys, key)
	}
	return keys
}

// Vraca listu svih vrednosti u mapi
func (hashmap *hashMap) Values() [][]byte {
	values := make([][]byte, 0, hashmap.size) //inicijalizacija liste u koju smestamo sve vrednosti,za pocetak postavili na 0
	for _, value := range hashmap.data {
		values = append(values, value)
	}
	return values
}

// Vraca sve parove
func (hashmap *hashMap) Items() []struct {
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
