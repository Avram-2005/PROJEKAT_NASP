package sstable

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	merkleTree "github.com/Avram-2005/PROJEKAT_NASP/MerkleTree"
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

// TODO: Could store the BloomFilter and Summary here as well
type SSTable struct {
	path        string
	size        uint64
	isMultFiles bool
}

func createSSTable(path string) (*SSTable, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %v", path, err)
	}
	return &SSTable{
		path:        path,
		size:        uint64(info.Size()),
		isMultFiles: info.IsDir(),
	}, nil
}

// TODO: Compression (1.3[DZ3])
func (m *SSTableManager) Flush(mem Memtable, tableNum int) (*SSTable, error) {
	if m.config.MultipleFiles {
		return m.multipleFilesFlush(mem, tableNum)
	}
	return m.oneFileFlush(mem, tableNum)
}

func (sstm *SSTableManager) Get(key string, sst *SSTable) (*Record, error) {
	if sst.isMultFiles {
		return sstm.getMultipleFiles(key, sst.path)
	}
	return sstm.getOneFile(key, sst.path)
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

func (m *SSTableManager) sstableFilepath(level int, tableNum int) string {
	return filepath.Join(m.TablesRoot, fmt.Sprintf("L%d-%010d", level, tableNum))
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

// TODO: This should be a method of SSTable
func (m *SSTableManager) ValidateSSTable(tableNum int) (bool, [][]byte, error) {
	filename := m.sstableFilepath(0, tableNum)
	if isSSTableMultFiles(filename) {
		return m.validateMultipleFiles(tableNum)
	}
	return m.validateOneFile(filename)
}

// TODO: Move this to a separate file
func (m *SSTableManager) validateOneFile(filename string) (bool, [][]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return false, nil, fmt.Errorf("failed to open SSTable file: %v", err)
	}
	defer f.Close()

	footer, err := m.readOneFileFooter(f)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read SSTable footer: %v", err)
	}

	metadataReader := newBlockReader(f, m.bm, footer.MetadataStart)

	stat, _ := f.Stat()
	footerStart := uint64(stat.Size()) - FOOTER_L
	metadataSize := footerStart - footer.MetadataStart

	metadataData := make([]byte, metadataSize)
	_, err = metadataReader.Read(metadataData)
	if err != nil && err != io.EOF {
		return false, nil, err
	}

	originalTree := merkleTree.Deserialize(metadataData)
	if originalTree == nil {
		return false, nil, fmt.Errorf("failed to deserialize merkle tree")
	}

	dataReader := newBlockReader(f, m.bm, footer.DataStart)

	var currentData [][]byte
	for {
		currentOffset := dataReader.CurrOffset()
		if currentOffset >= footer.IndexStart {
			break
		}

		var dataHeaderBuf [DATA_HEADER_L]byte
		_, err := dataReader.Read(dataHeaderBuf[:])
		if err != nil {
			break
		}

		header := DeserializeRecordHeader(dataHeaderBuf[:])
		dataReader.Skip(header.KeySize)
		valueBuf := make([]byte, header.ValueSize)
		_, err = dataReader.Read(valueBuf)
		if err != nil {
			break
		}

		currentData = append(currentData, valueBuf)
	}

	currentTree, err := merkleTree.NewMerkleTree(currentData)
	if err != nil {
		return false, nil, err
	}

	if originalTree.Verify(currentTree.RootHash()) {
		return true, nil, nil
	}

	diffs := merkleTree.FindDifference(originalTree.Root(), currentTree.Root())

	return false, diffs, nil
}

func (m *SSTableManager) validateMultipleFiles(tableNum int) (bool, [][]byte, error) {
	sstablePath := m.sstableFilepath(0, tableNum)
	metadataFilename := sstableFilenameMultFile(sstablePath, "Metadata")
	metadataFile, err := os.Open(metadataFilename)
	if err != nil {
		return false, nil, fmt.Errorf("failed to open metadata file: %v", err)
	}
	defer metadataFile.Close()

	metadataReader := newBlockReader(metadataFile, m.bm, 0)

	sizeHeader := make([]byte, 4)
	_, err = metadataReader.Read(sizeHeader)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read size header: %v", err)
	}
	treeSize := binary.BigEndian.Uint32(sizeHeader)

	if treeSize == 0 {
		return false, nil, fmt.Errorf("invalid tree size: %d", treeSize)
	}

	metadataData := make([]byte, treeSize)
	_, err = metadataReader.Read(metadataData)
	if err != nil {
		return false, nil, err
	}

	originalTree := merkleTree.Deserialize(metadataData)
	if originalTree == nil {
		return false, nil, fmt.Errorf("failed to deserialize merkle tree")
	}

	dataFilename := sstableFilenameMultFile(sstablePath, "Data")
	dataFile, err := os.Open(dataFilename)
	if err != nil {
		return false, nil, fmt.Errorf("failed to open data file: %v", err)
	}
	defer dataFile.Close()

	stat, err := dataFile.Stat()
	if err != nil {
		return false, nil, err
	}
	fileSize := stat.Size()

	dataReader := newBlockReader(dataFile, m.bm, 0)

	var currentData [][]byte
	for {
		if int64(dataReader.CurrOffset()) >= fileSize {
			break
		}

		var dataHeaderBuf [DATA_HEADER_L]byte
		_, err := dataReader.Read(dataHeaderBuf[:])
		if err != nil {
			break
		}

		header := DeserializeRecordHeader(dataHeaderBuf[:])
		dataReader.Skip(header.KeySize)
		valueBuf := make([]byte, header.ValueSize)
		_, err = dataReader.Read(valueBuf)
		if err != nil {
			break
		}

		currentData = append(currentData, valueBuf)
	}

	currentTree, err := merkleTree.NewMerkleTree(currentData)
	if err != nil {
		return false, nil, err
	}

	if originalTree.Verify(currentTree.RootHash()) {
		return true, nil, nil
	}

	diffs := merkleTree.FindDifference(originalTree.Root(), currentTree.Root())

	return false, diffs, nil
}
