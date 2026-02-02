package wal

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	BlockManager "github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

// Sve potrebno za WAL
// Nisu konacne velicine
// CRC 4B | Timestamp 8B | Tombstone 1B | KeySize 4B | ValueSize 8B | Key ... | Value ...
type WAL struct {
	segmentList          []string //lista WAL segmenata
	readFile             string   //putanja trenutno aktivnog fajla za čitanje
	writeFile            string   //putanja trenutno aktivnog fajla za pisanje
	currentWritePosition int      //pozicija gde se sledeći log upisuje
	currentReadPosition  int      //pozicija za čitanje
	segmentSize          int      //maksimalna veličina segmenta
	blockManager         *BlockManager.BlockManager
}

const (
	MAX_SIZE       = 1000
	SEGMENT_NAME   = "wal_"
	FILE_PATH      = "./WAL/walDATA"
	FILE_EXTENSION = ".wal"
	HEADER_SIZE    = 29
)

// Kreira novi WAL ili ucitava postojeci sa diska
func CreatNewWAL(sizeSegment int, blocksize int) (*WAL, error) {
	bm, err := BlockManager.NewBlockManager(sizeSegment, blocksize)
	if err != nil {
		return nil, err
	}

	segments := make([]string, 0)

	//Ucitavanje svih WAL segmenata iz foldera
	entries, err := os.ReadDir(FILE_PATH)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasPrefix(entry.Name(), SEGMENT_NAME) {
			segments = append(segments, filepath.Join(FILE_PATH, entry.Name()))
		}
	}

	var readFile string
	var writeFile string
	var currentWritePosition int

	//Ako nema nijednog segmenta, odnosno pravi se prvi put
	if len(segments) == 0 {
		firstSegment := filepath.Join(FILE_PATH, fmt.Sprintf("%s0000%s", SEGMENT_NAME, FILE_EXTENSION))

		_, err := os.OpenFile(firstSegment, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}

		segments = append(segments, firstSegment)
		readFile = firstSegment
		writeFile = firstSegment
		currentWritePosition = 0
	} else {
		//Prvi segment se koristi za citanje
		readFile = segments[0]

		//Poslednji segment se koristi za pisanje
		writeFile = segments[len(segments)-1]

		//Pozicija pisanja se postavlja na kraj fajla
		file, err := os.Open(writeFile)
		if err != nil {
			log.Fatal(err)
		}
		pos, err := file.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		currentWritePosition = int(pos)
	}

	return &WAL{
		segmentList:          segments,
		readFile:             readFile,
		writeFile:            writeFile,
		currentWritePosition: currentWritePosition,
		currentReadPosition:  0,
		segmentSize:          sizeSegment,
		blockManager:         bm,
	}, nil
}

// Dodavanje novog log zapisa
func (wal *WAL) AddLog(key string, value []byte) error {
	if len(key) == 0 {
		return fmt.Errorf("Kljuc je prazan!")
	}

	logEntry, err := NewLog(key, value, false, time.Now())
	if err != nil {
		return err
	}

	return wal.appendLog(logEntry)
}

// Dodavanje log zapisa za brisanje kljuca
func (wal *WAL) DeleteLog(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("Kljuc je prazan!")
	}

	logEntry, err := NewLog(key, nil, true, time.Now())
	if err != nil {
		return err
	}

	return wal.appendLog(logEntry)
}

// Fizicki upis log zapisa u WAL
func (wal *WAL) appendLog(newLog *Log) error {
	binaryData, err := newLog.ToBinary()
	if err != nil {
		return err
	}

	//Ako nema mesta u trenutnom segmentu pravi se novi
	//Poraditi na logici
	if wal.currentWritePosition+len(binaryData) > wal.segmentSize {
		newIndex := len(wal.segmentList)
		newSegment := filepath.Join(FILE_PATH, fmt.Sprintf("%s%04d%s", SEGMENT_NAME, newIndex, FILE_EXTENSION))

		file, err := os.OpenFile(newSegment, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		file.Close()

		wal.segmentList = append(wal.segmentList, newSegment)
		wal.writeFile = newSegment
		wal.currentWritePosition = 0
	}

	//Racunanje bloka i offseta
	blockSize := wal.blockManager.GetBlockSize()
	blockNumber := wal.currentWritePosition / blockSize
	offset := wal.currentWritePosition % blockSize

	//Upis podataka
	err = wal.blockManager.PutSpecific(
		wal.writeFile,
		blockNumber,
		offset,
		len(binaryData),
		&binaryData,
	)
	if err != nil {
		return err
	}

	wal.currentWritePosition += len(binaryData)
	return nil
}

// Brise WAL =(
func (wal *WAL) ClearWAL() {
}

// Cita sve WAL zapise =(
func (wal *WAL) ReadAll() {
}

// Cita samo zapise za dati kljuc =(
func (wal *WAL) ReadSpecific(key string) {
}

// Konvertuk binarni zapis u Log
func FromBinary(data []byte) (*Log, error) {
	offset := 0

	//Provera minimalne velicine header-a
	if len(data) < HEADER_SIZE {
		return nil, fmt.Errorf("Losi podaci, premalo podataka!")
	}

	//CRC
	crc := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	//Timestamp
	ts := make([]byte, 8)
	copy(ts, data[offset:offset+8])
	offset += 8

	//Tombstone
	tombstone := data[offset] == 1
	offset++

	//Velicine kljuca i vrednosti
	keySize := binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	valueSize := binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	//Provera duzine podataka
	if len(data) < HEADER_SIZE+int(keySize)+int(valueSize) {
		return nil, fmt.Errorf("Greska pri konvertovanju iz binarnog oblika!")
	}

	//Kljuc
	key := string(data[offset : offset+int(keySize)])
	offset += int(keySize)

	//Vrednost
	value := make([]byte, valueSize)
	copy(value, data[offset:offset+int(valueSize)])

	return &Log{
		CRC:       crc,
		Timestamp: ts,
		Tombstone: tombstone,
		KeySize:   keySize,
		ValueSize: valueSize,
		Key:       key,
		Value:     value,
	}, nil
}

func (wal *WAL) Recovery() {
}
