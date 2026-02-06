package sstable

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

// FIXME: DELETE AFTER Memtable MERGE /
// ////////////////////////////////////

type KeyValue struct {
	Key       string
	Value     []byte
	Tombstone bool //za brisanje, true ako je obrisan
}

type Memtable interface {
	GetSortedEntries() []KeyValue //povratna vred/ parovi kljuc-vred neophodni za sstable
}

//////////////////////////////////////

// TODO: Compression (1.3[DZ3])

// TODO: Save to multiple files (Cassandra) or in one file (LevelDB) (1.3[DZ2])

// FIXME: Delete this after DB structure is done
var tablesRoot string

func SetupDirectory(root string) error {
	tablesRoot = filepath.Join(root, "tables")
	return os.MkdirAll(tablesRoot, os.ModePerm)
}

func sstableFilename(tableNum int, fileType string) string {
	return filepath.Join(tablesRoot, fmt.Sprintf("usertable-%d-%s.txt", tableNum, fileType))
}

func createSSTableFile(fileType string, tableNum int) (*os.File, error) {
	filename := sstableFilename(tableNum, fileType)
	if _, err := os.Stat(filename); err == nil {
		return nil, fmt.Errorf("file %s already exists", filename)
	}
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return f, nil
}

const (
	CRC_L          = 4
	TIMESTAMP_L    = 8
	TOMBSTONE_L    = 1
	KEY_SIZE_L     = 4
	VALUE_SIZE_L   = 4
	OFFSET_L       = 8
	DATA_HEADER_L  = CRC_L + TIMESTAMP_L + TOMBSTONE_L + KEY_SIZE_L + VALUE_SIZE_L
	INDEX_HEADER_L = KEY_SIZE_L + OFFSET_L
)

func writeData(writer *blockWriter, entry KeyValue) int {
	oldOffset := writer.currBlockNum*cap(writer.block) + writer.currByte
	bytesWritten := 0

	var dataHeaderBuf [DATA_HEADER_L]byte

	bytesWritten += CRC_L // Placeholder for CRC

	binary.LittleEndian.PutUint64(dataHeaderBuf[bytesWritten:], uint64(time.Now().UnixNano()))
	bytesWritten += TIMESTAMP_L
	if entry.Tombstone {
		dataHeaderBuf[CRC_L+TIMESTAMP_L] = 1
	} else {
		dataHeaderBuf[CRC_L+TIMESTAMP_L] = 0
	}
	bytesWritten += TOMBSTONE_L

	binary.LittleEndian.PutUint32(dataHeaderBuf[bytesWritten:], uint32(len(entry.Key)))
	bytesWritten += KEY_SIZE_L
	binary.LittleEndian.PutUint32(dataHeaderBuf[bytesWritten:], uint32(len(entry.Value)))
	bytesWritten += VALUE_SIZE_L

	writer.Write(dataHeaderBuf[:])
	writer.Write([]byte(entry.Key))
	writer.Write(entry.Value)
	// FIXME: Calculate and write CRC
	return oldOffset
}

func writeIndex(writer *blockWriter, key string, offset int) int {
	oldOffset := writer.currBlockNum*cap(writer.block) + writer.currByte
	var indexHeaderBuf [INDEX_HEADER_L]byte
	binary.LittleEndian.PutUint32(indexHeaderBuf[0:], uint32(len(key)))
	binary.LittleEndian.PutUint64(indexHeaderBuf[KEY_SIZE_L:], uint64(offset))

	writer.Write(indexHeaderBuf[:])
	writer.Write([]byte(key))
	return oldOffset
}

const SUMMARY_INTERVAL = 100

func Flush(mem Memtable, tableNum int, bm *BlockManager.BlockManager) error {
	// FIXME: Fix after BlockManager file handle fix
	dataFile, err := createSSTableFile("Data", tableNum)
	if err != nil {
		return fmt.Errorf("failed to create data file: %v", err)
	}
	indexFile, err := createSSTableFile("Index", tableNum)
	if err != nil {
		return fmt.Errorf("failed to create index file: %v", err)
	}
	summaryFile, err := createSSTableFile("Summary", tableNum)
	if err != nil {
		return fmt.Errorf("failed to create summary file: %v", err)
	}

	dataWriter := newBlockWriter(dataFile, bm)
	indexWriter := newBlockWriter(indexFile, bm)
	summaryWriter := newBlockWriter(summaryFile, bm)
	for i, entry := range mem.GetSortedEntries() {
		offset := writeData(dataWriter, entry)
		offset = writeIndex(indexWriter, entry.Key, offset)
		if i%SUMMARY_INTERVAL == 0 {
			writeIndex(summaryWriter, entry.Key, offset)
		}
	}
	dataWriter.Finalize()
	indexWriter.Finalize()
	summaryWriter.Finalize()
	if dataWriter.currBlockNum == 0 && dataWriter.currByte == 0 {
		return fmt.Errorf("memtable is empty, no data written")
	}
	return nil
}

func searchForKey(key string, reader *blockReader) (uint64, error) {
	blockSize := uint64(cap(reader.block))
	lastOffset := blockSize*uint64(reader.currBlockNum) + uint64(reader.currByte)
	for {
		var indexHeaderBuf [INDEX_HEADER_L]byte
		n, err := reader.Read(indexHeaderBuf[:])
		log.Printf("Read index header: % x, bytes read: %d", indexHeaderBuf[:n], n)
		if err != nil {
			return 0, fmt.Errorf("failed to read index header: %v", err)
		}
		if n == 0 {
			return lastOffset, nil
		}
		keySize := binary.LittleEndian.Uint32(indexHeaderBuf[0:])
		offset := binary.LittleEndian.Uint64(indexHeaderBuf[KEY_SIZE_L:])

		keyBuf := make([]byte, keySize)
		n, err = reader.Read(keyBuf)
		if err != nil {
			return 0, fmt.Errorf("failed to read key: %v", err)
		}
		if n == 0 {
			return lastOffset, nil
		}
		readKey := string(keyBuf)
		log.Printf("Read key: %s, offset: %d", readKey, offset)
		if key <= readKey {
			return offset, nil
		}
		lastOffset = offset
	}
}

// TODO: Consider doing this zero-copy
func Get(key string, tableNum int, bm *BlockManager.BlockManager) ([]byte, error) {
	offset := uint64(0)

	summaryFilename := sstableFilename(tableNum, "Summary")
	summaryFile, err := os.Open(summaryFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open summary file: %v", err)
	}
	defer summaryFile.Close()
	summaryReader := newBlockReader(summaryFile, bm, offset)

	offset, err = searchForKey(key, summaryReader)
	log.Printf("Summary search returned offset: %d", offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search summary file: %v", err)
	}

	indexFilename := sstableFilename(tableNum, "Index")
	indexFile, err := os.Open(indexFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open index file: %v", err)
	}
	defer indexFile.Close()
	indexReader := newBlockReader(indexFile, bm, offset)

	offset, err = searchForKey(key, indexReader)
	log.Printf("Index search returned offset: %d", offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search index file: %v", err)
	}

	dataFilename := sstableFilename(tableNum, "Data")
	dataFile, err := os.Open(dataFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to open data file: %v", err)
	}
	defer dataFile.Close()
	dataReader := newBlockReader(dataFile, bm, offset)

	var dataHeaderBuf [DATA_HEADER_L]byte
	_, err = dataReader.Read(dataHeaderBuf[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read data header: %v", err)
	}

	// FIXME: Verify CRC
	// crc := dataHeaderBuf[0:CRC_L]
	currByte := CRC_L + TIMESTAMP_L + TOMBSTONE_L
	keySize := binary.LittleEndian.Uint32(dataHeaderBuf[currByte:])
	currByte += KEY_SIZE_L
	valueSize := binary.LittleEndian.Uint32(dataHeaderBuf[currByte:])

	valueBuf := make([]byte, valueSize)
	keyBuf := make([]byte, keySize)
	_, err = dataReader.Read(keyBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %v", err)
	}
	readKey := string(keyBuf)
	if readKey != key {
		return nil, fmt.Errorf("key mismatch: expected %s, got %s", key, readKey)
	}
	_, err = dataReader.Read(valueBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read value: %v", err)
	}
	return valueBuf, nil
}
