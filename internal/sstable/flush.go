package sstable

import (
	"fmt"
	"os"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
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

func writeOneFileFooter(writer *blockWriter, summaryStart uint64, indexStart uint64, dataStart uint64) {
	footrerBuf := NewBufferWriter(FOOTER_L)
	footrerBuf.WriteOffset(summaryStart)
	footrerBuf.WriteOffset(indexStart)
	footrerBuf.WriteOffset(dataStart)
	writer.Write(footrerBuf.Buf)
}

func multipleFilesFlush(mem Memtable, tableNum int, bm *BlockManager.BlockManager) error {
	sstablePath := sstableFilepath(tableNum)
	files, err := createMultipleFiles(sstablePath)
	if err != nil {
		return err
	}
	defer files.close()

	sortedEntries := mem.GetSortedEntries()
	bf, err := BloomFilter.NewBloomFilter(uint(len(sortedEntries)), BLOOM_FILTER_RATE)
	if err != nil {
		return err
	}

	dataWriter := newBlockWriter(files.dataFile, bm)
	indexWriter := newBlockWriter(files.indexFile, bm)
	summaryWriter := newBlockWriter(files.summaryFile, bm)
	filterWriter := newBlockWriter(files.filterFile, bm)

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
	sstableFilename := sstableFilepath(tableNum)
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
