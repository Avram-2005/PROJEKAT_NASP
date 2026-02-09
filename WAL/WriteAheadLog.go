package wal

import (
	"encoding/binary"
	"fmt"
	"io"
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
	readFile             *os.File //putanja trenutno aktivnog fajla za čitanje
	writeFile            *os.File //putanja trenutno aktivnog fajla za pisanje
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

func openFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
}

func openFileRead(path string) (*os.File, error) {
	return os.Open(path)
}

// Kreira novi WAL ili ucitava postojeci sa diska
func CreatNewWAL(sizeSegment int, blocksize int) (*WAL, error) {
	bm, err := BlockManager.NewBlockManager(2, blocksize)
	if err != nil {
		return nil, err
	}

	segments := make([]string, 0)

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

	// Ako nema segmenata – napravi prvi
	if len(segments) == 0 {
		firstSegment := filepath.Join(FILE_PATH, fmt.Sprintf("%s0000%s", SEGMENT_NAME, FILE_EXTENSION))
		file, err := openFile(firstSegment)
		if err != nil {
			return nil, err
		}
		file.Close()
		segments = append(segments, firstSegment)
	}

	readPath := segments[0]
	writePath := segments[len(segments)-1]

	readFile, err := openFileRead(readPath)
	if err != nil {
		return nil, err
	}

	writeFile, err := openFile(writePath)
	if err != nil {
		return nil, err
	}

	pos, err := writeFile.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	return &WAL{
		segmentList:          segments,
		readFile:             readFile,
		writeFile:            writeFile,
		currentWritePosition: int(pos),
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

	// Ako nema mesta → rotacija segmenta
	if wal.currentWritePosition+len(binaryData) > wal.segmentSize {
		wal.writeFile.Close()

		newIndex := len(wal.segmentList)
		newSegment := filepath.Join(FILE_PATH, fmt.Sprintf("%s%04d%s", SEGMENT_NAME, newIndex, FILE_EXTENSION))

		file, err := openFile(newSegment)
		if err != nil {
			return err
		}

		wal.segmentList = append(wal.segmentList, newSegment)
		wal.writeFile = file
		wal.currentWritePosition = 0
	}

	blockNumber := wal.currentWritePosition / wal.blockManager.GetBlockSize()
	if (wal.currentWritePosition+len(binaryData))/wal.blockManager.GetBlockSize() > blockNumber {
		blockNumber++
	}
	fmt.Println(wal.currentWritePosition, wal.blockManager.GetBlockSize(), blockNumber)
	offset := wal.currentWritePosition % wal.blockManager.GetBlockSize()
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
	err = wal.writeFile.Sync()
	if err != nil {
		panic(err)
	}
	return nil
}

// Brise WAL =(
func (wal *WAL) ClearWAL() {
}

// Cita sve WAL zapise i radi X sa njima
func (wal *WAL) ReadAll() {
	blockSize := wal.blockManager.GetBlockSize()

	for _, segmentPath := range wal.segmentList {
		file, err := openFileRead(segmentPath)
		if err != nil {
			fmt.Println("Greška pri otvaranju:", err)
			continue
		}

		blockNumber := 0
		offset := 0

		for {
			block, err := wal.blockManager.Get(file, blockNumber)
			if err != nil || block == nil {
				break
			}

			if offset+HEADER_SIZE > blockSize {
				blockNumber++
				offset = 0
				continue
			}

			header := (*block)[offset : offset+HEADER_SIZE]

			keySize := binary.BigEndian.Uint64(header[13:21])
			valueSize := binary.BigEndian.Uint64(header[21:29])
			recordSize := HEADER_SIZE + int(keySize) + int(valueSize)
			if keySize == 0 && valueSize == 0 {
				break
			}
			if offset+recordSize > blockSize {
				blockNumber++
				offset = 0
				continue
			}

			data := (*block)[offset : offset+recordSize]

			logEntry, err := FromBinary(data)
			if err != nil {
				break
			}

			fmt.Println(logEntry)
			offset += recordSize
		}

		file.Close()
	}
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

func (wal *WAL) Close() {
	if wal.readFile != nil {
		wal.readFile.Close()
	}
	if wal.writeFile != nil {
		wal.writeFile.Close()
	}
}
