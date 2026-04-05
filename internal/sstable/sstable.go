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
// TODO: Use an automatic counter for tableNum instead of passing it as a parameter
func Flush(mem Memtable, tableNum int, bm *BlockManager.BlockManager) error {
	if multipleFiles {
		return multipleFilesFlush(mem, tableNum, bm)
	}
	return oneFileFlush(mem, tableNum, bm)
}

// FIXME: Deal with tombstones after Record merge
func Get(key string, bm *BlockManager.BlockManager) ([]byte, error) {
	files, err := os.ReadDir(tablesRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read tables directory: %v", err)
	}

	for _, file := range files {
		sstablePath := filepath.Join(tablesRoot, file.Name())
		val, err := GetSpecific(key, sstablePath, bm)
		if err != nil {
			return nil, fmt.Errorf("error getting key from SSTable %s: %v", sstablePath, err)
		}
		if val != nil {
			return val, nil
		}
	}

	return nil, fmt.Errorf("key %s not found in any SSTable", key)
}

func GetSpecific(key string, sstablePath string, bm *BlockManager.BlockManager) ([]byte, error) {
	if isSSTableMultFiles(sstablePath) {
		return getMultipleFiles(key, sstablePath, bm)
	}
	return getOneFile(key, sstablePath, bm)
}

func isSSTableMultFiles(sstablePath string) bool {
	info, err := os.Stat(sstablePath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func sstableFilenameMultFile(sstablePath string, fileType string) string {
	return filepath.Join(sstablePath, fmt.Sprintf("usertable-%s.txt", fileType))
}

// FIXME: Set the level after LSM Tree is implemented
func sstableFilepath(tableNum int) string {
	return filepath.Join(tablesRoot, fmt.Sprintf("L%d-%010d", 1, tableNum))
}

func createSSTableFile(fileType string, sstablePath string) (*os.File, error) {
	filename := sstableFilenameMultFile(sstablePath, fileType)
	if _, err := os.Stat(filename); err == nil {
		return nil, fmt.Errorf("file %s already exists", filename)
	}

	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return f, nil
}

type sstableFiles struct {
	dataFile    *os.File
	indexFile   *os.File
	summaryFile *os.File
	filterFile  *os.File
}

func createMultipleFiles(sstablePath string) (*sstableFiles, error) {
	dataFile, err := createSSTableFile("Data", sstablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create data file: %v", err)
	}
	indexFile, err := createSSTableFile("Index", sstablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create index file: %v", err)
	}
	summaryFile, err := createSSTableFile("Summary", sstablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create summary file: %v", err)
	}
	filterFile, err := createSSTableFile("Filter", sstablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create filter file: %v", err)
	}

	return &sstableFiles{
		dataFile:    dataFile,
		indexFile:   indexFile,
		summaryFile: summaryFile,
		filterFile:  filterFile,
	}, nil
}

func (files *sstableFiles) close() {
	files.dataFile.Close()
	files.indexFile.Close()
	files.summaryFile.Close()
	files.filterFile.Close()
}
