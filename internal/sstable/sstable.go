package sstable

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
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

func SetupSSTable(root string, summaryInt int, multFiles bool) error {
	summaryInterval = summaryInt
	multipleFiles = multFiles
	tablesRoot = filepath.Join(root, "tables")
	return os.MkdirAll(tablesRoot, os.ModePerm)
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

const (
	CRC_L          = 4
	TIMESTAMP_L    = 8
	TOMBSTONE_L    = 1
	KEY_SIZE_L     = 4
	VALUE_SIZE_L   = 4
	OFFSET_L       = 8
	DATA_HEADER_L  = CRC_L + TIMESTAMP_L + TOMBSTONE_L + KEY_SIZE_L + VALUE_SIZE_L
	INDEX_HEADER_L = KEY_SIZE_L + OFFSET_L
	FOOTER_L       = 2 * OFFSET_L
)

func writeData(writer *blockWriter, entry KeyValue) int {
	oldOffset := writer.CurrOffset()
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
	return int(oldOffset)
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

func writeSummaryHeader(writer *blockWriter, firstKey string, lastKey string) {
	var summaryHeaderBuf [2 * KEY_SIZE_L]byte
	binary.LittleEndian.PutUint32(summaryHeaderBuf[0:], uint32(len(firstKey)))
	binary.LittleEndian.PutUint32(summaryHeaderBuf[KEY_SIZE_L:], uint32(len(lastKey)))

	writer.Write(summaryHeaderBuf[:])
	writer.Write([]byte(firstKey))
	writer.Write([]byte(lastKey))
}

// TODO: Compression (1.3[DZ3])
func Flush(mem Memtable, tableNum int, bm *BlockManager.BlockManager) error {
	if multipleFiles {
		return multipleFilesFlush(mem, tableNum, bm)
	}
	return oneFileFlush(mem, tableNum, bm)
}

func multipleFilesFlush(mem Memtable, tableNum int, bm *BlockManager.BlockManager) error {
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
	filterFile, err := createSSTableFile("Filter", tableNum)
	if err != nil {
		return fmt.Errorf("failed to create filter file: %v", err)
	}

	sortedEntries := mem.GetSortedEntries()
	bf, err := BloomFilter.NewBloomFilter(uint(len(sortedEntries)), 0.01)
	if err != nil {
		return err
	}

	dataWriter := newBlockWriter(dataFile, bm)
	indexWriter := newBlockWriter(indexFile, bm)
	summaryWriter := newBlockWriter(summaryFile, bm)
	filterWriter := newBlockWriter(filterFile, bm)

	firstEntry, lastEntry := sortedEntries[0], sortedEntries[len(sortedEntries)-1]
	writeSummaryHeader(summaryWriter, firstEntry.Key, lastEntry.Key)

	for i, entry := range sortedEntries {
		bf.Set([]byte(entry.Key)) // dodaj kljuc u filter
		offset := writeData(dataWriter, entry)
		offset = writeIndex(indexWriter, entry.Key, offset)
		if i%summaryInterval == 0 {
			writeIndex(summaryWriter, entry.Key, offset)
		}
	}
	filterWriter.Write(bf.Dump())

	dataWriter.Finalize()
	indexWriter.Finalize()
	summaryWriter.Finalize()
	filterWriter.Finalize()

	return nil
}

type indexEntry struct {
	Key    string
	Offset int
}

func oneFileFlush(mem Memtable, tableNum int, bm *BlockManager.BlockManager) error {
	sstableFilename := sstableFilenameOneFile(tableNum)
	f, err := os.Create(sstableFilename)
	if err != nil {
		return fmt.Errorf("failed to create SSTable file: %v", err)
	}
	defer f.Close()

	writer := newBlockWriter(f, bm)

	bf, err := BloomFilter.NewBloomFilter(uint(len(mem.GetSortedEntries())), 0.01)
	if err != nil {
		return err
	}

	for _, entry := range mem.GetSortedEntries() {
		bf.Set([]byte(entry.Key))
	}

	filterStart := writer.CurrOffset()
	filterData := bf.Dump()
	writer.Write(filterData)

	index := make([]indexEntry, 0)
	for _, entry := range mem.GetSortedEntries() {
		offset := writeData(writer, entry)
		index = append(index, indexEntry{
			Key:    entry.Key,
			Offset: offset,
		})
	}

	indexStart := writer.CurrOffset()
	summaryOffsets := make([]indexEntry, 0, len(index)/summaryInterval+1)
	i := 0
	for _, entry := range index {
		indexOffset := writeIndex(writer, entry.Key, entry.Offset)
		if i%summaryInterval == 0 {
			summaryOffsets = append(summaryOffsets, indexEntry{
				Key:    entry.Key,
				Offset: indexOffset,
			})
		}
		i++
	}

	summaryStart := writer.CurrOffset()
	firstEntry, lastEntry := summaryOffsets[0], summaryOffsets[len(summaryOffsets)-1]
	writeSummaryHeader(writer, firstEntry.Key, lastEntry.Key)
	for _, entry := range summaryOffsets {
		writeIndex(writer, entry.Key, entry.Offset)
	}

	if writer.currBlockNum == 0 && writer.currByte == 0 {
		return fmt.Errorf("memtable is empty, no data written")
	}

	var footrerBuf [3 * OFFSET_L]byte
	binary.LittleEndian.PutUint64(footrerBuf[0:], uint64(summaryStart))
	binary.LittleEndian.PutUint64(footrerBuf[OFFSET_L:], uint64(indexStart))
	binary.LittleEndian.PutUint64(footrerBuf[2*OFFSET_L:], uint64(filterStart))
	writer.Write(footrerBuf[:])

	writer.Finalize()
	return nil
}

func searchForKey(key string, reader *blockReader) (uint64, error) {
	lastOffset := reader.CurrOffset()
	for {
		var indexHeaderBuf [INDEX_HEADER_L]byte
		n, err := reader.Read(indexHeaderBuf[:])
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
		if key <= readKey {
			return offset, nil
		}
		lastOffset = offset
	}
}

func searchIndex(tableNum int, key string, bm *BlockManager.BlockManager, oldOffset uint64) (uint64, error) {
	summaryFilename := sstableFilename(tableNum, "Index")
	summaryFile, err := os.Open(summaryFilename)
	if err != nil {
		return 0, fmt.Errorf("failed to open summary file: %v", err)
	}
	defer summaryFile.Close()
	summaryReader := newBlockReader(summaryFile, bm, oldOffset)

	return searchForKey(key, summaryReader)
}

func searchSummary(tableNum int, key string, bm *BlockManager.BlockManager) (bool, uint64, error) {
	summaryFilename := sstableFilename(tableNum, "Summary")
	summaryFile, err := os.Open(summaryFilename)
	if err != nil {
		return false, 0, fmt.Errorf("failed to open summary file: %v", err)
	}
	defer summaryFile.Close()
	summaryReader := newBlockReader(summaryFile, bm, 0)

	var summaryHeaderBuf [2 * KEY_SIZE_L]byte
	_, err = summaryReader.Read(summaryHeaderBuf[:])
	if err != nil {
		return false, 0, fmt.Errorf("failed to read summary header: %v", err)
	}
	firstKeySize := binary.LittleEndian.Uint32(summaryHeaderBuf[0:])
	lastKeySize := binary.LittleEndian.Uint32(summaryHeaderBuf[KEY_SIZE_L:])
	firstKeyBuf := make([]byte, firstKeySize)
	lastKeyBuf := make([]byte, lastKeySize)
	_, err = summaryReader.Read(firstKeyBuf)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read first key from summary header: %v", err)
	}
	_, err = summaryReader.Read(lastKeyBuf)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read last key from summary header: %v", err)
	}
	firstKey := string(firstKeyBuf)
	lastKey := string(lastKeyBuf)
	if key < firstKey || key > lastKey {
		return false, 0, nil
	}

	offset, err := searchForKey(key, summaryReader)
	return true, offset, err
}

func searchFilter(tableNum int, key string, bm *BlockManager.BlockManager) (bool, error) {
	filterFilename := sstableFilename(tableNum, "Filter")
	filterFile, err := os.Open(filterFilename)
	if err != nil {
		return false, fmt.Errorf("failed to open filter file: %v", err)
	}
	defer filterFile.Close()

	stat, err := filterFile.Stat()
	if err != nil {
		return false, err
	}

	filterReader := newBlockReader(filterFile, bm, 0)
	filterData := make([]byte, stat.Size())
	_, err = filterReader.Read(filterData)
	if err != nil {
		return false, err
	}

	bf := BloomFilter.LoadBloomFilter(filterData)

	return bf.IsFound([]byte(key)), nil
}

func Get(key string, tableNum int, bm *BlockManager.BlockManager) ([]byte, error) {
	oneFileTableFilename := sstableFilenameOneFile(tableNum)
	if _, err := os.Stat(oneFileTableFilename); err == nil {
		return getOneFile(key, tableNum, bm)
	}
	return getMultipleFiles(key, tableNum, bm)
}

// TODO: Consider doing this zero-copy
func getMultipleFiles(key string, tableNum int, bm *BlockManager.BlockManager) ([]byte, error) {
	isFound, err := searchFilter(tableNum, key, bm)
	if err != nil {
		return nil, fmt.Errorf("failed to read boom filter: %v", err)
	}

	// kljuc se ne nalazi u sstable
	if !isFound {
		return nil, nil
	}

	isFound, offset, err := searchSummary(tableNum, key, bm)
	if err != nil {
		return nil, fmt.Errorf("failed to search summary file: %v", err)
	}
	if !isFound {
		return nil, nil
	}

	offset, err = searchIndex(tableNum, key, bm, offset)
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

func getOneFile(key string, tableNum int, bm *BlockManager.BlockManager) ([]byte, error) {
	return nil, fmt.Errorf("one file get not implemented yet")
}
