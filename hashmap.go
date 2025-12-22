package projekatnasp

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
func (hashmap *HashMap) Put(key string, value []byte) {
	_, exists := hashmap.data[key] //provera da li kljuc vec postoji
	hashmap.data[key] = value      //dodamo vrednost
	if !exists {
		hashmap.size++ //brojac povecavamo samo ukoliko dodajemo element, ne i prilikom azuriranja
	}
}

// pretraga elemenata po kljucu
// povratna vrednost je par vrednost,bool
func (hashmap *HashMap) Get(key string) ([]byte, bool) {
	value, found := hashmap.data[key] //trazimo vrednost u mapi, ako smo nasli bice true,ako ne onda false
	if found {
		return value, true
	}
	return nil, false //nismo nasli element
}

// brisanje elementa iz mape
// povratna vrednost je boolean, true-pronadjen i obrisan, false-nije pronadjen
func (hashmap *HashMap) Delete(key string) bool {
	_, exists := hashmap.data[key] //trazimo element
	if exists {
		delete(hashmap.data, key)
		hashmap.size--
		return true //element je uspesno pronadjen i obrisan
	}
	return false //element nije pronadjen
}

// Povratna vrednost - broj elemenata mape
func (hashmap *HashMap) Size() int {
	return hashmap.size
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
