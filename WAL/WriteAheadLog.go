package wal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	utils "github.com/Avram-2005/PROJEKAT_NASP/utils"

	BlockManager "github.com/Avram-2005/PROJEKAT_NASP/BlockManager"

	Record "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

// Sve potrebno za WAL
// Nisu konacne velicine
// CRC 4B | Timestamp 8B | Tombstone 1B | KeySize 4B | ValueSize 4B | Key ... | Value ...
type WAL struct {
	segmentList          []string //lista WAL segmenata
	readFile             *os.File //putanja trenutno aktivnog fajla za čitanje
	writeFile            *os.File //putanja trenutno aktivnog fajla za pisanje
	currentWritePosition int      //pozicija gde se sledeći Record upisuje
	currentReadPosition  int      //pozicija za čitanje
	segmentSize          int      //maksimalna veličina segmenta
	blockManager         *BlockManager.BlockManager
}

const (
	MAX_SIZE       = 1000
	SEGMENT_NAME   = "wal_"
	FILE_PATH      = "./WAL/walDATA"
	FILE_EXTENSION = ".wal"
	HEADER_SIZE    = 21
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
		os.MkdirAll(FILE_PATH, 0755)
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

// Dodavanje novog Record zapisa
func (wal *WAL) AddRecord(key string, value []byte) error {
	if len(key) == 0 {
		return fmt.Errorf("Kljuc je prazan!")
	}

	RecordEntry, err := Record.NewRecord(key, value, false, time.Now())
	if err != nil {
		return err
	}

	return wal.appendRecord(RecordEntry)
}

// Dodavanje Record zapisa za brisanje kljuca
func (wal *WAL) DeleteRecord(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("Kljuc je prazan!")
	}

	RecordEntry, err := Record.NewRecord(key, nil, true, time.Now())
	if err != nil {
		return err
	}

	return wal.appendRecord(RecordEntry)
}

// Fizicki upis Record zapisa u WAL
func (wal *WAL) appendRecord(newRecord *Record.Record) error {
	binaryData := newRecord.Serialize()
	recordLen := len(binaryData)
	blockSize := wal.blockManager.GetBlockSize()

	if recordLen > blockSize {
		return fmt.Errorf("zapis je veći od bloka, fragmentacija trenutno nije podržana")
	}

	offset := wal.currentWritePosition % blockSize
	if offset+recordLen > blockSize {
		padding := blockSize - offset
		wal.currentWritePosition += padding
	}

	if wal.currentWritePosition+recordLen > wal.segmentSize {
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

	blockNumber := wal.currentWritePosition / blockSize
	offset = wal.currentWritePosition % blockSize

	err := wal.blockManager.PutSpecific(
		wal.writeFile,
		blockNumber,
		offset,
		recordLen,
		&binaryData,
	)
	if err != nil {
		return err
	}

	wal.currentWritePosition += recordLen
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
	fmt.Println("Reading all records from WAL:")

	for _, path := range wal.segmentList {
		file, err := openFileRead(path)
		if err != nil {
			continue
		}
		defer file.Close()

		blockNum := 0
		for {
			block, err := wal.blockManager.Get(file, blockNum)
			if err == nil && block != nil && len(*block) != 0 {
				break
			}

			data := *block
			offset := 0

			for offset+HEADER_SIZE <= len(data) {

				reader := utils.NewBufferReaderReuse(data[offset : offset+HEADER_SIZE])

				_ = reader.ReadCRC()
				_ = reader.ReadTimestamp()
				_ = reader.ReadTombstone()
				kSize := int(reader.ReadKeySize())
				vSize := int(reader.ReadValueSize())

				recordSize := HEADER_SIZE + kSize + vSize

				if kSize == 0 && vSize == 0 {
					break
				}

				if offset+recordSize > len(data) {
					break
				}

				rec, err := Record.DeserializeRecord(data[offset : offset+recordSize])
				if err != nil {
					break
				}

				fmt.Printf("[%s] Key: %s | Value: %s\n", path, rec.Key, string(rec.Value))

				offset += recordSize
			}
			blockNum++
		}
	}
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
