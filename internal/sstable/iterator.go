package sstable

import (
	"os"
	"strings"

	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

type SSTableIterator struct {
	Rec        *Record
	br         *blockReader
	sstm       *SSTableManager
	sst        *SSTable
	stopOffset uint64
}

func (sstm *SSTableManager) NewSSTableIterator(sst *SSTable) (*SSTableIterator, error) {
	var file *os.File
	var err error
	var stopOffset uint64

	if sst.isMultFiles {
		path := sstableFilenameMultFile(sst.path, "Data")
		file, err = os.Open(path)
		if err != nil {
			return nil, err
		}
		info, err := file.Stat()
		if err != nil {
			return nil, err
		}
		stopOffset = uint64(info.Size())
	} else {
		file, err = os.Open(sst.path)
		if err != nil {
			return nil, err
		}
		stopOffset = sst.footer.IndexStart
	}

	it := &SSTableIterator{
		br:         newBlockReader(file, sstm.bm, 0),
		sstm:       sstm,
		sst:        sst,
		stopOffset: stopOffset,
	}

	it.Next()
	return it, nil
}

func (it *SSTableIterator) Next() (bool, error) {
	if it.br.CurrOffset() >= it.stopOffset {
		it.Rec = nil
		return false, nil
	}
	record, err := it.sstm.parseData(it.br)
	if err != nil {
		return false, err
	}
	it.Rec = record
	return true, nil
}

func (it *SSTableIterator) Close() error {
	return it.br.file.Close()
}

func (it *SSTableIterator) Seek(key string) error {
	var offset uint64
	var err error

	if it.sst.isMultFiles {
		indexPath := sstableFilenameMultFile(it.sst.path, "Index")
		indexFile, err := os.Open(indexPath)
		if err != nil {
			return err
		}
		defer indexFile.Close()

		offset, err = it.sstm.searchIndex(indexFile, key, 0, 0)
		if err != nil {
			return err
		}
	} else {
		footer := it.sst.footer
		offset, err = it.sstm.searchIndex(it.br.file, key, footer.IndexStart, footer.SummaryStart)
		if err != nil {
			return err
		}
	}

	if offset >= it.stopOffset {
		it.Rec = nil
		return nil
	}

	it.br = newBlockReader(it.br.file, it.sstm.bm, offset)
	_, err = it.Next()
	return err
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

	iter, err := sstm.NewSSTableIterator(sst)
	if err != nil {
		return nil, err
	}

	if err := iter.Seek(prefix); err != nil {
		iter.Close()
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

	iter, err := sstm.NewSSTableIterator(sst)
	if err != nil {
		return nil, err
	}

	if err := iter.Seek(startKey); err != nil {
		iter.Close()
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
