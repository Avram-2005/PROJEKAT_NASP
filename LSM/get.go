package sstable

import (
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
	. "github.com/Avram-2005/PROJEKAT_NASP/utils"
)

const (
	DATA_HEADER_VARINT_L = CRC_VARINT_MAX_L + TIMESTAMP_VARINT_MAX_L + TOMBSTONE_L + VALUE_SIZE_VARINT_MAX_L
	INDEX_HEADER_L       = 2*KEY_SIZE_VARINT_MAX_L + OFFSET_VARINT_MAX_L
	FOOTER_L             = 5 * OFFSET_L
)

type indexReader struct {
	br      *blockReader
	prevKey string
}

func newIndexReader(file *os.File, bm *BlockManager.BlockManager, offset uint64) *indexReader {
	return &indexReader{
		br:      newBlockReader(file, bm, offset),
		prevKey: "",
	}
}

func (ir *indexReader) Next() (indexEntry, int, error) {
	bufferReader := NewBufferReader(INDEX_HEADER_L)
	n, err := ir.br.Read(bufferReader.Buf)
	if err != nil {
		if err == io.EOF && n == 0 {
			return indexEntry{}, 0, nil
		}
		return indexEntry{}, 0, fmt.Errorf("failed to read index header: %v", err)
	}
	if n == 0 {
		return indexEntry{}, 0, nil
	}
	if n < INDEX_HEADER_L {
		return indexEntry{}, 0, fmt.Errorf("failed to read index header: short read (%d/%d)", n, INDEX_HEADER_L)
	}

	prefixSize, err := bufferReader.ReadKeySizeVarint()
	if err != nil {
		return indexEntry{}, 0, fmt.Errorf("failed to read key size: %v", err)
	}
	suffixSize, err := bufferReader.ReadKeySizeVarint()
	if err != nil {
		return indexEntry{}, 0, fmt.Errorf("failed to read key size: %v", err)
	}
	offset, err := bufferReader.ReadOffsetVarint()
	if err != nil {
		return indexEntry{}, 0, fmt.Errorf("failed to read offset: %v", err)
	}

	bufferReader = NewBufferReader(suffixSize)
	n, err = ir.br.Read(bufferReader.Buf)
	if err != nil {
		return indexEntry{}, 0, fmt.Errorf("failed to read key: %v", err)
	}
	if n == 0 {
		return indexEntry{}, 0, nil
	}
	if n < suffixSize {
		return indexEntry{}, 0, fmt.Errorf("failed to read key: short read (%d/%d)", n, prefixSize)
	}
	suffix := string(bufferReader.Buf[:suffixSize])
	key := ir.prevKey[:prefixSize] + suffix
	ir.prevKey = key

	return indexEntry{
		Key:    key,
		Offset: offset,
	}, n, nil
}

func findExactKeyOffset(searchKey string, reader *indexReader, sectionEnd uint64) (uint64, bool, error) {
	for {
		if sectionEnd > 0 {
			curr := reader.br.CurrOffset()
			if curr >= sectionEnd || sectionEnd-curr < INDEX_HEADER_L {
				return 0, false, nil
			}
		}

		indexEntry, n, err := reader.Next()
		if err != nil {
			return 0, false, err
		}
		if n == 0 {
			return 0, false, nil
		}

		if searchKey == indexEntry.Key {
			return indexEntry.Offset, true, nil
		}
		if searchKey < indexEntry.Key {
			return 0, false, nil
		}
	}
}

func findPreviousKeyOffset(searchKey string, reader *indexReader, sectionEnd uint64) (uint64, error) {
	lastOffset := uint64(0)
	for {
		if sectionEnd > 0 {
			curr := reader.br.CurrOffset()
			if curr >= sectionEnd || sectionEnd-curr < INDEX_HEADER_L {
				return lastOffset, nil
			}
		}

		indexEntry, n, err := reader.Next()
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return lastOffset, nil
		}

		if searchKey < indexEntry.Key {
			return lastOffset, nil
		}

		lastOffset = indexEntry.Offset
	}
}

func (sstm *SSTableManager) searchIndex(file *os.File, key string, oldOffset uint64, sectionEnd uint64) (uint64, bool, error) {
	indexReader := newIndexReader(file, sstm.bm, oldOffset)
	return findExactKeyOffset(key, indexReader, sectionEnd)
}

func readKey(reader *blockReader, keySize int) (string, error) {
	keyBuf := make([]byte, keySize)
	_, err := reader.Read(keyBuf)
	if err != nil {
		return "", fmt.Errorf("failed to read key: %v", err)
	}
	return string(keyBuf), nil
}

// Because the Summary is in memory, this isn't used.
func (sstm *SSTableManager) searchSummary(file *os.File, offset uint64, sectionEnd uint64, key string) (bool, uint64, error) {
	summaryReader := newIndexReader(file, sstm.bm, offset)

	firstKey, lastKey, err := sstm.loadFirstLastSummaryKeys(summaryReader.br)
	if err != nil {
		return false, 0, fmt.Errorf("failed to load first and last summary keys: %v", err)
	}

	if key < firstKey || key > lastKey {
		return false, 0, nil
	}

	offset, err = findPreviousKeyOffset(key, summaryReader, sectionEnd)
	return true, offset, err
}

func (sstm *SSTableManager) loadFirstLastSummaryKeys(reader *blockReader) (string, string, error) {
	bufferReader := NewBufferReader(2 * KEY_SIZE_L)
	_, err := reader.Read(bufferReader.Buf)
	if err != nil {
		return "", "", fmt.Errorf("failed to read summary header: %v", err)
	}

	firstKeySize, err := bufferReader.ReadKeySizeVarint()
	if err != nil {
		return "", "", fmt.Errorf("failed to read first key size: %v", err)
	}
	lastKeySize, err := bufferReader.ReadKeySizeVarint()
	if err != nil {
		return "", "", fmt.Errorf("failed to read last key size: %v", err)
	}

	firstKey, err := readKey(reader, firstKeySize)
	if err != nil {
		return "", "", fmt.Errorf("failed to read first key: %v", err)
	}
	lastKey, err := readKey(reader, lastKeySize)
	if err != nil {
		return "", "", fmt.Errorf("failed to read last key: %v", err)
	}

	return firstKey, lastKey, nil
}

type RecordHeader struct {
	CRC       uint32
	Timestamp time.Time
	Tombstone bool
	ValueSize int
}

func DeserializeRecordHeaderVarInt(data []byte) (*RecordHeader, int, int, error) {
	reader := NewBufferReaderReuse(data)

	crc, err := reader.ReadCRCVarint()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read CRC: %v", err)
	}

	crcStart := reader.CurrOffset()

	timestamp, err := reader.ReadTimestampVarint()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read timestamp: %v", err)
	}
	tombstone := reader.ReadTombstone()
	valueSize, err := reader.ReadValueSizeVarint()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read value size: %v", err)
	}

	return &RecordHeader{
		CRC:       crc,
		Timestamp: timestamp,
		Tombstone: tombstone,
		ValueSize: valueSize,
	}, reader.CurrOffset(), crcStart, nil
}

func (sstm *SSTableManager) parseData(key string, reader *blockReader, checkCRC bool) (*Record, error) {
	offset := reader.CurrOffset()
	var dataHeaderBuf [DATA_HEADER_VARINT_L]byte
	_, err := reader.Read(dataHeaderBuf[:])
	if err != io.EOF && err != nil {
		return nil, fmt.Errorf("failed to read data header: %v", err)
	}

	header, headerLen, crcStart, err := DeserializeRecordHeaderVarInt(dataHeaderBuf[:])
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize record header: %v", err)
	}
	offset += uint64(headerLen)
	reader.Seek(int(offset))

	valueBuf := make([]byte, header.ValueSize)

	_, err = reader.Read(valueBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read value: %v", err)
	}

	if checkCRC {
		crcHash := crc32.NewIEEE()
		crcHash.Write(dataHeaderBuf[crcStart:headerLen])
		crcHash.Write(valueBuf)
		crcHash.Write([]byte(key))
		realCrc := crcHash.Sum32()
		if header.CRC != realCrc {
			return nil, fmt.Errorf("CRC mismatch: expected %d, got %d", header.CRC, realCrc)
		}
	}

	return &Record{
		Timestamp: header.Timestamp,
		Tombstone: header.Tombstone,
		Key:       key,
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

	var offset uint64
	var isFound bool
	if sst.summary != nil {
		isFound, offset, err = sst.summary.IsFound(key)
	} else {
		isFound, offset, err = sstm.searchSummary(files.summaryFile, 0, 0, key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to search summary file: %v", err)
	}
	if !isFound {
		return nil, nil
	}

	offset, isFound, err = sstm.searchIndex(files.indexFile, key, offset, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to search index file: %v", err)
	}
	if !isFound {
		return nil, nil
	}

	dataReader := newBlockReader(files.dataFile, sstm.bm, offset)
	fmt.Printf("Found key %s at offset %d, reading data from table %s\n", key, offset, sst.path)
	return sstm.parseData(key, dataReader, true)
}

func (sstm *SSTableManager) getOneFile(key string, sst *SSTable) (*Record, error) {
	file, err := os.Open(sst.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open SSTable file: %v", err)
	}
	defer file.Close()

	footer, err := sstm.loadOneFileFooter(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %v", err)
	}

	// kljuc se ne nalazi u sstable
	if !sst.filter.IsFound([]byte(key)) {
		return nil, nil
	}

	var offset uint64
	var isFound bool
	if sst.summary != nil {
		isFound, offset, err = sst.summary.IsFound(key)
	} else {
		isFound, offset, err = sstm.searchSummary(file, footer.SummaryStart, footer.MetadataStart, key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to search summary file: %v", err)
	}
	if !isFound {
		fmt.Printf("Key %s not found in summary\n", key)
		return nil, nil
	}

	offset, isFound, err = sstm.searchIndex(file, key, offset, footer.SummaryStart)
	if err != nil {
		return nil, fmt.Errorf("failed to search index file: %v", err)
	}
	if !isFound {
		fmt.Printf("Key %s not found in index\n", key)
		return nil, nil
	}

	dataReader := newBlockReader(file, sstm.bm, offset)
	return sstm.parseData(key, dataReader, true)
}
