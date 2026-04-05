package sstable

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
)

func readNextIndexEntry(reader *blockReader) (key string, offset uint64, n int, err error) {
	bufferReader := newBufferReader(INDEX_HEADER_L)
	n, err = reader.Read(bufferReader.buf)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to read index header: %v", err)
	}
	if n == 0 {
		return "", 0, 0, nil
	}

	keySize := bufferReader.ReadKeySize()
	offset = bufferReader.ReadOffset()

	bufferReader = newBufferReader(keySize)
	n, err = reader.Read(bufferReader.buf)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to read key: %v", err)
	}
	if n == 0 {
		return "", 0, 0, nil
	}

	return string(bufferReader.buf), offset, n, nil
}

func findNextKeyOffset(searchKey string, reader *blockReader) (uint64, error) {
	lastOffset := uint64(0)
	for {
		readKey, offset, n, err := readNextIndexEntry(reader)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return lastOffset, nil
		}

		if searchKey <= readKey {
			return offset, nil
		}

		lastOffset = offset
	}
}

func findPreviousKeyOffset(searchKey string, reader *blockReader) (uint64, error) {
	lastOffset := uint64(0)
	for {
		readKey, offset, n, err := readNextIndexEntry(reader)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return lastOffset, nil
		}

		if searchKey < readKey {
			return lastOffset, nil
		}

		lastOffset = offset
	}
}

func searchIndex(indexFilename string, key string, bm *BlockManager.BlockManager, oldOffset uint64) (uint64, error) {
	indexFile, err := os.Open(indexFilename)
	if err != nil {
		return 0, fmt.Errorf("failed to open index file: %v", err)
	}
	defer indexFile.Close()
	indexReader := newBlockReader(indexFile, bm, oldOffset)

	return findNextKeyOffset(key, indexReader)
}

func readKey(reader *blockReader, keySize int) (string, error) {
	keyBuf := make([]byte, keySize)
	_, err := reader.Read(keyBuf)
	if err != nil {
		return "", fmt.Errorf("failed to read key: %v", err)
	}
	return string(keyBuf), nil
}

func searchSummary(summaryFilename string, offset uint64, key string, bm *BlockManager.BlockManager) (bool, uint64, error) {
	summaryFile, err := os.Open(summaryFilename)
	if err != nil {
		return false, 0, fmt.Errorf("failed to open summary file: %v", err)
	}
	defer summaryFile.Close()
	summaryReader := newBlockReader(summaryFile, bm, offset)

	bufferReader := newBufferReader(2 * KEY_SIZE_L)
	_, err = summaryReader.Read(bufferReader.buf)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read summary header: %v", err)
	}

	firstKeySize := bufferReader.ReadKeySize()
	lastKeySize := bufferReader.ReadKeySize()

	firstKey, err := readKey(summaryReader, firstKeySize)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read first key: %v", err)
	}
	lastKey, err := readKey(summaryReader, lastKeySize)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read last key: %v", err)
	}

	if key < firstKey || key > lastKey {
		return false, 0, nil
	}

	offset, err = findPreviousKeyOffset(key, summaryReader)
	return true, offset, err
}

func searchFilter(filterFilename string, offset uint64, readSize uint64, key string, bm *BlockManager.BlockManager) (bool, error) {
	filterFile, err := os.Open(filterFilename)
	if err != nil {
		return false, fmt.Errorf("failed to open filter file: %v", err)
	}
	defer filterFile.Close()

	if readSize == 0 {
		stat, err := filterFile.Stat()
		if err != nil {
			return false, err
		}
		readSize = uint64(stat.Size())
	}

	filterReader := newBlockReader(filterFile, bm, offset)
	filterData := make([]byte, readSize)
	_, err = filterReader.Read(filterData)
	if err != nil {
		return false, err
	}

	bf := BloomFilter.LoadBloomFilter(filterData)

	return bf.IsFound([]byte(key)), nil
}

// FIXME: Verify CRC
// FIXME: Return a Record struct
func parseData(dataFilename string, offset uint64, key string, bm *BlockManager.BlockManager) ([]byte, error) {
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

	currByte := CRC_L + TIMESTAMP_L + TOMBSTONE_L
	keySize := binary.BigEndian.Uint32(dataHeaderBuf[currByte:])
	currByte += KEY_SIZE_L
	valueSize := binary.BigEndian.Uint32(dataHeaderBuf[currByte:])

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

func getMultipleFiles(key string, sstablePath string, bm *BlockManager.BlockManager) ([]byte, error) {
	filterFilename := sstableFilenameMultFile(sstablePath, "Filter")
	isFound, err := searchFilter(filterFilename, 0, 0, key, bm)
	if err != nil {
		return nil, fmt.Errorf("failed to read bloom filter: %v", err)
	}

	// kljuc se ne nalazi u sstable
	if !isFound {
		return nil, nil
	}

	summaryFilename := sstableFilenameMultFile(sstablePath, "Summary")
	isFound, offset, err := searchSummary(summaryFilename, 0, key, bm)
	if err != nil {
		return nil, fmt.Errorf("failed to search summary file: %v", err)
	}
	if !isFound {
		return nil, nil
	}

	indexFilename := sstableFilenameMultFile(sstablePath, "Index")
	offset, err = searchIndex(indexFilename, key, bm, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search index file: %v", err)
	}

	dataFilename := sstableFilenameMultFile(sstablePath, "Data")
	return parseData(dataFilename, offset, key, bm)
}

type oneFileFooter struct {
	SummaryStart uint64
	IndexStart   uint64
	DataStart    uint64
}

func readOneFileFooter(sstablePath string, bm *BlockManager.BlockManager) (*oneFileFooter, error) {
	f, err := os.Open(sstablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSTable file: %v", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat SSTable file: %v", err)
	}
	if stat.Size() < FOOTER_L {
		return nil, fmt.Errorf("file size is too small to contain footer")
	}

	offset := uint64(stat.Size() - FOOTER_L)
	reader := newBlockReader(f, bm, offset)

	bufferReader := newBufferReader(FOOTER_L)
	_, err = reader.Read(bufferReader.buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %v", err)
	}

	footer := &oneFileFooter{
		SummaryStart: bufferReader.ReadOffset(),
		IndexStart:   bufferReader.ReadOffset(),
		DataStart:    bufferReader.ReadOffset(),
	}

	return footer, nil
}

func getOneFile(key string, sstablePath string, bm *BlockManager.BlockManager) ([]byte, error) {
	footer, err := readOneFileFooter(sstablePath, bm)
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %v", err)
	}

	isFound, err := searchFilter(sstablePath, 0, footer.DataStart, key, bm)
	if err != nil {
		return nil, fmt.Errorf("failed to read bloom filter: %v", err)
	}

	// kljuc se ne nalazi u sstable
	if !isFound {
		return nil, nil
	}

	isFound, offset, err := searchSummary(sstablePath, footer.SummaryStart, key, bm)
	if err != nil {
		return nil, fmt.Errorf("failed to search summary file: %v", err)
	}
	if !isFound {
		return nil, nil
	}

	offset, err = searchIndex(sstablePath, key, bm, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search index file: %v", err)
	}

	return parseData(sstablePath, offset, key, bm)
}
