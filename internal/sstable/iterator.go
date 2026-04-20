package sstable

import (
	"os"

	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

type SSTableIterator struct {
	Rec        *Record
	br         *blockReader
	sstm       *SSTableManager
	stopOffset uint64
}

func (sstm *SSTableManager) NewSSTableIterator(sst *SSTable) (*SSTableIterator, error) {
	var file *os.File
	var offset uint64
	var err error
	var stopOffset uint64

	if sst.isMultFiles {
		path := sstableFilenameMultFile(sst.path, "Data")
		file, err = os.Open(path)
		if err != nil {
			return nil, err
		}
		offset = uint64(0)
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
		offset = sst.footer.DataStart
		stopOffset = sst.footer.IndexStart
	}

	it := &SSTableIterator{
		br:         newBlockReader(file, sstm.bm, offset),
		sstm:       sstm,
		stopOffset: stopOffset,
	}

	it.Next()
	return it, nil
}

func (it *SSTableIterator) Next() (bool, error) {
	if it.br.CurrOffset() >= it.stopOffset {
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
