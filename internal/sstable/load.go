package sstable

import (
	"fmt"
	"os"

	"github.com/Avram-2005/PROJEKAT_NASP/BloomFilter"
	. "github.com/Avram-2005/PROJEKAT_NASP/utils"
)

func (sstm *SSTableManager) loadSSTable(path string) (*SSTable, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	sst := &SSTable{
		path:        path,
		size:        uint64(info.Size()),
		isMultFiles: info.IsDir(),
	}

	if sst.isMultFiles {
		filter, err := sstm.loadFilterMultFile(sst)
		if err != nil {
			return nil, fmt.Errorf("failed to load bloom filter: %v", err)
		}
		sst.filter = filter

		summary, err := sstm.loadSummaryMultFile(sst)
		if err != nil {
			return nil, fmt.Errorf("failed to load summary: %v", err)
		}
		sst.summary = summary
	} else {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open SSTable file: %v", err)
		}
		defer file.Close()

		footer, err := sstm.loadOneFileFooter(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read SSTable footer: %v", err)
		}
		sst.footer = footer

		filter, err := sstm.loadFilterOneFile(footer, file)
		if err != nil {
			return nil, fmt.Errorf("failed to load bloom filter: %v", err)
		}
		sst.filter = filter

		summary, err := sstm.loadSummaryOneFile(footer, file)
		if err != nil {
			return nil, fmt.Errorf("failed to load summary: %v", err)
		}
		sst.summary = summary
	}

	return sst, nil
}

func (sstm *SSTableManager) loadFilterMultFile(sst *SSTable) (*BloomFilter.BloomFilter, error) {
	path := sstableFilenameMultFile(sst.path, "Filter")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open filter file: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat filter file: %v", err)
	}
	size := uint64(info.Size())
	start := uint64(0)

	return sstm.loadBloomFilter(file, start, size)
}

func (sstm *SSTableManager) loadFilterOneFile(footer *OneFileFooter, file *os.File) (*BloomFilter.BloomFilter, error) {
	size := footer.FooterStart - footer.FilterStart
	start := footer.FilterStart

	return sstm.loadBloomFilter(file, start, size)
}

func (sstm *SSTableManager) loadBloomFilter(file *os.File, start uint64, size uint64) (*BloomFilter.BloomFilter, error) {
	reader := newBlockReader(file, sstm.bm, start)
	data := make([]byte, size)
	_, err := reader.Read(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read bloom filter data: %v", err)
	}

	return BloomFilter.LoadBloomFilter(data), nil
}

func (sstm *SSTableManager) loadSummaryMultFile(sst *SSTable) (*Summary, error) {
	path := sstableFilenameMultFile(sst.path, "Summary")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open summary file: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat summary file: %v", err)
	}
	size := uint64(info.Size())
	start := uint64(0)

	return sstm.loadSummary(file, start, size)
}

func (sstm *SSTableManager) loadSummaryOneFile(footer *OneFileFooter, file *os.File) (*Summary, error) {
	size := footer.MetadataStart - footer.SummaryStart
	start := footer.SummaryStart

	return sstm.loadSummary(file, start, size)
}

func (sstm *SSTableManager) loadSummary(file *os.File, start uint64, size uint64) (*Summary, error) {
	summary := sstm.NewSummary(0)

	reader := newBlockReader(file, sstm.bm, start)
	first, last, err := sstm.loadFirstLastSummaryKeys(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read first and last keys from summary: %v", err)
	}
	summary.SetFirstAndLast(first, last)

	for reader.CurrOffset() < start+size {
		indexEntry, _, err := readNextIndexEntry(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read summary entry: %v", err)
		}
		summary.entries = append(summary.entries, indexEntry)
	}

	return summary, nil
}

func (sstm *SSTableManager) loadOneFileFooter(file *os.File) (*OneFileFooter, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat SSTable file: %v", err)
	}
	if stat.Size() < FOOTER_L {
		return nil, fmt.Errorf("file size is too small to contain footer")
	}

	offset := uint64(stat.Size() - FOOTER_L)
	reader := newBlockReader(file, sstm.bm, offset)

	bufferReader := NewBufferReader(FOOTER_L)
	_, err = reader.Read(bufferReader.Buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read footer: %v", err)
	}

	footer := &OneFileFooter{
		IndexStart:    bufferReader.ReadOffset(),
		SummaryStart:  bufferReader.ReadOffset(),
		MetadataStart: bufferReader.ReadOffset(),
		FilterStart:   bufferReader.ReadOffset(),
		FooterStart:   bufferReader.ReadOffset(),
	}

	return footer, nil
}
