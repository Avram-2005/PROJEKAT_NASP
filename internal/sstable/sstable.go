package sstable

import (
	"fmt"
	"os"
	"path/filepath"

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

// FIXME: Delete this after config is done
var tablesRoot string
var summaryInterval int
var multipleFiles bool

const BLOOM_FILTER_RATE = 0.01

func SetupSSTable(root string, summaryInt int, multFiles bool) error {
	summaryInterval = summaryInt
	multipleFiles = multFiles
	tablesRoot = filepath.Join(root, "tables")
	return os.MkdirAll(tablesRoot, os.ModePerm)
}

// TODO: Compression (1.3[DZ3])
func Flush(mem Memtable, tableNum int, bm *BlockManager.BlockManager) error {
	if multipleFiles {
		return multipleFilesFlush(mem, tableNum, bm)
	}
	return oneFileFlush(mem, tableNum, bm)
}

// FIXME: Get that looks through all the sstables
func Get(key string, tableNum int, bm *BlockManager.BlockManager) ([]byte, error) {
	oneFileTableFilename := sstableFilenameOneFile(tableNum)
	if _, err := os.Stat(oneFileTableFilename); err == nil {
		return getOneFile(key, tableNum, bm)
	}
	return getMultipleFiles(key, tableNum, bm)
}

func sstableFilename(tableNum int, fileType string) string {
	return filepath.Join(tablesRoot, fmt.Sprintf("usertable-%d-%s.txt", tableNum, fileType))
}

func sstableFilenameOneFile(tableNum int) string {
	return filepath.Join(tablesRoot, fmt.Sprintf("sstable%d", tableNum))
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
