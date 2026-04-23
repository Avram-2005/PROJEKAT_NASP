package sstable

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
	. "github.com/Avram-2005/PROJEKAT_NASP/utils"
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

// FIXME: Load the Summary as well
func (sstm *SSTableManager) loadSSTable(path string) (*SSTable, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	sst := &SSTable{
		path:        path,
		size:        uint64(info.Size()),
		isMultFiles: info.IsDir(),
	}

	if sst.isMultFiles {
		filter, err := sstm.getFilterMultFile(sst)
		if err != nil {
			return nil, fmt.Errorf("failed to load bloom filter: %v", err)
		}
		sst.filter = filter

		summary, err := sstm.loadSummaryMultFile(sst)
		if err != nil {
			return nil, fmt.Errorf("failed to load summary: %v", err)
		}
		sst.summary = summary
	} else {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open SSTable file: %v", err)
		}
		defer file.Close()

		footer, err := sstm.GetOneFileFooter(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSTable footer: %v", err)
		}
		sst.footer = footer

		filter, err := sstm.getFilterOneFile(footer, file)
		if err != nil {
			return nil, fmt.Errorf("failed to load bloom filter: %v", err)
		}
		sst.filter = filter

		summary, err := sstm.loadSummaryOneFile(footer, file)
		if err != nil {
			return nil, fmt.Errorf("failed to load summary: %v", err)
		}
		sst.summary = summary
	}

	return sst, nil
}

func (sstm *SSTableManager) getFilterMultFile(sst *SSTable) (*BloomFilter.BloomFilter, error) {
	path := sstableFilenameMultFile(sst.path, "Filter")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open filter file: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat filter file: %v", err)
	}
	size := uint64(info.Size())
	start := uint64(0)

	return sstm.loadBloomFilter(file, start, size)
}

func (sstm *SSTableManager) getFilterOneFile(footer *OneFileFooter, file *os.File) (*BloomFilter.BloomFilter, error) {
	size := footer.IndexStart - footer.FilterStart
	start := footer.FilterStart

	return sstm.loadBloomFilter(file, start, size)
}

func (sstm *SSTableManager) loadBloomFilter(file *os.File, start uint64, size uint64) (*BloomFilter.BloomFilter, error) {
	reader := newBlockReader(file, sstm.bm, start)
	data := make([]byte, size)
	_, err := reader.Read(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read bloom filter data: %v", err)
	}

	return BloomFilter.LoadBloomFilter(data), nil
}

func (sstm *SSTableManager) loadSummaryMultFile(sst *SSTable) (*Summary, error) {
	path := sstableFilenameMultFile(sst.path, "Summary")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open summary file: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat summary file: %v", err)
	}
	size := uint64(info.Size())
	start := uint64(0)

	return sstm.loadSummary(file, start, size)
}

func (sstm *SSTableManager) loadSummaryOneFile(footer *OneFileFooter, file *os.File) (*Summary, error) {
	size := footer.MetadataStart - footer.SummaryStart
	start := footer.SummaryStart

	return sstm.loadSummary(file, start, size)
}

func (sstm *SSTableManager) loadSummary(file *os.File, start uint64, size uint64) (*Summary, error) {
	summary := sstm.NewSummary(0)

	reader := newBlockReader(file, sstm.bm, start)
	first, last, err := sstm.loadFirstLastSummaryKeys(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read first and last keys from summary: %v", err)
	}
	summary.SetFirstAndLast(first, last)

	for reader.CurrOffset() < start+size {
		indexEntry, _, err := readNextIndexEntry(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read summary entry: %v", err)
		}
		summary.entries = append(summary.entries, indexEntry)
	}

	return summary, nil
}

// TODO: Compression (1.3[DZ3])
func (sstm *SSTableManager) Flush(mem Memtable, tableNum int) (*SSTable, error) {
	if sstm.config.MultipleFiles {
		return sstm.multipleFilesFlush(mem, tableNum)
	}
	return sstm.oneFileFlush(mem, tableNum)
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

type OneFileFooter struct {
	FilterStart   uint64
	IndexStart    uint64
	SummaryStart  uint64
	MetadataStart uint64
}

func (sstm *SSTableManager) GetOneFileFooter(file *os.File) (*OneFileFooter, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat SSTable file: %v", err)
	}
	if stat.Size() < FOOTER_L {
		return nil, fmt.Errorf("file size is too small to contain footer")
	}

	offset := uint64(stat.Size() - FOOTER_L)
	reader := newBlockReader(file, sstm.bm, offset)

	bufferReader := NewBufferReader(FOOTER_L)
	_, err = reader.Read(bufferReader.Buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %v", err)
	}

	footer := &OneFileFooter{
		FilterStart:   bufferReader.ReadOffset(),
		IndexStart:    bufferReader.ReadOffset(),
		SummaryStart:  bufferReader.ReadOffset(),
		MetadataStart: bufferReader.ReadOffset(),
	}

	return footer, nil
}

func (off *OneFileFooter) Write(writer *blockWriter) {
	footrerBuf := NewBufferWriter(FOOTER_L)
	footrerBuf.WriteOffset(off.FilterStart)
	footrerBuf.WriteOffset(off.IndexStart)
	footrerBuf.WriteOffset(off.SummaryStart)
	footrerBuf.WriteOffset(off.MetadataStart)
	writer.Write(footrerBuf.Buf)
}
