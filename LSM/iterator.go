package sstable

import (
	"fmt"
	"os"
	"strings"

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

func openSizedSectionIterator(file *os.File, bm *BlockManager.BlockManager, start uint64) (*sectionIterator, error) {
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

func (sstm *SSTableManager) NewSSTableIterator(sst *SSTable, startKey string, checkCRC bool) (*SSTableIterator, error) {
	it := &SSTableIterator{
		sstm:     sstm,
		checkCRC: checkCRC,
	}

	var err error

	if sst.isMultFiles {
		indexFile, err := os.Open(sstableFilenameMultFile(sst.path, "Index"))
		if err != nil {
			return nil, fmt.Errorf("failed to open index file: %v", err)
		}
		indexOffset, dataOffset := uint64(0), uint64(0)
		if startKey != "" {
			isFound, indexOffset, err := sst.summary.IsFound(startKey)
			if err != nil {
				indexFile.Close()
				return nil, fmt.Errorf("failed to search for start key in summary and index: %v", err)
			}
			if !isFound {
				indexOffset = sst.footer.IndexStart
			}
			restartReader := newIndexReader(indexFile, sstm.bm, indexOffset)
			restartEntry, n, err := restartReader.Next()
			if err != nil {
				indexFile.Close()
				return nil, fmt.Errorf("failed to read restart index entry: %v", err)
			}
			if n > 0 {
				dataOffset = restartEntry.Offset
			}
		}

		it.indexIterator, err = openSizedSectionIterator(indexFile, sstm.bm, indexOffset)
		if err != nil {
			it.Close()
			return nil, err
		}
		it.indexReader = newIndexReader(it.indexIterator.br.file, sstm.bm, indexOffset)

		dataFile, err := os.Open(sstableFilenameMultFile(sst.path, "Data"))
		if err != nil {
			it.Close()
			return nil, fmt.Errorf("failed to open data file: %v", err)
		}
		it.dataIterator, err = openSizedSectionIterator(dataFile, sstm.bm, dataOffset)
		if err != nil {
			return nil, err
		}
	} else {
		file, err := os.Open(sst.path)
		if err != nil {
			return nil, fmt.Errorf("failed to open SSTable file: %v", err)
		}
		indexOffset, dataOffset := sst.footer.IndexStart, uint64(0)
		if startKey != "" {
			isFound, indexOffset, err := sst.summary.IsFound(startKey)
			if err != nil {
				file.Close()
				return nil, fmt.Errorf("failed to search for start key in summary and index: %v", err)
			}
			if !isFound {
				indexOffset = sst.footer.IndexStart
			}
			restartReader := newIndexReader(file, sstm.bm, indexOffset)
			restartEntry, n, err := restartReader.Next()
			if err != nil {
				file.Close()
				return nil, fmt.Errorf("failed to read restart index entry: %v", err)
			}
			if n > 0 {
				dataOffset = restartEntry.Offset
			}
		}

		it.indexIterator = newSectionIterator(file, sstm.bm, indexOffset, sst.footer.SummaryStart)
		it.indexReader = newIndexReader(it.indexIterator.br.file, sstm.bm, indexOffset)

		it.dataIterator = newSectionIterator(file, sstm.bm, dataOffset, sst.footer.IndexStart)
	}

	_, err = it.Next()
	if err != nil {
		it.Close()
		return nil, err
	}

	for startKey != "" && it.Rec != nil && it.Rec.Key < startKey {
		hasNext, err := it.Next()
		if err != nil {
			it.Close()
			return nil, err
		}
		if !hasNext {
			break
		}
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
	if it.dataIterator.br.file == it.indexIterator.br.file {
		if err := it.dataIterator.Close(); err != nil {
			closeErr = fmt.Errorf("failed to close data iterator: %v", err)
		}
	} else {
		if err := it.dataIterator.Close(); err != nil {
			closeErr = fmt.Errorf("failed to close data iterator: %v", err)
		}
		if err := it.indexIterator.Close(); err != nil && closeErr == nil {
			closeErr = fmt.Errorf("failed to close index iterator: %v", err)
		}
	}
	return closeErr
}

type PrefixIterator struct {
	iterator *SSTableIterator
	prefix   string
	started  bool
}

func (sstm *SSTableManager) NewPrefixIterator(sst *SSTable, prefix string) (*PrefixIterator, error) {
	// ukoliko je prefix veci od poslednjeg kljuca ili manji od prvog, znaci da kljuc nije tu
	if prefix != "" {
		if sst.summary.lastKey < prefix {
			return &PrefixIterator{
				iterator: nil,
				prefix:   prefix,
				started:  true,
			}, nil
		}

		if prefix < sst.summary.firstKey && !strings.HasPrefix(sst.summary.firstKey, prefix) {
			return &PrefixIterator{
				iterator: nil,
				prefix:   prefix,
				started:  true,
			}, nil
		}
	}

	iter, err := sstm.NewSSTableIterator(sst, prefix, false)
	if err != nil {
		return nil, err
	}

	return &PrefixIterator{
		iterator: iter,
		prefix:   prefix,
		started:  false,
	}, nil
}
func (it *PrefixIterator) Next() (bool, error) {
	if it.iterator == nil {
		return false, nil
	}

	if it.iterator.Rec == nil {
		return false, nil
	}

	if it.started {
		ok, err := it.iterator.Next()
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	it.started = true

	if strings.HasPrefix(it.iterator.Rec.Key, it.prefix) {
		return true, nil
	}

	if it.prefix != "" && it.iterator.Rec.Key > it.prefix {
		return false, nil
	}
	return true, nil
}

func (it *PrefixIterator) Close() error {
	if it.iterator != nil {
		return it.iterator.Close()
	}
	return nil
}

func (it *PrefixIterator) Stop() {
	it.Close()
	it.iterator = nil
}

type RangeIterator struct {
	iterator *SSTableIterator
	startKey string
	endKey   string
	started  bool
}

func (sstm *SSTableManager) NewRangeIterator(sst *SSTable, startKey, endKey string) (*RangeIterator, error) {

	if sst.summary.lastKey < startKey || endKey < sst.summary.firstKey {
		return &RangeIterator{
			iterator: nil,
			startKey: startKey,
			endKey:   endKey,
			started:  true,
		}, nil
	}

	if endKey < startKey {
		return &RangeIterator{
			iterator: nil,
			startKey: startKey,
			endKey:   endKey,
			started:  true,
		}, nil
	}

	iter, err := sstm.NewSSTableIterator(sst, startKey, false)
	if err != nil {
		return nil, err
	}

	return &RangeIterator{
		iterator: iter,
		startKey: startKey,
		endKey:   endKey,
		started:  false,
	}, nil
}
func (it *RangeIterator) Next() (bool, error) {
	if it.iterator == nil || it.iterator.Rec == nil {
		return false, nil
	}

	if it.started {
		ok, err := it.iterator.Next()
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	it.started = true

	if it.iterator.Rec.Key < it.startKey {
		return false, nil
	}

	if it.iterator.Rec.Key > it.endKey {
		return false, nil
	}

	return true, nil
}

func (it *RangeIterator) Close() error {
	if it.iterator != nil {
		return it.iterator.Close()
	}
	return nil
}

func (it *RangeIterator) Stop() {
	it.Close()
	it.iterator = nil
}
