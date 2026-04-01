package sstable

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
)

func writeData(writer *blockWriter, entry KeyValue) uint64 {
	bufferWriter := newBufferWriter(DATA_HEADER_L)
	// FIXME: Calculate and write CRC
	bufferWriter.WriteCRC(0)
	bufferWriter.WriteTimestamp()
	bufferWriter.WriteTombstone(entry.Tombstone)
	bufferWriter.WriteKeySize(len(entry.Key))
	bufferWriter.WriteValueSize(len(entry.Value))

	oldOffset := writer.CurrOffset()
	writer.Write(bufferWriter.buf)
	writer.Write([]byte(entry.Key))
	writer.Write(entry.Value)
	return oldOffset
}

func writeIndex(writer *blockWriter, key string, offset uint64) uint64 {
	bufferWriter := newBufferWriter(INDEX_HEADER_L)
	bufferWriter.WriteKeySize(len(key))
	bufferWriter.WriteOffset(offset)

	oldOffset := writer.CurrOffset()
	writer.Write(bufferWriter.buf)
	writer.Write([]byte(key))
	return oldOffset
}

func writeSummaryHeader(writer *blockWriter, firstKey string, lastKey string) {
	bufferWriter := newBufferWriter(2 * KEY_SIZE_L)
	bufferWriter.WriteKeySize(len(firstKey))
	bufferWriter.WriteKeySize(len(lastKey))

	writer.Write(bufferWriter.buf)
	writer.Write([]byte(firstKey))
	writer.Write([]byte(lastKey))
}

func writeOneFileFooter(writer *blockWriter, summaryStart uint64, indexStart uint64, dataStart uint64) {
	var footrerBuf [3 * OFFSET_L]byte
	binary.BigEndian.PutUint64(footrerBuf[0:], summaryStart)
	binary.BigEndian.PutUint64(footrerBuf[OFFSET_L:], indexStart)
	binary.BigEndian.PutUint64(footrerBuf[2*OFFSET_L:], dataStart)
	writer.Write(footrerBuf[:])
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
	defer dataFile.Close()
	defer indexFile.Close()
	defer summaryFile.Close()
	defer filterFile.Close()

	sortedEntries := mem.GetSortedEntries()
	bf, err := BloomFilter.NewBloomFilter(uint(len(sortedEntries)), BLOOM_FILTER_RATE)
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
	Offset uint64
}

func oneFileFlush(mem Memtable, tableNum int, bm *BlockManager.BlockManager) error {
	sstableFilename := sstableFilenameOneFile(tableNum)
	f, err := os.Create(sstableFilename)
	if err != nil {
		return fmt.Errorf("failed to create SSTable file: %v", err)
	}
	defer f.Close()

	writer := newBlockWriter(f, bm)

	bf, err := BloomFilter.NewBloomFilter(uint(len(mem.GetSortedEntries())), BLOOM_FILTER_RATE)
	if err != nil {
		return err
	}

	// TODO: Consider doing this in a single pass
	sortedEntries := mem.GetSortedEntries()
	for _, entry := range sortedEntries {
		bf.Set([]byte(entry.Key))
	}

	filterData := bf.Dump()
	writer.Write(filterData)
	dataStart := writer.CurrOffset()

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
	firstEntry, lastEntry := sortedEntries[0], sortedEntries[len(sortedEntries)-1]
	writeSummaryHeader(writer, firstEntry.Key, lastEntry.Key)
	for _, entry := range summaryOffsets {
		writeIndex(writer, entry.Key, entry.Offset)
	}

	if writer.currBlockNum == 0 && writer.currByte == 0 {
		return fmt.Errorf("memtable is empty, no data written")
	}

	writeOneFileFooter(writer, summaryStart, indexStart, dataStart)

	writer.Finalize()
	return nil
}
