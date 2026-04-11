package wal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	BlockManager "github.com/Avram-2005/PROJEKAT_NASP/BlockManager"

	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"

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
	lowWatermarks        []string //lista segmenata koji su sigurni za brisanje
}

const (
	SEGMENT_NAME   = "wal_"
	FILE_PATH      = "./WAL/walDATA"
	FILE_EXTENSION = ".wal"
	HEADER_SIZE    = 21

	ChunkTypeZero   byte = 0
	ChunkTypeFull   byte = 1
	ChunkTypeFirst  byte = 2
	ChunkTypeMiddle byte = 3
	ChunkTypeLast   byte = 4

	ChunkHeaderSize = 1
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
	data := newRecord.Serialize()
	left := len(data)
	offset := 0
	isFirst := true
	blockSize := wal.blockManager.GetBlockSize()

	for left > 0 {
		if wal.currentWritePosition >= wal.segmentSize {
			if err := wal.rotateSegment(); err != nil {
				return err
			}
		}

		blockOffset := wal.currentWritePosition % blockSize
		remainingInBlock := blockSize - blockOffset

		if remainingInBlock <= 10 {
			wal.currentWritePosition += remainingInBlock
			continue
		}

		payloadSize := remainingInBlock - 1
		if left < payloadSize {
			payloadSize = left
		}

		var chunkType byte
		if isFirst {
			if left == payloadSize {
				chunkType = ChunkTypeFull
			} else {
				chunkType = ChunkTypeFirst
			}
		} else {
			if left == payloadSize {
				chunkType = ChunkTypeLast
			} else {
				chunkType = ChunkTypeMiddle
			}
		}

		chunk := make([]byte, 1+payloadSize)
		chunk[0] = chunkType
		copy(chunk[1:], data[offset:offset+payloadSize])

		err := wal.blockManager.PutSpecific(wal.writeFile, wal.currentWritePosition/blockSize, blockOffset, len(chunk), &chunk)
		if err != nil {
			return err
		}

		wal.currentWritePosition += len(chunk)
		offset += payloadSize
		left -= payloadSize
		isFirst = false
	}

	return wal.writeFile.Sync()
}

func (wal *WAL) rotateSegment() error {
	wal.writeFile.Close()
	newIndex := len(wal.segmentList)
	newPath := filepath.Join(FILE_PATH, fmt.Sprintf("%s%04d%s", SEGMENT_NAME, newIndex, FILE_EXTENSION))

	file, err := openFile(newPath)
	if err != nil {
		return err
	}

	wal.segmentList = append(wal.segmentList, newPath)
	wal.writeFile = file
	wal.currentWritePosition = 0
	return nil
}

func (wal *WAL) memtableRotation() {
	wal.lowWatermarks = append(wal.lowWatermarks, wal.segmentList[len(wal.segmentList)-1])
	if len(wal.lowWatermarks) >= 10 { //Treba uzeti broj iz konfiguracije za memtable
		wal.FlushWAL()
	}
}

// Brise WAL segmente koji su sigurni za brisanje (koji su ispod low water marka)
func (wal *WAL) FlushWAL() error {
	keepIndex := 0
	for i, path := range wal.segmentList {
		if path == wal.lowWatermarks[0] {
			keepIndex = i
			break
		}
		os.Remove(wal.segmentList[i])
	}

	wal.segmentList = wal.segmentList[keepIndex:]
	wal.lowWatermarks = wal.lowWatermarks[1:]
	return nil

}

// Cita sve WAL zapise i upisuje ih u memtable
func (wal *WAL) Recovery(memtableManager *memtable.MemtableManager) error {
	var recordBuffer []byte
	for _, path := range wal.segmentList {
		file, _ := openFileRead(path)
		defer file.Close()

		for blockNum := 0; ; blockNum++ {
			block, err := wal.blockManager.Get(file, blockNum)
			if err != nil || block == nil || len(*block) == 0 {
				break
			}

			data := *block
			offset := 0

			for offset < len(data) {
				chunkType := data[offset]
				if chunkType == ChunkTypeZero {
					break
				}

				payloadStart := offset + 1
				switch chunkType {
				case ChunkTypeFirst, ChunkTypeMiddle:
					recordBuffer = append(recordBuffer, data[payloadStart:]...)
					offset = len(data)

				case ChunkTypeFull, ChunkTypeLast:
					fullData := append(recordBuffer, data[payloadStart:]...)
					rec, consumed, err := Record.DeserializeRecord(fullData)
					if err != nil {
						return err
					}
					if rec.Tombstone {
						err = memtableManager.Delete(rec.Key)
						fmt.Printf("Deleted Key: %s\n", rec.Key)
						if err != nil {
							return err
						}
					} else {
						err = memtableManager.Put(rec.Key, rec.Value)
						fmt.Printf("Read Key: %s Read Value: %s\n", rec.Key, string(rec.Value))
						if err != nil {
							return err
						}
					}

					bytesFromThisBlock := consumed - len(recordBuffer)

					offset = payloadStart + bytesFromThisBlock
					recordBuffer = nil
				}
			}
		}
	}
	return nil
}

func (wal *WAL) Close() {
	if wal.readFile != nil {
		wal.readFile.Close()
	}
	if wal.writeFile != nil {
		wal.writeFile.Close()
	}
}
