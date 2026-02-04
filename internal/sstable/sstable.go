package sstable

import (
	"encoding/binary"
	"errors"
	"fmt"
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

func Get(key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// FIXME: Delete this after DB structure is done
var tablesRoot string

func SetupDirectory(root string) error {
	tablesRoot = filepath.Join(root, "tables")
	return os.MkdirAll(tablesRoot, os.ModePerm)
}

func createSSTableFile(fileType string, tableNum int) (*os.File, string, error) {
	filename := filepath.Join(tablesRoot, fmt.Sprintf("usertable-%d-%s.txt", tableNum, fileType))
	if _, err := os.Stat(filename); err == nil {
		return nil, filename, fmt.Errorf("file %s already exists", filename)
	}
	f, err := os.Create(filename)
	if err != nil {
		return nil, filename, err
	}

	return f, filename, nil
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
	oldOffset := (writer.currBlockNum-1)*cap(writer.block) + writer.currByte
	bytesWritten := 0

	var dataHeaderBuf [DATA_HEADER_L]byte

	bytesWritten += CRC_L // Placeholder for CRC

	binary.LittleEndian.PutUint64(dataHeaderBuf[bytesWritten:], uint64(time.Now().UnixNano()))
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

func writeIndex(writer *blockWriter, key string, offset int) {
	var indexHeaderBuf [INDEX_HEADER_L]byte
	binary.LittleEndian.PutUint32(indexHeaderBuf[0:], uint32(len(key)))
	binary.LittleEndian.PutUint64(indexHeaderBuf[KEY_SIZE_L:], uint64(offset))

	writer.Write(indexHeaderBuf[:])
	writer.Write([]byte(key))
}

const SUMMARY_INTERVAL = 100

func Flush(mem Memtable, tableNum int, bm *BlockManager.BlockManager) error {
	// FIXME: Fix after BlockManager file handle fix
	_, dataFilename, err := createSSTableFile("Data", tableNum)
	if err != nil {
		return err
	}
	_, indexFilename, err := createSSTableFile("Index", tableNum)
	if err != nil {
		return err
	}
	_, summaryFilename, err := createSSTableFile("Summary", tableNum)
	if err != nil {
		return err
	}

	dataWriter := newBlockWriter(dataFilename, bm)
	indexWriter := newBlockWriter(indexFilename, bm)
	summaryWriter := newBlockWriter(summaryFilename, bm)
	for i, entry := range mem.GetSortedEntries() {
		offset := writeData(dataWriter, entry)
		writeIndex(indexWriter, entry.Key, offset)
		if i%SUMMARY_INTERVAL == 0 {
			writeIndex(summaryWriter, entry.Key, offset)
		}
	}
	dataWriter.Finalize()
	indexWriter.Finalize()
	summaryWriter.Finalize()
	return nil
}
