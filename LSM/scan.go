package sstable

import (
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func (sstm *SSTableManager) PrefixScan(sst *SSTable, prefix string) ([]Record, error) {
	iter, err := sstm.NewPrefixIterator(sst, prefix)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var rec []Record
	for {
		ok, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		rec = append(rec, *iter.iterator.Rec)
	}

	return rec, nil
}

func (sstm *SSTableManager) RangeScan(sst *SSTable, startKey, endKey string) ([]Record, error) {
	iter, err := sstm.NewRangeIterator(sst, startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var rec []Record
	for {
		ok, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		rec = append(rec, *iter.iterator.Rec)
	}

	return rec, nil
}
