package sstable

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
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
	SummaryInterval int
	MultipleFiles   bool
}

type SSTableManager struct {
	config     SSTableConfig
	TablesRoot string
	bm         *BlockManager.BlockManager
}

const BLOOM_FILTER_RATE = 0.01

func SetupSSTableManager(root string, config SSTableConfig, bm *BlockManager.BlockManager) (*SSTableManager, error) {
	tablesRoot := filepath.Join(root, "tables")
	err := os.MkdirAll(tablesRoot, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables directory: %v", err)
	}

	return &SSTableManager{
		TablesRoot: tablesRoot,
		config:     config,
		bm:         bm,
	}, nil
}

// TODO: Could store the Summary here as well
type SSTable struct {
	path        string
	size        uint64
	isMultFiles bool
	footer      *OneFileFooter
	filter      *BloomFilter.BloomFilter
	summary     *Summary
}

type OneFileFooter struct {
	IndexStart    uint64
	SummaryStart  uint64
	MetadataStart uint64
	FilterStart   uint64
	FooterStart   uint64
}

type Summary struct {
	firstKey string
	lastKey  string
	entries  []indexEntry
}

type indexEntry struct {
	Key    string
	Offset uint64
}

func (sstm *SSTableManager) NewSummary(numRecs uint) *Summary {
	return &Summary{
		entries: make([]indexEntry, 0, numRecs/uint(sstm.config.SummaryInterval+1)),
	}
}

func (s *Summary) SetFirstAndLast(firstKey string, lastKey string) {
	s.firstKey = firstKey
	s.lastKey = lastKey
}

func (s *Summary) AddEntry(key string, offset uint64) {
	s.entries = append(s.entries, indexEntry{
		Key:    key,
		Offset: offset,
	})
}

func (s *Summary) IsFound(key string) (bool, uint64, error) {
	if key < s.firstKey || key > s.lastKey {
		return false, 0, nil
	}

	low, high := 0, len(s.entries)-1
	for low <= high {
		mid := low + (high-low)/2
		if s.entries[mid].Key == key {
			return true, s.entries[mid].Offset, nil
		} else if s.entries[mid].Key < key {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return true, s.entries[high].Offset, nil
}

// TODO: Compression (1.3[DZ3])
func (sstm *SSTableManager) Flush(mem Memtable, tableNum int) (*SSTable, error) {
	if sstm.config.MultipleFiles {
		return sstm.multipleFilesFlush(mem, tableNum)
	}
	return sstm.oneFileFlush(mem, tableNum)
}

func (sstm *SSTableManager) Merge(ssts []*SSTable, level int, tableNum int) (*SSTable, error) {
	if sstm.config.MultipleFiles {
		return sstm.multipleFilesMerge(ssts, level, tableNum)
	}
	return sstm.oneFileMerge(ssts, level, tableNum)
}

func (sstm *SSTableManager) Get(key string, sst *SSTable) (*Record, error) {
	if sst.isMultFiles {
		return sstm.getMultipleFiles(key, sst)
	}
	return sstm.getOneFile(key, sst)
}

func (sstm *SSTableManager) ValidateSSTable(sst *SSTable) (bool, [][]byte, error) {
	if sst.isMultFiles {
		return sstm.validateMultipleFiles(sst.path)
	}
	return sstm.validateOneFile(sst.path)
}

func sstableFilenameMultFile(sstablePath string, fileType string) string {
	return filepath.Join(sstablePath, fmt.Sprintf("usertable-%s.txt", fileType))
}

func (sstm *SSTableManager) sstableFilepath(level int, tableNum int) string {
	return filepath.Join(sstm.TablesRoot, fmt.Sprintf("L%d-%010d", level, tableNum))
}

func extractLevelNum(filename string) (int, error) {
	var levelNum int
	_, err := fmt.Sscanf(filename, "L%d-", &levelNum)
	if err != nil {
		return 0, fmt.Errorf("failed to extract level number from filename %s: %v", filename, err)
	}
	return levelNum, nil
}

type sstableFiles struct {
	dataFile     *os.File
	indexFile    *os.File
	summaryFile  *os.File
	filterFile   *os.File
	metadataFile *os.File
}

func openMultipleFiles(sstablePath string) (*sstableFiles, error) {
	if err := os.MkdirAll(sstablePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory for multiple files: %v", err)
	}

	dataFile, err := os.OpenFile(sstableFilenameMultFile(sstablePath, "Data"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open data file: %v", err)
	}
	indexFile, err := os.OpenFile(sstableFilenameMultFile(sstablePath, "Index"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open index file: %v", err)
	}
	summaryFile, err := os.OpenFile(sstableFilenameMultFile(sstablePath, "Summary"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open summary file: %v", err)
	}
	filterFile, err := os.OpenFile(sstableFilenameMultFile(sstablePath, "Filter"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open filter file: %v", err)
	}
	metadataFile, err := os.OpenFile(sstableFilenameMultFile(sstablePath, "Metadata"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata file: %v", err)
	}

	return &sstableFiles{
		dataFile:     dataFile,
		indexFile:    indexFile,
		summaryFile:  summaryFile,
		filterFile:   filterFile,
		metadataFile: metadataFile,
	}, nil
}

func (files *sstableFiles) Close() {
	files.dataFile.Close()
	files.indexFile.Close()
	files.summaryFile.Close()
	files.filterFile.Close()
	files.metadataFile.Close()
}
