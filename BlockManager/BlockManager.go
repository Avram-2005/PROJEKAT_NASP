package BlockManager

import (
	"fmt"
	"io"

	BufferPool "github.com/Avram-2005/PROJEKAT_NASP/BufferPool"
)

// Trenutno, BlockManager sluzi manje vise kao interfejs sa BufferPool-om
// Manje vise je tu samo da bi sakrio kod BufferPool-a, da vi ne morate oko toga da se brinete
// Najverovatnije cu kasnije uraditi refactor koji ce neke funkcionalnosti premestiti u BlockManager,
// ali to nece uopste uticati na sam api.
type BlockManager struct {
	blockCache *BufferPool.BufferPool
}

// Max size je maksimalna velicina bufferpoola sa kojim block manager raspolaze,
// a blockSize velicina blokova sa kojima radimo, a u dokumentaciji pise da su dozvoljene vrednosti 4, 8 ili 16KB.
//
// Iz tog razloga, jedine dozvoljene vrednosti za blockSize su 4, 8 i 16, ako stavite ista drugo dobicete error.
//
// Konverzija iz kilobajta u bajtove desava se unutar konstruktora za BufferPool, vi se oko toga ne brinete.
func NewBlockManager(maxSize int, blockSize int) (*BlockManager, error) {
	if blockSize != 4 && blockSize != 8 && blockSize != 16 {
		return nil, fmt.Errorf("dozvoljene vrednosti za blockSize su 4, 8, ili 16")
	}
	cache, err := BufferPool.NewBufferPool(maxSize, blockSize)
	if err != nil {
		return nil, fmt.Errorf("error pri inicijalizaciji block cache-a")
	}
	return &BlockManager{
		blockCache: cache,
	}, nil
}

// filepath-path do fajla iz kojeg se dobavljaju informacije
// blockNumber-broj bloka koji se trazi
//
// funkcija vraca niz bajtova, i error u slucaju da je nesto poslo po zlu
//
// BITNO: funkcija vam vraca CEO SADRZAJ BLOKA, na vama je da nadjete sta vam treba unutar njega
func (bm *BlockManager) Get(filepath string, blockNumber int) (*[]byte, error) {
	valueFound, err := bm.blockCache.Get(filepath, blockNumber)
	if err != nil {
		return nil, err
	}
	return valueFound, err
}

// filepath-path do fajla iz kojeg se dobavljaju informacije
// blockNumber-broj bloka koji se trazi
// offset-offset od pocetka bloka
// size-kolicina bajtova koja treba da se vrati
//
// na primer, sa offsetom 3, i size 3, vratili bi se cetvrti, peti i sesti bajt bloka
//
// funkcija vraca niz bajtova, i error u slucaju da je nesto poslo po zlu
//
// BITNO:funkcija vraca error za negativan offset, i ako se zbog offseta i size-a izadje van
// opsega podataka koji su trenutno zapisani na tom bloku
func (bm *BlockManager) GetSpecific(filepath string, blockNumber int, offset int, size int) (*[]byte, error) {
	valueFound, err := bm.blockCache.Get(filepath, blockNumber)
	if err != nil {
		return nil, err
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset ne sme biti negativan")
	}
	if offset+size > len(*valueFound) {
		return nil, fmt.Errorf("offset i size su takvi da se traze bajtovi van opsega jednog bloka, toest njegovog sadrzaja")
	}
	//iz celokupnog bloka uzima bajtove koje je korisnik zatrazio-od offseta, do trazene kolicine
	specific := (*valueFound)[offset : offset+size]
	return &specific, err
}

// filepath-path do fajla iz kojeg se dobavljaju informacije
// blockNumber-broj bloka koji se trazi
// writeValue-niz bajtova koji treba da se upise na trazen blok
//
// funkcija vraca samo error, u slucaju da je nesto poslo po zlu
//
// BITNO: kada upisujete nesto u fajl ovom metodom, CEO SADRZAJ BLOKA se override-uje
// na vam je da unutar vasih struktura regulisete kako se vasi zapisi dele po blokovima.
func (bm *BlockManager) Put(filepath string, blockNumber int, writeValue *[]byte) error {
	err := bm.blockCache.Put(filepath, blockNumber, writeValue)
	return err
}

// filepath-path do fajla iz kojeg se dobavljaju informacije
// blockNumber-broj bloka koji se trazi
// writeValue-niz bajtova koji treba da se upise na trazen blok
// offset-offset od pocetka bloka
// size-kolicina bajtova koja treba da se upise
//
// funkcija vraca samo error, u slucaju da je nesto poslo po zlu
//
// BITNO:funkcija vraca error za negativan offset, i ako se zbog offseta i size-a izadje van
// opsega samog bloka
func (bm *BlockManager) PutSpecific(filepath string, blockNumber int, offset int, size int, writeValue *[]byte) error {
	if offset < 0 {
		return fmt.Errorf("offset ne sme biti negativan")
	}
	valueFound, err := bm.blockCache.Get(filepath, blockNumber)
	if err != nil {
		if err == io.EOF {
			err = bm.blockCache.Put(filepath, blockNumber, writeValue)
			return err
		} else {
			return err
		}
	}
	if offset+size > bm.blockCache.GetBlockSize() {
		return fmt.Errorf("offset i size su takvi da se traze bajtovi van opsega jednog bloka")
	}
	i := 0
	//Za svaki bajt, krecuci od offseta, menjamo bajt unutar vrednosti koja je pronadjena na bloku
	//sa vrednoscu koju mi zelimo da stoji tu.
	//Ovo radimo do momenta kada ili zapisemo sve sto smo hteli da zapisemo, ili do momenta kad
	//naletimo na kraj onoga sto je tu trenutno zapisano.
	for i+offset < len(*valueFound) && i < int(size) {
		(*valueFound)[i+offset] = (*writeValue)[i]
		i++
	}
	//ako se ispostavlja da je velicina niza bajtova kojeg smo nasli na bloku premala da bi stalo sve sto zelimo
	//da zapisemo, mi cemo sve sto nije uspelo da stane u vec postojeci niz da konkateniramo na kraj.
	if size+offset > len(*valueFound) {
		remainingBytes := (*writeValue)[i:size]
		finalWriteValue := append(*valueFound, remainingBytes...)
		err = bm.blockCache.Put(filepath, blockNumber, &finalWriteValue)
		return err
	} else {
		//Ako je sve stalo pri inicijalnom prolasku kroz oba niza, nista ne konkateniramo
		//nego zapisujemo nas izmenjeni valueFound
		err = bm.blockCache.Put(filepath, blockNumber, valueFound)
		return err
	}

}

// filepath-path do fajla iz kojeg se dobavljaju informacije
// blockNumber-broj bloka koji se trazi
//
// Funkcija dodaje buffer na kraj bloka-proverava koliko je ostalo mesta do kraja bloka,
// i konkatenira onoliko nula koliko je potrebno.
//
// Posle toga upisuje blok u fajl, i vraca vrednost koja je upisana, radi provere
func (bm *BlockManager) AddBuffer(filepath string, blockNumber int) (*[]byte, error) {
	valueFound, err := bm.blockCache.Get(filepath, blockNumber)
	if err != nil {
		return nil, err
	}
	if len(*valueFound) < bm.blockCache.GetBlockSize() {
		zeroesToAdd := bm.blockCache.GetBlockSize() - len(*valueFound)
		zeroBytes := make([]byte, zeroesToAdd)
		returnValue := append(*valueFound, zeroBytes...)
		err = bm.Put(filepath, blockNumber, &returnValue)
		if err != nil {
			return nil, err
		}
		return &returnValue, nil
	} else {
		return valueFound, nil
	}
}

// Funkcija koja vraća veličinu blokova unutar BufferPool-a, u bajtima
func (bm *BlockManager) GetBlockSize() int {
	return bm.blockCache.GetBlockSize()
}
