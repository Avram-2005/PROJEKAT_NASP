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
	indexWriter    *blockWriter
	summaryWriter  *blockWriter
	filterWriter   *blockWriter
	metadataWriter *blockWriter
	bf             *BloomFilter.BloomFilter
	merkleData     [][]byte
	summary        *Summary
}

func (sstm *SSTableManager) multipleFilesFlushInit(level int, tableNum int, numRecs uint) (*multipleFilesFlushState, error) {
	sstablePath := sstm.sstableFilepath(level, tableNum)
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
		summary:        sstm.NewSummary(numRecs),
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
		state.summary.AddEntry(record.Key, offset)
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
		summary:     state.summary,
	}, nil
}

func (sstm *SSTableManager) multipleFilesFlush(mem Memtable, tableNum int) (*SSTable, error) {
	entries := mem.GetSortedEntries()

	state, err := sstm.multipleFilesFlushInit(0, tableNum, uint(len(entries)))
	if err != nil {
		return nil, err
	}

	firstEntry, lastEntry := entries[0], entries[len(entries)-1]
	writeSummaryHeader(state.summaryWriter, firstEntry.Key, lastEntry.Key)
	state.summary.SetFirstAndLast(firstEntry.Key, lastEntry.Key)

	for i, entry := range entries {
		shouldWriteSummary := i%sstm.config.SummaryInterval == 0
		sstm.multipleFilesFlushRecord(entry, state, shouldWriteSummary)
	}

	return sstm.multipleFilesFlushFinalize(state, tableNum)
}

func (sstm *SSTableManager) multipleFilesMerge(ssts []*SSTable, level int, tableNum int) (*SSTable, error) {
	// Cannot calculate number of records in advance, so we set it to 0 for now
	state, err := sstm.multipleFilesFlushInit(level, tableNum, 0)
	if err != nil {
		return nil, err
	}

	var minKey, maxKey string
	for _, sst := range ssts {
		if minKey == "" || sst.summary.firstKey < minKey {
			minKey = sst.summary.firstKey
		}
		if maxKey == "" || sst.summary.lastKey > maxKey {
			maxKey = sst.summary.lastKey
		}
	}
	state.summary.SetFirstAndLast(minKey, maxKey)
	writeSummaryHeader(state.summaryWriter, state.summary)

	iters := make([]*SSTableIterator, len(ssts))
	for i, sst := range ssts {
		iter, err := sstm.NewSSTableIterator(sst)
		if err != nil {
			return nil, fmt.Errorf("failed to create iterator for SSTable %s: %v", sst.path, err)
		}
		iters[i] = iter
	}

	// FIXME: Use a priority queue for better performance when merging many SSTables
	i := 0
	for {
		var minRec *Record
		var minIter *SSTableIterator

		for _, iter := range iters {
			if iter.Rec != nil && !iter.Rec.Tombstone && (minRec == nil || iter.Rec.Key < minRec.Key || (iter.Rec.Key == minRec.Key && iter.Rec.Timestamp.After(minRec.Timestamp))) {
				minRec = iter.Rec
				minIter = iter
			}
		}
		if minRec == nil {
			break
		}
		shouldWriteSummary := i%sstm.config.SummaryInterval == 0
		sstm.multipleFilesFlushRecord(*minRec, state, shouldWriteSummary)
		i++
		if _, err := minIter.Next(); err != nil {
			return nil, fmt.Errorf("failed to advance iterator: %v", err)
		}
	}
	sst, err := sstm.multipleFilesFlushFinalize(state, tableNum)

	bf, err := BloomFilter.NewBloomFilter(uint(i), BLOOM_FILTER_RATE)
	if err != nil {
		return nil, fmt.Errorf("failed to create Bloom filter: %v", err)
	}

	iter, err := sstm.NewSSTableIterator(sst)
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator for new SSTable: %v", err)
	}

	for iter.Rec != nil {
		bf.Set([]byte(iter.Rec.Key))
		if _, err := iter.Next(); err != nil {
			return nil, fmt.Errorf("failed to advance iterator: %v", err)
		}
	}

	sst.filter = bf
	filterData := bf.Dump()
	state.filterWriter.Write(filterData)
	state.filterWriter.Finalize()

	for _, iter := range iters {
		if err := iter.Close(); err != nil {
			return nil, fmt.Errorf("failed to close iterator: %v", err)
		}
	}

	return sst, err
}

type oneFileFlushState struct {
	writer     *blockWriter
	bf         *BloomFilter.BloomFilter
	merkleData [][]byte
	index      []indexEntry
	summary    *Summary
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
		writer:     writer,
		bf:         bf,
		index:      make([]indexEntry, 0, numRecs),
		merkleData: make([][]byte, 0, numRecs),
		summary:    sstm.NewSummary(numRecs),
	}, nil
}

func (sstm *SSTableManager) oneFileFlushRecord(i int, entry Record, state *oneFileFlushState) {
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

	footer.IndexStart = state.writer.CurrOffset()
	for i, entry := range state.index {
		indexOffset := writeIndex(state.writer, entry.Key, entry.Offset)
		if i%sstm.config.SummaryInterval == 0 {
			state.summary.AddEntry(entry.Key, indexOffset)
		}
	}

	footer.SummaryStart = state.writer.CurrOffset()
	firstEntry, lastEntry := state.index[0], state.index[len(state.index)-1]
	writeSummaryHeader(state.writer, firstEntry.Key, lastEntry.Key)
	state.summary.SetFirstAndLast(firstEntry.Key, lastEntry.Key)
	for _, entry := range state.summary.entries {
		writeIndex(state.writer, entry.Key, entry.Offset)
	}

	footer.MetadataStart = state.writer.CurrOffset()
	tree, err := merkleTree.NewMerkleTree(state.merkleData)
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
		path:        sstm.sstableFilepath(0, tableNum),
		size:        state.writer.CurrOffset(),
		isMultFiles: false,
		footer:      &footer,
		filter:      state.bf,
		summary:     state.summary,
	}, nil
}

func (sstm *SSTableManager) oneFileFlush(mem Memtable, tableNum int) (*SSTable, error) {
	state, err := sstm.oneFileFlushInit(tableNum, uint(len(mem.GetSortedEntries())))
	if err != nil {
		return nil, err
	}

	for i, entry := range mem.GetSortedEntries() {
		sstm.oneFileFlushRecord(i, entry, state)
	}

	return sstm.oneFileFlushFinalize(state, tableNum)
}

func (sstm *SSTableManager) oneFileMerge(ssts []*SSTable, level int, tableNum int) (*SSTable, error) {
	return nil, fmt.Errorf("multiple files merge not implemented yet")
}
