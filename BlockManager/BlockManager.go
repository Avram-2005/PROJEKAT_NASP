package BlockManager

import (
	"fmt"

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
// Iz tog razloga, jedine dozvoljene vrednosti za blockSize su 4, 8 i 16, ako stavite ista drugo dobicete error.
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
// funkcija vraca niz bajtova, i error u slucaju da je nesto poslo po zlu
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
// writeValue-niz bajtova koji treba da se upise na trazen blok
// funkcija vraca samo error, u slucaju da je nesto poslo po zlu
// BITNO: kada upisujete nesto u fajl ovom metodom, CEO SADRZAJ BLOKA se override-uje
// na vam je da unutar vasih struktura regulisete kako se vasi zapisi dele po blokovima.
func (bm *BlockManager) Put(filepath string, blockNumber int, writeValue *[]byte) error {
	err := bm.blockCache.Put(filepath, blockNumber, writeValue)
	return err
}
