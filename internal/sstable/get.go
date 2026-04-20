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

func (sstm *SSTableManager) searchIndex(file *os.File, key string, oldOffset uint64, sectionEnd uint64) (uint64, error) {
	indexReader := newBlockReader(file, sstm.bm, oldOffset)
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

func (sstm *SSTableManager) searchSummary(file *os.File, offset uint64, sectionEnd uint64, key string) (bool, uint64, error) {
	summaryReader := newBlockReader(file, sstm.bm, offset)

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

func (sstm *SSTableManager) GetFilter(file *os.File, offset uint64, readSize uint64) (*BloomFilter.BloomFilter, error) {
	filterReader := newBlockReader(file, sstm.bm, offset)
	filterData := make([]byte, readSize)
	_, err := filterReader.Read(filterData)
	if err != nil {
		return nil, err
	}

	bf := BloomFilter.LoadBloomFilter(filterData)

	return bf, nil
}

func (sstm *SSTableManager) parseData(reader *blockReader) (*Record, error) {
	var dataHeaderBuf [DATA_HEADER_L]byte
	_, err := reader.Read(dataHeaderBuf[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read data header: %v", err)
	}

	header := DeserializeRecordHeader(dataHeaderBuf[:])

	valueBuf := make([]byte, header.ValueSize)
	keyBuf := make([]byte, header.KeySize)
	_, err = reader.Read(keyBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %v", err)
	}
	readKey := string(keyBuf)

	_, err = reader.Read(valueBuf)
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

func (sstm *SSTableManager) getMultipleFiles(key string, sst *SSTable) (*Record, error) {
	files, err := openMultipleFiles(sst.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open files: %v", err)
	}
	defer files.Close()

	// kljuc se ne nalazi u sstable
	if !sst.filter.IsFound([]byte(key)) {
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

	dataReader := newBlockReader(files.dataFile, sstm.bm, offset)
	return sstm.parseData(dataReader)
}

func (sstm *SSTableManager) getOneFile(key string, sst *SSTable) (*Record, error) {
	file, err := os.Open(sst.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSTable file: %v", err)
	}
	defer file.Close()

	footer, err := sstm.GetOneFileFooter(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %v", err)
	}

	// kljuc se ne nalazi u sstable
	if !sst.filter.IsFound([]byte(key)) {
		return nil, nil
	}

	isFound, offset, err := sstm.searchSummary(file, footer.SummaryStart, footer.MetadataStart, key)
	if err != nil {
		return nil, fmt.Errorf("failed to search summary file: %v", err)
	}
	if !isFound {
		return nil, nil
	}

	offset, err = sstm.searchIndex(file, key, offset, footer.SummaryStart)
	if err != nil {
		return nil, fmt.Errorf("failed to search index file: %v", err)
	}

	dataReader := newBlockReader(file, sstm.bm, offset)
	return sstm.parseData(dataReader)
}
