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

// FIXME: Do this per entry. Consider dumping the data section first
func (sstm *SSTableManager) oneFileFlush(mem Memtable, tableNum int) (*SSTable, error) {
	sstableFilename := sstm.sstableFilepath(0, tableNum)
	f, err := os.Create(sstableFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSTable file: %v", err)
	}
	defer f.Close()
	footer := OneFileFooter{}

	writer := newBlockWriter(f, sstm.bm)

	bf, err := BloomFilter.NewBloomFilter(uint(len(mem.GetSortedEntries())), BLOOM_FILTER_RATE)
	if err != nil {
		return nil, err
	}

	// TODO: Consider doing this in a single pass
	sortedEntries := mem.GetSortedEntries()
	for _, entry := range sortedEntries {
		bf.Set([]byte(entry.Key))
	}

	filterData := bf.Dump()
	writer.Write(filterData)
	footer.DataStart = writer.CurrOffset()

	// FIXME: Do this without copying into a new slice
	var merkleData [][]byte

	index := make([]indexEntry, 0)
	for _, entry := range mem.GetSortedEntries() {
		merkleData = append(merkleData, entry.Value)
		offset := writeData(writer, entry)
		index = append(index, indexEntry{
			Key:    entry.Key,
			Offset: offset,
		})
	}

	footer.IndexStart = writer.CurrOffset()
	summaryOffsets := make([]indexEntry, 0, len(index)/sstm.config.SummaryInterval+1)
	i := 0
	for _, entry := range index {
		indexOffset := writeIndex(writer, entry.Key, entry.Offset)
		if i%sstm.config.SummaryInterval == 0 {
			summaryOffsets = append(summaryOffsets, indexEntry{
				Key:    entry.Key,
				Offset: indexOffset,
			})
		}
		i++
	}

	footer.SummaryStart = writer.CurrOffset()
	firstEntry, lastEntry := sortedEntries[0], sortedEntries[len(sortedEntries)-1]
	writeSummaryHeader(writer, firstEntry.Key, lastEntry.Key)
	for _, entry := range summaryOffsets {
		writeIndex(writer, entry.Key, entry.Offset)
	}

	footer.MetadataStart = writer.CurrOffset()
	tree, err := merkleTree.NewMerkleTree(merkleData)
	if err != nil {
		return nil, err
	}
	serializedTree := tree.Serialize()
	writer.Write(serializedTree)

	if writer.currBlockNum == 0 && writer.currByte == 0 {
		return nil, fmt.Errorf("memtable is empty, no data written")
	}

	footer.Write(writer)

	writer.Finalize()
	return &SSTable{
		path:        sstableFilename,
		size:        writer.CurrOffset(),
		isMultFiles: false,
		footer:      &footer,
		filter:      bf,
	}, nil
}
