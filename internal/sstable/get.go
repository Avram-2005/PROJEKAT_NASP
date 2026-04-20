package sstable

import (
	"fmt"
	"hash/crc32"
	"io"
	"os"

	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
	. "github.com/Avram-2005/PROJEKAT_NASP/utils"
)

const (
	DATA_HEADER_L  = CRC_L + TIMESTAMP_L + TOMBSTONE_L + KEY_SIZE_L + VALUE_SIZE_L
	INDEX_HEADER_L = KEY_SIZE_L + OFFSET_L
	FOOTER_L       = 4 * OFFSET_L
)

func readNextIndexEntry(reader *blockReader) (key string, offset uint64, n int, err error) {
	bufferReader := NewBufferReader(INDEX_HEADER_L)
	n, err = reader.Read(bufferReader.Buf)
	if err != nil {
		if err == io.EOF && n == 0 {
			return "", 0, 0, nil
		}
		return "", 0, 0, fmt.Errorf("failed to read index header: %v", err)
	}
	if n == 0 {
		return "", 0, 0, nil
	}
	if n < INDEX_HEADER_L {
		return "", 0, 0, fmt.Errorf("failed to read index header: short read (%d/%d)", n, INDEX_HEADER_L)
	}

	keySize := bufferReader.ReadKeySize()
	offset = bufferReader.ReadOffset()

	bufferReader = NewBufferReader(keySize)
	n, err = reader.Read(bufferReader.Buf)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to read key: %v", err)
	}
	if n == 0 {
		return "", 0, 0, nil
	}
	if n < keySize {
		return "", 0, 0, fmt.Errorf("failed to read key: short read (%d/%d)", n, keySize)
	}

	return string(bufferReader.Buf), offset, n, nil
}

func findNextKeyOffset(searchKey string, reader *blockReader, sectionEnd uint64) (uint64, error) {
	lastOffset := uint64(0)
	for {
		if sectionEnd > 0 {
			curr := reader.CurrOffset()
			if curr >= sectionEnd || sectionEnd-curr < INDEX_HEADER_L {
				return lastOffset, nil
			}
		}

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

func findPreviousKeyOffset(searchKey string, reader *blockReader, sectionEnd uint64) (uint64, error) {
	lastOffset := uint64(0)
	for {
		if sectionEnd > 0 {
			curr := reader.CurrOffset()
			if curr >= sectionEnd || sectionEnd-curr < INDEX_HEADER_L {
				return lastOffset, nil
			}
		}

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

func (m *SSTableManager) searchIndex(file *os.File, key string, oldOffset uint64, sectionEnd uint64) (uint64, error) {
	indexReader := newBlockReader(file, m.bm, oldOffset)
	return findNextKeyOffset(key, indexReader, sectionEnd)
}

func readKey(reader *blockReader, keySize int) (string, error) {
	keyBuf := make([]byte, keySize)
	_, err := reader.Read(keyBuf)
	if err != nil {
		return "", fmt.Errorf("failed to read key: %v", err)
	}
	return string(keyBuf), nil
}

func (m *SSTableManager) searchSummary(file *os.File, offset uint64, sectionEnd uint64, key string) (bool, uint64, error) {
	summaryReader := newBlockReader(file, m.bm, offset)

	bufferReader := NewBufferReader(2 * KEY_SIZE_L)
	_, err := summaryReader.Read(bufferReader.Buf)
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

	offset, err = findPreviousKeyOffset(key, summaryReader, sectionEnd)
	return true, offset, err
}

func (m *SSTableManager) searchFilter(file *os.File, offset uint64, readSize uint64, key string) (bool, error) {
	if readSize == 0 {
		stat, err := file.Stat()
		if err != nil {
			return false, err
		}
		readSize = uint64(stat.Size())
	}

	filterReader := newBlockReader(file, m.bm, offset)
	filterData := make([]byte, readSize)
	_, err := filterReader.Read(filterData)
	if err != nil {
		return false, err
	}

	bf := BloomFilter.LoadBloomFilter(filterData)

	return bf.IsFound([]byte(key)), nil
}

func (m *SSTableManager) parseData(file *os.File, offset uint64, key string) (*Record, error) {
	dataReader := newBlockReader(file, m.bm, offset)

	var dataHeaderBuf [DATA_HEADER_L]byte
	_, err := dataReader.Read(dataHeaderBuf[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read data header: %v", err)
	}

	header := DeserializeRecordHeader(dataHeaderBuf[:])

	valueBuf := make([]byte, header.ValueSize)
	keyBuf := make([]byte, header.KeySize)
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

	crcHash := crc32.NewIEEE()
	crcHash.Write(dataHeaderBuf[CRC_L:])
	crcHash.Write(keyBuf)
	crcHash.Write(valueBuf)
	realCrc := crcHash.Sum32()
	if header.CRC != realCrc {
		return nil, fmt.Errorf("CRC mismatch: expected %d, got %d", header.CRC, realCrc)
	}

	return &Record{
		Timestamp: header.Timestamp,
		Tombstone: header.Tombstone,
		Key:       readKey,
		Value:     valueBuf,
	}, nil
}

func (sstm *SSTableManager) getMultipleFiles(key string, sstablePath string) (*Record, error) {
	files, err := openMultipleFiles(sstablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open files: %v", err)
	}
	defer files.Close()

	isFound, err := sstm.searchFilter(files.filterFile, 0, 0, key)
	if err != nil {
		return nil, fmt.Errorf("failed to read bloom filter: %v", err)
	}

	// kljuc se ne nalazi u sstable
	if !isFound {
		return nil, nil
	}

	isFound, offset, err := sstm.searchSummary(files.summaryFile, 0, 0, key)
	if err != nil {
		return nil, fmt.Errorf("failed to search summary file: %v", err)
	}
	if !isFound {
		return nil, nil
	}

	offset, err = sstm.searchIndex(files.indexFile, key, offset, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to search index file: %v", err)
	}

	return sstm.parseData(files.dataFile, offset, key)
}

// FIXME: Use the struct for flush as well
type oneFileFooter struct {
	SummaryStart  uint64
	IndexStart    uint64
	DataStart     uint64
	MetadataStart uint64
}

func (m *SSTableManager) readOneFileFooter(file *os.File) (*oneFileFooter, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat SSTable file: %v", err)
	}
	if stat.Size() < FOOTER_L {
		return nil, fmt.Errorf("file size is too small to contain footer")
	}

	offset := uint64(stat.Size() - FOOTER_L)
	reader := newBlockReader(file, m.bm, offset)

	bufferReader := NewBufferReader(FOOTER_L)
	_, err = reader.Read(bufferReader.Buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %v", err)
	}

	footer := &oneFileFooter{
		SummaryStart:  bufferReader.ReadOffset(),
		IndexStart:    bufferReader.ReadOffset(),
		DataStart:     bufferReader.ReadOffset(),
		MetadataStart: bufferReader.ReadOffset(),
	}

	return footer, nil
}

func (m *SSTableManager) getOneFile(key string, sstablePath string) (*Record, error) {
	file, err := os.Open(sstablePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSTable file: %v", err)
	}
	defer file.Close()

	footer, err := m.readOneFileFooter(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %v", err)
	}

	isFound, err := m.searchFilter(file, 0, footer.DataStart, key)
	if err != nil {
		return nil, fmt.Errorf("failed to read bloom filter: %v", err)
	}

	// kljuc se ne nalazi u sstable
	if !isFound {
		return nil, nil
	}

	isFound, offset, err := m.searchSummary(file, footer.SummaryStart, footer.MetadataStart, key)
	if err != nil {
		return nil, fmt.Errorf("failed to search summary file: %v", err)
	}
	if !isFound {
		return nil, nil
	}

	offset, err = m.searchIndex(file, key, offset, footer.SummaryStart)
	if err != nil {
		return nil, fmt.Errorf("failed to search index file: %v", err)
	}

	return m.parseData(file, offset, key)
}
