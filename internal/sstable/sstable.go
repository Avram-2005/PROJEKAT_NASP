package sstable

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

// FIXME: DELETE AFTER Memtable MERGE /
// ////////////////////////////////////

type Memtable interface {
	GetSortedEntries() []Record //povratna vred/ parovi kljuc-vred neophodni za sstable
}

//////////////////////////////////////

// FIXME: Delete this after config is done
type SSTableConfig struct {
	TablesRoot      string
	SummaryInterval int
	MultipleFiles   bool
}

type SSTableManager struct {
	config     SSTableConfig
	TablesRoot string
	bm         *BlockManager.BlockManager
}

const BLOOM_FILTER_RATE = 0.01

func SetupSSTableManager(root string, summaryInt int, multFiles bool, bm *BlockManager.BlockManager) (*SSTableManager, error) {
	tablesRoot := filepath.Join(root, "tables")
	err := os.MkdirAll(tablesRoot, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables directory: %v", err)
	}

	return &SSTableManager{
		TablesRoot: tablesRoot,
		config: SSTableConfig{
			SummaryInterval: summaryInt,
			MultipleFiles:   multFiles,
		},
		bm: bm,
	}, nil
}

// TODO: Compression (1.3[DZ3])
// TODO: Use an automatic counter for tableNum instead of passing it as a parameter
func (m *SSTableManager) Flush(mem Memtable, tableNum int) error {
	if m.config.MultipleFiles {
		return m.multipleFilesFlush(mem, tableNum)
	}
	return m.oneFileFlush(mem, tableNum)
}

func (m *SSTableManager) Get(key string, bm *BlockManager.BlockManager) ([]byte, error) {
	files, err := os.ReadDir(m.TablesRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read tables directory: %v", err)
	}

	for _, file := range files {
		sstablePath := filepath.Join(m.TablesRoot, file.Name())
		rec, err := GetSpecific(key, sstablePath, bm)
		if err != nil {
			return nil, fmt.Errorf("error getting key from SSTable %s: %v", sstablePath, err)
		}
		if rec != nil {
			return rec.Value, nil
		}
	}

	return nil, fmt.Errorf("key %s not found in any SSTable", key)
}

func GetSpecific(key string, sstablePath string, bm *BlockManager.BlockManager) (*Record, error) {
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
func (m *SSTableManager) sstableFilepath(tableNum int) string {
	return filepath.Join(m.TablesRoot, fmt.Sprintf("L%d-%010d", 1, tableNum))
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
	dataFile     *os.File
	indexFile    *os.File
	summaryFile  *os.File
	filterFile   *os.File
	metadataFile *os.File
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
	metadataFile, err := createSSTableFile("Metadata", sstablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata file: %v", err)
	}

	return &sstableFiles{
		dataFile:     dataFile,
		indexFile:    indexFile,
		summaryFile:  summaryFile,
		filterFile:   filterFile,
		metadataFile: metadataFile,
	}, nil
}

func (files *sstableFiles) close() {
	files.dataFile.Close()
	files.indexFile.Close()
	files.summaryFile.Close()
	files.filterFile.Close()
	files.metadataFile.Close()
}

func (m *SSTableManager) ValidateSSTable(tableNum int) (bool, []Record, error) {
	filename := m.sstableFilepath(tableNum)
	if isSSTableMultFiles(filename) {
		return m.validateMultipleFiles(tableNum)
	}
	return m.validateOneFile(filename)
}
