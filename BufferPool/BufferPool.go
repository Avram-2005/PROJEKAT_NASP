package bufferpool

import (
	"container/list"
	"fmt"
	"log"
	"os"
	"strconv"
)

type bufferPool struct {
	maxSize     int                //maksimalna kolicina blokova koje bufferPool sadrzi
	currentSize int                //trenutna velicina buffer poola-a
	blockSize   int                //velicina bloka unutar buffer pool-a
	lruList     *list.List         //lista za lru algoritam
	cacheMap    *map[string][]byte //mapa koja sadrzi parove filepath+blok:zadrzaj bloka
}

// konstruktor za buffer pool validne vrednosti za block size su: 4, 8, 16
func NewBufferPool(maxSize int, blockSize int) (*bufferPool, error) {
	//u dokumentaciji pise da block size mogu biti 4, 8, ili 16-ako nisu, bacamo error
	if blockSize != 4 && blockSize != 8 && blockSize != 16 {
		return nil, fmt.Errorf("Block size must be either 4, 8, or 16")
	}
	//instanciramo cachemap
	cacheMap := make(map[string][]byte, maxSize)
	return &bufferPool{
		maxSize:     maxSize,
		currentSize: 0,
		blockSize:   blockSize,
		lruList:     list.New(),
		cacheMap:    &cacheMap,
	}, nil
}

// Funkcija koja dobavlja informacije zapisane u odredjenom bloku nekog fajla
func (bp *bufferPool) Get(filepath string, blockNumber int) (*[]byte, error) {
	//konkateniramo broj trazenog bloka na path fajla da bi dobili vrednost kljuca
	key := filepath + strconv.Itoa(blockNumber)
	//proveravamo da li se kljuc nalazi u mapi
	value, ok := (*bp.cacheMap)[key]
	if !ok {
		file, err := os.Open(filepath)
		if err != nil {
			log.Fatal(err)
		} else {
			//ako nismo nasli kljuc, moramo naci unutar fajla bajtove koji nam trebaju
			defer file.Close()
			//idemo do bajtova koji nam trebaju
			file.Seek(0, blockNumber*bp.blockSize)
			returnBytes := make([]byte, bp.blockSize)
			_, err := file.Read(returnBytes)
			if err != nil {
				return nil, fmt.Errorf("Error while reading file")
			}
			//dodajemo bajtove u lru listu i mapu, i povecavamo current size
			(*bp.cacheMap)[key] = returnBytes
			(*bp.lruList).PushFront(key)
			bp.currentSize += 1
			//TODO1: Ako je current size veci nego max size, pokrenuti lru algoritam!
			return &returnBytes, nil
		}
	}
	//ako se kljuc nalazi u mapi, samo vracamo sta smo nasli
	return &value, nil
}

// Funkcija koja zapisuje podatke u blok fajla, i onda dodaje taj blok u bufferpool ako nije vec tu
func (bp *bufferPool) Put(filepath string, blockNumber int, writeValue *[]byte) error {
	//otvaramo fajl
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	//fajl ce se zatvoriti kad returnujemo
	defer file.Close()
	//idemo do bloka koji trazimo
	file.Seek(0, blockNumber*bp.blockSize)

	_, err = file.Write(*writeValue)
	if err != nil {
		return fmt.Errorf("Error while reading file")
	}
	//kad smo vec zapisali sta smo trebali, proveravamo da li je taj blok vec u bufferu
	key := filepath + strconv.Itoa(blockNumber)
	_, ok := (*bp.cacheMap)[key]
	//ako blok nije u bufferu, dodajwemo ga u lru listu, mapu, i povecavamo current size
	if !ok {
		(*bp.cacheMap)[key] = *writeValue
		(*bp.lruList).PushFront(key)
		bp.currentSize += 1
		//TODO2: Implementirati da se pokrene lru algoritam kad se prekoraci velicina buffera
	}
	return nil

}

//TODO3: IMPLEMENTIRATI SAM LRU ALGORITAM

// funkcija koja brise kljuc iz mape i iz liste-podrazumeva se da je u pitanju poslednji element liste
func (bp *bufferPool) Delete(filepath string, blockNumber int) error {
	key := filepath + strconv.Itoa(blockNumber)
	_, ok := (*bp.cacheMap)[key]
	if !ok {
		return fmt.Errorf("The specified file does not contain the specified block")
	} else {
		delete((*bp.cacheMap), key)
		//TODO4: Naleteo sam na problem, a to je cinjenica da *list.Element ne specificira sa kojim tipom radi.
		//s time sto ova funkcija zapravo samo brise poslednji element, mogli bi smo da samo iz mape obrisemo
		//element sa kraja liste[delete((*bp.cacheMap),(*bp.lruList).Back().Value)], ali mapa ne prepoznaje da
		//je taj element tipa string-ako se ovo nekako prepravi delete se moze pozvati bez parametara,
		//i zahtevati mnogo manje koda.
		(*bp.lruList).Remove((*bp.lruList).Back())
		return nil
	}
}
