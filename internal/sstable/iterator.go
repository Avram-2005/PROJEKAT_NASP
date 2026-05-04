package sstable

import (
	"fmt"
	"os"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

type sectionIterator struct {
	br         *blockReader
	start      uint64
	stopOffset uint64
}

func newSectionIterator(file *os.File, bm *BlockManager.BlockManager, start uint64, stopOffset uint64) *sectionIterator {
	return &sectionIterator{
		br:         newBlockReader(file, bm, start),
		start:      start,
		stopOffset: stopOffset,
	}
}

func openSizedSectionIterator(path string, bm *BlockManager.BlockManager, start uint64) (*sectionIterator, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	return newSectionIterator(file, bm, start, uint64(info.Size())), nil
}

func openRangedSectionIterator(path string, bm *BlockManager.BlockManager, start uint64, stopOffset uint64) (*sectionIterator, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return newSectionIterator(file, bm, start, stopOffset), nil
}

func (it *sectionIterator) hasNext() bool {
	return it.br.CurrOffset() < it.stopOffset
}

func (it *sectionIterator) Close() error {
	if it == nil || it.br == nil {
		return nil
	}
	return it.br.Close()
}

type SSTableIterator struct {
	Rec           *Record
	sstm          *SSTableManager
	dataIterator  *sectionIterator
	indexIterator *sectionIterator
	indexReader   *indexReader
	checkCRC      bool
}

func (sstm *SSTableManager) NewSSTableIterator(sst *SSTable, checkCRC bool) (*SSTableIterator, error) {
	it := &SSTableIterator{
		sstm:     sstm,
		checkCRC: checkCRC,
	}

	var err error

	if sst.isMultFiles {
		it.dataIterator, err = openSizedSectionIterator(sstableFilenameMultFile(sst.path, "Data"), sstm.bm, 0)
		if err != nil {
			return nil, err
		}
		it.indexIterator, err = openSizedSectionIterator(sstableFilenameMultFile(sst.path, "Index"), sstm.bm, 0)
		if err != nil {
			it.Close()
			return nil, err
		}
		it.indexReader = newIndexReader(it.indexIterator.br.file, sstm.bm, 0)
	} else {
		if sst.footer == nil {
			it.Close()
			return nil, fmt.Errorf("missing SSTable footer for one-file iterator")
		}

		it.dataIterator, err = openRangedSectionIterator(sst.path, sstm.bm, 0, sst.footer.IndexStart)
		if err != nil {
			it.Close()
			return nil, err
		}
		it.indexIterator, err = openRangedSectionIterator(sst.path, sstm.bm, sst.footer.IndexStart, sst.footer.SummaryStart)
		if err != nil {
			it.Close()
			return nil, err
		}
		it.indexReader = newIndexReader(it.indexIterator.br.file, sstm.bm, sst.footer.IndexStart)
	}

	_, err = it.Next()
	if err != nil {
		it.Close()
		return nil, err
	}
	return it, nil
}

func (it *SSTableIterator) Next() (bool, error) {
	indexCurrOffset := it.indexReader.br.CurrOffset()
	if indexCurrOffset >= it.indexIterator.stopOffset || it.indexIterator.stopOffset-indexCurrOffset < INDEX_HEADER_L {
		it.Rec = nil
		return false, nil
	}
	entry, n, err := it.indexReader.Next()
	if err != nil {
		return false, err
	}
	if n == 0 {
		it.Rec = nil
		return false, nil
	}

	if !it.dataIterator.hasNext() {
		return false, fmt.Errorf("data section ended before index section")
	}
	record, err := it.sstm.parseData(entry.Key, it.dataIterator.br, it.checkCRC)
	if err != nil {
		return false, err
	}
	it.Rec = record
	return true, nil
}

func (it *SSTableIterator) Close() error {
	var closeErr error
	if err := it.dataIterator.Close(); err != nil {
		closeErr = fmt.Errorf("failed to close data iterator: %v", err)
	}
	if err := it.indexIterator.Close(); err != nil && closeErr == nil {
		closeErr = fmt.Errorf("failed to close index iterator: %v", err)
	}
	return closeErr
}
