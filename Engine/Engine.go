package engine

import (
	blockmanager "github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	cache "github.com/Avram-2005/PROJEKAT_NASP/Cache"
	configuration "github.com/Avram-2005/PROJEKAT_NASP/Config"
	lsm "github.com/Avram-2005/PROJEKAT_NASP/LSM"
	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
	wal "github.com/Avram-2005/PROJEKAT_NASP/WAL"
)

type Engine struct {
	configuration configuration.Config
	cache         cache.Cache
	memtable      memtable.MemtableManager
	lsmTree       lsm.LSM
	writeAheadLog wal.WAL
	blockManager  blockmanager.BlockManager
}

func (engine *Engine) NewEngine() (*Engine, error) {
	return nil, nil
}

func (engine *Engine) WritePath(key string, value []byte) error {
	return nil
}

func (engine *Engine) ReadPath(key string) ([]byte, error) {
	return nil, nil
}

func (engine *Engine) Put(key string, value []byte) error {
	return nil
}

func (engin *Engine) Get(key string) ([]byte, error) {
	return nil, nil
}

func (engine *Engine) Delete(key string) error {
	return nil
}

func (engine *Engine) PrefixScan(prefix string) *[]record.Record {
	return nil
}

func (engine *Engine) RangeScan(startKey, endKey string) *[]record.Record {
	return nil
}
