package sstable

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
	merkleTree "github.com/Avram-2005/PROJEKAT_NASP/MerkleTree"
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
	. "github.com/Avram-2005/PROJEKAT_NASP/utils"
)

func writeData(writer *blockWriter, record Record) uint64 {
	oldOffset := writer.CurrOffset()
	writer.Write(record.Serialize())
	return oldOffset
}

func writeIndex(writer *blockWriter, key string, offset uint64) uint64 {
	bufferWriter := NewBufferWriter(INDEX_HEADER_L)
	bufferWriter.WriteKeySize(len(key))
	bufferWriter.WriteOffset(offset)

	oldOffset := writer.CurrOffset()
	writer.Write(bufferWriter.Buf)
	writer.Write([]byte(key))
	return oldOffset
}

func writeSummaryHeader(writer *blockWriter, firstKey string, lastKey string) {
	bufferWriter := NewBufferWriter(2 * KEY_SIZE_L)
	bufferWriter.WriteKeySize(len(firstKey))
	bufferWriter.WriteKeySize(len(lastKey))

	writer.Write(bufferWriter.Buf)
	writer.Write([]byte(firstKey))
	writer.Write([]byte(lastKey))
}

type multipleFilesFlushState struct {
	dataWriter     *blockWriter
	indexWriter    *blockWriter
	summaryWriter  *blockWriter
	filterWriter   *blockWriter
	metadataWriter *blockWriter
	bf             *BloomFilter.BloomFilter
	merkleData     [][]byte
}

func (sstm *SSTableManager) multipleFilesFlushInit(tableNum int, numRecs uint) (*multipleFilesFlushState, error) {
	sstablePath := sstm.sstableFilepath(0, tableNum)
	files, err := openMultipleFiles(sstablePath)
	if err != nil {
		return nil, err
	}
	state := &multipleFilesFlushState{
		dataWriter:     newBlockWriter(files.dataFile, sstm.bm),
		indexWriter:    newBlockWriter(files.indexFile, sstm.bm),
		summaryWriter:  newBlockWriter(files.summaryFile, sstm.bm),
		filterWriter:   newBlockWriter(files.filterFile, sstm.bm),
		metadataWriter: newBlockWriter(files.metadataFile, sstm.bm),
		merkleData:     make([][]byte, 0, numRecs),
	}
	state.bf, err = BloomFilter.NewBloomFilter(numRecs, BLOOM_FILTER_RATE)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (sstm *SSTableManager) multipleFilesFlushRecord(record Record, state *multipleFilesFlushState, shouldWriteSummary bool) {
	state.bf.Set([]byte(record.Key)) // dodaj kljuc u filter
	state.merkleData = append(state.merkleData, record.Value)
	offset := writeData(state.dataWriter, record)
	offset = writeIndex(state.indexWriter, record.Key, offset)
	if shouldWriteSummary {
		writeIndex(state.summaryWriter, record.Key, offset)
	}
}

func (sstm *SSTableManager) multipleFilesFlushFinalize(state *multipleFilesFlushState, tableNum int) (*SSTable, error) {
	filterData := state.bf.Dump()
	state.filterWriter.Write(filterData)

	tree, err := merkleTree.NewMerkleTree(state.merkleData)
	if err != nil {
		return nil, err
	}
	serializedTree := tree.Serialize()

	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(len(serializedTree)))
	state.metadataWriter.Write(sizeHeader)

	state.metadataWriter.Write(serializedTree)

	state.dataWriter.Finalize()
	state.indexWriter.Finalize()
	state.summaryWriter.Finalize()
	state.filterWriter.Finalize()
	state.metadataWriter.Finalize()

	return &SSTable{
		path:        sstm.sstableFilepath(0, tableNum),
		size:        state.dataWriter.CurrOffset() + state.indexWriter.CurrOffset() + state.summaryWriter.CurrOffset() + state.filterWriter.CurrOffset(),
		isMultFiles: true,
		footer:      nil,
		filter:      state.bf,
	}, nil
}

func (sstm *SSTableManager) multipleFilesFlush(mem Memtable, tableNum int) (*SSTable, error) {
	entries := mem.GetSortedEntries()

	state, err := sstm.multipleFilesFlushInit(tableNum, uint(len(entries)))
	if err != nil {
		return nil, err
	}

	firstEntry, lastEntry := entries[0], entries[len(entries)-1]
	writeSummaryHeader(state.summaryWriter, firstEntry.Key, lastEntry.Key)

	for i, entry := range entries {
		shouldWriteSummary := i%sstm.config.SummaryInterval == 0
		sstm.multipleFilesFlushRecord(entry, state, shouldWriteSummary)
	}

	return sstm.multipleFilesFlushFinalize(state, tableNum)
}

type indexEntry struct {
	Key    string
	Offset uint64
}

type oneFileFlushState struct {
	writer         *blockWriter
	bf             *BloomFilter.BloomFilter
	merkleData     [][]byte
	index          []indexEntry
	summaryOffsets []indexEntry
}

func (sstm *SSTableManager) oneFileFlushInit(tableNum int, numRecs uint) (*oneFileFlushState, error) {
	sstableFilename := sstm.sstableFilepath(0, tableNum)
	f, err := os.Create(sstableFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSTable file: %v", err)
	}
	writer := newBlockWriter(f, sstm.bm)
	bf, err := BloomFilter.NewBloomFilter(numRecs, BLOOM_FILTER_RATE)
	if err != nil {
		return nil, err
	}
	return &oneFileFlushState{
		writer:         writer,
		bf:             bf,
		index:          make([]indexEntry, 0, numRecs),
		merkleData:     make([][]byte, 0, numRecs),
		summaryOffsets: make([]indexEntry, 0, numRecs/uint(sstm.config.SummaryInterval+1)),
	}, nil
}

func (sstm *SSTableManager) oneFileFlushRecord(entry Record, state *oneFileFlushState) {
	state.bf.Set([]byte(entry.Key)) // dodaj kljuc u filter
	state.merkleData = append(state.merkleData, entry.Value)
	offset := writeData(state.writer, entry)
	state.index = append(state.index, indexEntry{
		Key:    entry.Key,
		Offset: offset,
	})
}

func (sstm *SSTableManager) oneFileFlushFinalize(state *oneFileFlushState, tableNum int) (*SSTable, error) {
	footer := OneFileFooter{}

	footer.FilterStart = state.writer.CurrOffset()
	filterData := state.bf.Dump()
	state.writer.Write(filterData)

	footer.IndexStart = state.writer.CurrOffset()
	summaryOffsets := make([]indexEntry, 0, len(state.index)/sstm.config.SummaryInterval+1)
	for i, entry := range state.index {
		indexOffset := writeIndex(state.writer, entry.Key, entry.Offset)
		if i%sstm.config.SummaryInterval == 0 {
			summaryOffsets = append(summaryOffsets, indexEntry{
				Key:    entry.Key,
				Offset: indexOffset,
			})
		}
	}

	sortedEntries := state.index

	footer.SummaryStart = state.writer.CurrOffset()
	firstEntry, lastEntry := sortedEntries[0], sortedEntries[len(sortedEntries)-1]
	writeSummaryHeader(state.writer, firstEntry.Key, lastEntry.Key)
	for _, entry := range summaryOffsets {
		writeIndex(state.writer, entry.Key, entry.Offset)
	}

	footer.MetadataStart = state.writer.CurrOffset()
	tree, err := merkleTree.NewMerkleTree(state.merkleData)
	if err != nil {
		return nil, err
	}
	serializedTree := tree.Serialize()
	state.writer.Write(serializedTree)

	if state.writer.currBlockNum == 0 && state.writer.currByte == 0 {
		return nil, fmt.Errorf("memtable is empty, no data written")
	}

	footer.Write(state.writer)

	state.writer.Finalize()
	return &SSTable{
		path:        sstm.sstableFilepath(0, tableNum),
		size:        state.writer.CurrOffset(),
		isMultFiles: false,
		footer:      &footer,
		filter:      state.bf,
	}, nil
}

func (sstm *SSTableManager) oneFileFlush(mem Memtable, tableNum int) (*SSTable, error) {
	state, err := sstm.oneFileFlushInit(tableNum, uint(len(mem.GetSortedEntries())))
	if err != nil {
		return nil, err
	}

	for _, entry := range mem.GetSortedEntries() {
		sstm.oneFileFlushRecord(entry, state)
	}

	return sstm.oneFileFlushFinalize(state, tableNum)
}
