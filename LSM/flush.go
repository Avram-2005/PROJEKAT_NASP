package sstable

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"

	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
	merkleTree "github.com/Avram-2005/PROJEKAT_NASP/MerkleTree"
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
	. "github.com/Avram-2005/PROJEKAT_NASP/utils"
)

func serializeRecord(r *Record) []byte {
	value := r.Value

	payloadWriter := NewBufferWriter(3*binary.MaxVarintLen64 + TOMBSTONE_L + len(value))
	payloadLen := 0
	payloadLen += payloadWriter.WriteTimestampVarint(r.Timestamp)
	payloadWriter.WriteTombstone(r.Tombstone)
	payloadLen += TOMBSTONE_L
	payloadLen += payloadWriter.WriteValueSizeVarint(len(value))
	payloadWriter.WriteBytes(value)
	payloadLen += len(value)

	payload := payloadWriter.Buf[:payloadLen]
	crcHash := crc32.NewIEEE()
	crcHash.Write(payload)
	crcHash.Write([]byte(r.Key))

	writer := NewBufferWriter(binary.MaxVarintLen32 + payloadLen)
	totalLen := writer.WriteCRCVarint(crcHash.Sum32())
	writer.WriteBytes(payload)
	totalLen += payloadLen
	return writer.Buf[:totalLen]
}

func writeData(writer *blockWriter, record *Record) uint64 {
	oldOffset := writer.CurrOffset()
	writer.Write(serializeRecord(record))
	return oldOffset
}

type indexWriter struct {
	bw      *blockWriter
	prevKey string
}

func newIndexWriter(bw *blockWriter) *indexWriter {
	return &indexWriter{
		bw:      bw,
		prevKey: "",
	}
}

func commonPrefixLength(a, b string) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return minLen
}

func (iw *indexWriter) Write(key string, offset uint64) uint64 {
	prefixSize := 0
	if iw.prevKey != "" {
		prefixSize = commonPrefixLength(iw.prevKey, key)
	}
	suffixSize := len(key) - prefixSize
	bufferWriter := NewBufferWriter(INDEX_HEADER_L)
	bufferWriter.WriteKeySizeVarint(prefixSize)
	bufferWriter.WriteKeySizeVarint(suffixSize)
	bufferWriter.WriteOffsetVarint(offset)

	oldOffset := iw.bw.CurrOffset()
	iw.bw.Write(bufferWriter.Buf)
	iw.bw.Write([]byte(key[prefixSize:]))
	iw.prevKey = key
	return oldOffset
}

func writeSummaryHeader(writer *blockWriter, firstKey string, lastKey string) {
	bufferWriter := NewBufferWriter(2 * KEY_SIZE_L)
	bufferWriter.WriteKeySizeVarint(len(firstKey))
	bufferWriter.WriteKeySizeVarint(len(lastKey))

	writer.Write(bufferWriter.Buf)
	writer.Write([]byte(firstKey))
	writer.Write([]byte(lastKey))
}

func (off *OneFileFooter) Write(writer *blockWriter) {
	footrerBuf := NewBufferWriter(FOOTER_L)
	footrerBuf.WriteOffset(off.IndexStart)
	footrerBuf.WriteOffset(off.SummaryStart)
	footrerBuf.WriteOffset(off.MetadataStart)
	footrerBuf.WriteOffset(off.FilterStart)
	footrerBuf.WriteOffset(off.FooterStart)
	writer.Write(footrerBuf.Buf)
}

type multipleFilesFlushState struct {
	dataWriter     *blockWriter
	indexWriter    *indexWriter
	summaryWriter  *indexWriter
	filterWriter   *blockWriter
	metadataWriter *blockWriter
	bf             *BloomFilter.BloomFilter
	merkleData     []*Record
	summary        *Summary
	files          *sstableFiles
}

func (sstm *SSTableManager) multipleFilesFlushInit(level int, tableNum int, numRecs uint) (*multipleFilesFlushState, error) {
	sstablePath := sstm.sstableFilepath(level, tableNum)
	files, err := openMultipleFiles(sstablePath)
	if err != nil {
		return nil, err
	}
	state := &multipleFilesFlushState{
		dataWriter:     newBlockWriter(files.dataFile, sstm.bm),
		indexWriter:    newIndexWriter(newBlockWriter(files.indexFile, sstm.bm)),
		summaryWriter:  newIndexWriter(newBlockWriter(files.summaryFile, sstm.bm)),
		filterWriter:   newBlockWriter(files.filterFile, sstm.bm),
		metadataWriter: newBlockWriter(files.metadataFile, sstm.bm),
		merkleData:     make([]*Record, 0, numRecs),
		summary:        sstm.NewSummary(numRecs),
		files:          files,
	}
	state.bf, err = BloomFilter.NewBloomFilter(numRecs, BLOOM_FILTER_RATE)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (sstm *SSTableManager) multipleFilesFlushRecord(record *Record, state *multipleFilesFlushState, shouldWriteSummary bool) {
	state.bf.Set([]byte(record.Key)) // dodaj kljuc u filter
	state.merkleData = append(state.merkleData, record)
	offset := writeData(state.dataWriter, record)
	if shouldWriteSummary {
		state.indexWriter.prevKey = ""
	}
	offset = state.indexWriter.Write(record.Key, offset)
	if shouldWriteSummary {
		state.summaryWriter.Write(record.Key, offset)
		state.summary.AddEntry(record.Key, offset)
	}
}

func (sstm *SSTableManager) multipleFilesFlushFinalize(level int, state *multipleFilesFlushState, tableNum int) (*SSTable, error) {
	filterData := state.bf.Dump()
	state.filterWriter.Write(filterData)

	tree, err := merkleTree.NewMerkleTreeHashes(state.merkleData)
	if err != nil {
		return nil, err
	}
	serializedTree := tree.Serialize()

	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(len(serializedTree)))
	state.metadataWriter.Write(sizeHeader)

	state.metadataWriter.Write(serializedTree)

	state.dataWriter.Finalize()
	state.indexWriter.bw.Finalize()
	state.summaryWriter.bw.Finalize()
	state.filterWriter.Finalize()
	state.metadataWriter.Finalize()

	return &SSTable{
		path:        sstm.sstableFilepath(level, tableNum),
		size:        state.dataWriter.CurrOffset() + state.indexWriter.bw.CurrOffset() + state.summaryWriter.bw.CurrOffset() + state.filterWriter.CurrOffset(),
		isMultFiles: true,
		footer:      nil,
		filter:      state.bf,
		summary:     state.summary,
	}, nil
}

func (sstm *SSTableManager) multipleFilesFlush(entries []*Record, tableNum int) (*SSTable, error) {
	state, err := sstm.multipleFilesFlushInit(0, tableNum, uint(len(entries)))
	if err != nil {
		return nil, err
	}
	defer state.files.Close()

	firstEntry, lastEntry := entries[0], entries[len(entries)-1]
	writeSummaryHeader(state.summaryWriter.bw, firstEntry.Key, lastEntry.Key)
	state.summary.SetFirstAndLast(firstEntry.Key, lastEntry.Key)

	for i, entry := range entries {
		shouldWriteSummary := i%sstm.config.SummaryInterval == 0
		sstm.multipleFilesFlushRecord(entry, state, shouldWriteSummary)
	}

	return sstm.multipleFilesFlushFinalize(0, state, tableNum)
}

type oneFileFlushState struct {
	writer        *blockWriter
	bf            *BloomFilter.BloomFilter
	merkleData    []*Record
	index         []indexEntry
	summary       *Summary
	indexWriter   *indexWriter
	summaryWriter *indexWriter
	file          *os.File
}

func (sstm *SSTableManager) oneFileFlushInit(level int, tableNum int, numRecs uint) (*oneFileFlushState, error) {
	sstableFilename := sstm.sstableFilepath(level, tableNum)
	file, err := os.Create(sstableFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSTable file: %v", err)
	}
	writer := newBlockWriter(file, sstm.bm)
	bf, err := BloomFilter.NewBloomFilter(numRecs, BLOOM_FILTER_RATE)
	if err != nil {
		return nil, err
	}
	return &oneFileFlushState{
		writer:        writer,
		bf:            bf,
		index:         make([]indexEntry, 0, numRecs),
		merkleData:    make([]*Record, 0, numRecs),
		summary:       sstm.NewSummary(numRecs),
		file:          file,
		indexWriter:   newIndexWriter(writer),
		summaryWriter: newIndexWriter(writer),
	}, nil
}

func (sstm *SSTableManager) oneFileFlushRecord(i int, entry *Record, state *oneFileFlushState) {
	state.bf.Set([]byte(entry.Key)) // dodaj kljuc u filter
	state.merkleData = append(state.merkleData, entry)
	offset := writeData(state.writer, entry)
	state.index = append(state.index, indexEntry{
		Key:    entry.Key,
		Offset: offset,
	})
}

func (sstm *SSTableManager) oneFileFlushFinalize(level int, state *oneFileFlushState, tableNum int) (*SSTable, error) {
	footer := OneFileFooter{}

	footer.IndexStart = state.writer.CurrOffset()
	for i, entry := range state.index {
		if i%sstm.config.SummaryInterval == 0 {
			state.indexWriter.prevKey = ""
		}
		indexOffset := state.indexWriter.Write(entry.Key, entry.Offset)
		if i%sstm.config.SummaryInterval == 0 {
			state.summary.AddEntry(entry.Key, indexOffset)
		}
	}

	footer.SummaryStart = state.writer.CurrOffset()
	firstEntry, lastEntry := state.index[0], state.index[len(state.index)-1]
	writeSummaryHeader(state.writer, firstEntry.Key, lastEntry.Key)
	state.summary.SetFirstAndLast(firstEntry.Key, lastEntry.Key)
	for _, entry := range state.summary.entries {
		state.summaryWriter.Write(entry.Key, entry.Offset)
	}

	footer.MetadataStart = state.writer.CurrOffset()
	tree, err := merkleTree.NewMerkleTreeHashes(state.merkleData)
	if err != nil {
		return nil, err
	}
	serializedTree := tree.Serialize()
	state.writer.Write(serializedTree)

	footer.FilterStart = state.writer.CurrOffset()
	filterData := state.bf.Dump()
	state.writer.Write(filterData)

	if state.writer.currBlockNum == 0 && state.writer.currByte == 0 {
		return nil, fmt.Errorf("memtable is empty, no data written")
	}

	footer.FooterStart = state.writer.CurrOffset()
	footer.Write(state.writer)

	state.writer.Finalize()
	return &SSTable{
		path:        sstm.sstableFilepath(level, tableNum),
		size:        state.writer.CurrOffset(),
		isMultFiles: false,
		footer:      &footer,
		filter:      state.bf,
		summary:     state.summary,
	}, nil
}

func (sstm *SSTableManager) oneFileFlush(entries []*Record, tableNum int) (*SSTable, error) {
	state, err := sstm.oneFileFlushInit(0, tableNum, uint(len(entries)))
	if err != nil {
		return nil, err
	}
	defer state.file.Close()

	for i, entry := range entries {
		sstm.oneFileFlushRecord(i, entry, state)
	}

	return sstm.oneFileFlushFinalize(0, state, tableNum)
}
