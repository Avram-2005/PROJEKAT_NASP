package sstable

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

// FIXME: Delete this after config is done
type SSTableConfig struct {
	SummaryInterval int
	MultipleFiles   bool
}

type SSTableManager struct {
	config     SSTableConfig
	TablesRoot string
	bm         *BlockManager.BlockManager
	numTables  int
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
func (sstm *SSTableManager) Flush(entries []*Record) (*SSTable, error) {
	var sst *SSTable
	var err error
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Key != entries[j].Key {
			return entries[i].Key < entries[j].Key
		}
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})
	if sstm.config.MultipleFiles {
		sst, err = sstm.multipleFilesFlush(entries, sstm.numTables)
	} else {
		sst, err = sstm.oneFileFlush(entries, sstm.numTables)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to flush memtable: %v", err)
	}
	sstm.numTables++
	return sst, err
}

func (sstm *SSTableManager) Merge(ssts []*SSTable, level int, shouldDeleteTombstones bool) (*SSTable, error) {
	var sst *SSTable
	var err error
	if sstm.config.MultipleFiles {
		sst, err = sstm.multipleFilesMerge(ssts, level, sstm.numTables, shouldDeleteTombstones)
	} else {
		sst, err = sstm.oneFileMerge(ssts, level, sstm.numTables, shouldDeleteTombstones)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to merge SSTables: %v", err)
	}
	sstm.numTables++
	return sst, nil
}

func (sstm *SSTableManager) Get(key string, sst *SSTable) (*Record, error) {
	if sst.isMultFiles {
		return sstm.getMultipleFiles(key, sst)
	}
	return sstm.getOneFile(key, sst)
}

func (sstm *SSTableManager) ValidateSSTable(sst *SSTable) (bool, []Record, error) {
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
		dataFile.Close()
		return nil, fmt.Errorf("failed to open index file: %v", err)
	}
	summaryFile, err := os.OpenFile(sstableFilenameMultFile(sstablePath, "Summary"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		dataFile.Close()
		indexFile.Close()
		return nil, fmt.Errorf("failed to open summary file: %v", err)
	}
	filterFile, err := os.OpenFile(sstableFilenameMultFile(sstablePath, "Filter"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		dataFile.Close()
		indexFile.Close()
		summaryFile.Close()
		return nil, fmt.Errorf("failed to open filter file: %v", err)
	}
	metadataFile, err := os.OpenFile(sstableFilenameMultFile(sstablePath, "Metadata"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		dataFile.Close()
		indexFile.Close()
		summaryFile.Close()
		filterFile.Close()
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
