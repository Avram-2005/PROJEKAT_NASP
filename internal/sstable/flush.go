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

func writeOneFileFooter(writer *blockWriter, summaryStart uint64, indexStart uint64, dataStart uint64, metadataStart uint64) {
	footrerBuf := NewBufferWriter(FOOTER_L)
	footrerBuf.WriteOffset(summaryStart)
	footrerBuf.WriteOffset(indexStart)
	footrerBuf.WriteOffset(dataStart)
	footrerBuf.WriteOffset(metadataStart)
	writer.Write(footrerBuf.Buf)
}

func (m *SSTableManager) multipleFilesFlush(mem Memtable, tableNum int) error {
	sstablePath := m.sstableFilepath(tableNum)
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

	dataWriter := newBlockWriter(files.dataFile, m.bm)
	indexWriter := newBlockWriter(files.indexFile, m.bm)
	summaryWriter := newBlockWriter(files.summaryFile, m.bm)
	filterWriter := newBlockWriter(files.filterFile, m.bm)
	metadataWriter := newBlockWriter(files.metadataFile, m.bm)

	firstEntry, lastEntry := sortedEntries[0], sortedEntries[len(sortedEntries)-1]
	writeSummaryHeader(summaryWriter, firstEntry.Key, lastEntry.Key)

	for i, entry := range sortedEntries {
		bf.Set([]byte(entry.Key)) // dodaj kljuc u filter
		offset := writeData(dataWriter, entry)
		offset = writeIndex(indexWriter, entry.Key, offset)
		if i%m.config.SummaryInterval == 0 {
			writeIndex(summaryWriter, entry.Key, offset)
		}
	}
	filterWriter.Write(bf.Dump())

	// TODO: Seperate this into a different function
	tree, err := merkleTree.NewMerkleTreeHashes(sortedEntries)
	if err != nil {
		return err
	}
	serializedTree := tree.Serialize()

	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(len(serializedTree)))
	metadataWriter.Write(sizeHeader)

	metadataWriter.Write(serializedTree)

	dataWriter.Finalize()
	indexWriter.Finalize()
	summaryWriter.Finalize()
	filterWriter.Finalize()
	metadataWriter.Finalize()

	return nil
}

type indexEntry struct {
	Key    string
	Offset uint64
}

func (m *SSTableManager) oneFileFlush(mem Memtable, tableNum int) error {
	sstableFilename := m.sstableFilepath(tableNum)
	f, err := os.Create(sstableFilename)
	if err != nil {
		return fmt.Errorf("failed to create SSTable file: %v", err)
	}
	defer f.Close()

	writer := newBlockWriter(f, m.bm)

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
	summaryOffsets := make([]indexEntry, 0, len(index)/m.config.SummaryInterval+1)
	i := 0
	for _, entry := range index {
		indexOffset := writeIndex(writer, entry.Key, entry.Offset)
		if i%m.config.SummaryInterval == 0 {
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

	metadataStart := writer.CurrOffset()
	tree, err := merkleTree.NewMerkleTreeHashes(sortedEntries)
	if err != nil {
		return err
	}
	serializedTree := tree.Serialize()
	writer.Write(serializedTree)

	if writer.currBlockNum == 0 && writer.currByte == 0 {
		return fmt.Errorf("memtable is empty, no data written")
	}

	writeOneFileFooter(writer, summaryStart, indexStart, dataStart, metadataStart)

	writer.Finalize()
	return nil
}
