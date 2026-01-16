package BufferPool

import (
	"container/list"
	"fmt"
	"log"
	"os"
	"strconv"
)

type BufferPool struct {
	maxSize     int               //maksimalna kolicina blokova koje BufferPool sadrzi
	currentSize int               //trenutna velicina buffer poola-a
	blockSize   int               //velicina bloka unutar buffer pool-a
	lruList     *list.List        //lista za lru algoritam
	cacheMap    map[string][]byte //mapa koja sadrzi parove filepath+blok:sadrzaj bloka
}

// konstruktor za buffer pool - validne vrednosti za block size su: 4, 8, 16KB, koje unutar konstruktora bivaju prebacene u bajtove
func NewBufferPool(maxSize int, blockSize int) (*BufferPool, error) {
	//u dokumentaciji pise da block size mogu biti 4, 8, ili 16-ako nisu, bacamo error
	if blockSize != 4 && blockSize != 8 && blockSize != 16 { //PREDLOG - umesto KB da stavimo u bajtovima, jer sa njima i radimo , to bi onda bilo 4096,8192,16384
		return nil, fmt.Errorf("block size must be either 4, 8, or 16")
	}
	//instanciramo cachemap
	cacheMap := make(map[string][]byte, maxSize)
	return &BufferPool{
		maxSize:     maxSize,
		currentSize: 0,
		blockSize:   blockSize * 1024, //posto nam u konfig fajlu stoji u KB, ovde prevodimo u bajtove
		lruList:     list.New(),
		cacheMap:    cacheMap,
	}, nil
}

// Funkcija koja dobavlja informacije zapisane u odredjenom bloku nekog fajla
func (bp *BufferPool) Get(filepath string, blockNumber int) (*[]byte, error) {
	//konkateniramo broj trazenog bloka na path fajla da bi dobili vrednost kljuca
	key := filepath + strconv.Itoa(blockNumber)
	//proveravamo da li se kljuc nalazi u mapi
	value, ok := bp.cacheMap[key]
	if !ok {
		file, err := os.OpenFile(filepath, os.O_RDONLY, 0644)
		if err != nil {
			log.Fatal(err)
		} else {
			//ako nismo nasli kljuc, moramo naci unutar fajla bajtove koji nam trebaju
			defer file.Close()
			//idemo do bajtova koji nam trebaju
			block := int64((blockNumber - 1) * bp.blockSize)
			file.Seek(block, 0)
			returnBytes := make([]byte, bp.blockSize)
			_, err := file.Read(returnBytes)
			if err != nil {
				return nil, err
			}
			//dodajemo bajtove u lru listu i mapu, i povecavamo current size
			bp.cacheMap[key] = returnBytes
			(*bp.lruList).PushFront(key)
			bp.currentSize += 1

			var foundElem *list.Element
			for e := bp.lruList.Front(); e != nil; e = e.Next() {
				if e.Value.(string) == key {
					foundElem = e
					break
				}
			}
			if foundElem != nil {
				bp.lruList.MoveToFront(foundElem)
			}
			return &returnBytes, nil
		}
	}
	//ako se kljuc nalazi u mapi, samo vracamo sta smo nasli
	return &value, nil
}

// Funkcija koja zapisuje podatke u blok fajla, i onda dodaje taj blok u BufferPool ako nije vec tu
func (bp *BufferPool) Put(filepath string, blockNumber int, writeValue *[]byte) error {
	//otvaramo fajl
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	//fajl ce se zatvoriti kad returnujemo
	defer file.Close()
	//idemo do bloka koji trazimo
	block := int64((blockNumber - 1) * bp.blockSize)

	_, err = file.Seek(block, 0)
	if err != nil {
		return err
	}
	_, err = file.Write(*writeValue)
	if err != nil {
		return err
	}
	//kad smo vec zapisali sta smo trebali, proveravamo da li je taj blok vec u bufferu
	key := filepath + strconv.Itoa(blockNumber)
	_, ok := bp.cacheMap[key]
	//ako blok nije u bufferu, dodajemo ga u lru listu, mapu, i povecavamo current size
	if !ok {
		bp.cacheMap[key] = *writeValue
		//TODO: Ispraviti!
		//Na prvi pogled, meni se čini kao da je problem ovde
		//Na početku rada, sama lru Lista je prazna
		//Ako blok nije nadjen, ovde se on dodaje u cachemap,
		//Ali onda pokusavamo da nadjemo foundElem sa vrednoscu kljuca na koji dodajemo blok
		//A mi zbog cinjenice da smo unutar !ok bloka znamo da taj kljuc nije u listi,
		//jer nam ok daj True kad je kljuc pronadjen, a False kada nije
		//Proveri i get!
		var foundElem *list.Element
		for e := bp.lruList.Front(); e != nil; e = e.Next() {
			if e.Value.(string) == key {
				foundElem = e
				break
			}
		}
		//Ovo nikad nece biti istina-found elem bi trebalo da uvek bude nil,
		//jer u cachemap uopste nismo pronasli kljuc koji trazimo!!!
		if foundElem != nil {
			bp.lruList.MoveToFront(foundElem)
		}
		return nil

	}
	if bp.currentSize >= bp.maxSize {
		bp.evictOldest()
	}
	//TODO: ovaj kod bi verovatno zapravo trebao da bude unutar !ok bloka
	//zato što ovaj kod zapravo dodaje ključ i vrednost u cachemap i lruList
	bp.cacheMap[key] = *writeValue
	bp.lruList.PushFront(key)
	bp.currentSize += 1
	return nil

}

// Glavna komponenta LRU algoritma
// Izbacuje najstariji (poslednji) element liste kada se lista popuni
func (bp *BufferPool) evictOldest() {
	if bp.lruList.Len() == 0 {
		return
	}
	oldest := bp.lruList.Back()
	if oldest != nil {
		key := oldest.Value.(string)
		delete(bp.cacheMap, key)
		bp.lruList.Remove(oldest)
		bp.currentSize--
	}
}

// funkcija koja brise kljuc iz mape i iz liste-podrazumeva se da je u pitanju poslednji element liste (lru algoritam)
func (bp *BufferPool) Delete(filepath string, blockNumber int) error {
	key := filepath + strconv.Itoa(blockNumber)
	_, ok := (bp.cacheMap)[key]
	if !ok {
		return fmt.Errorf("the specified file does not contain the specified block")
	} else {
		delete(bp.cacheMap, key)

		var foundElem *list.Element
		for e := bp.lruList.Front(); e != nil; e = e.Next() {
			if e.Value.(string) == key {
				foundElem = e
				break
			}
		}
		if foundElem != nil {
			bp.lruList.Remove(foundElem)
		}
		bp.currentSize--
		return nil
	}
}
func (bp *BufferPool) Size() int {
	return bp.currentSize
}
func (bp *BufferPool) GetBlockSize() int {
	return bp.blockSize
}
