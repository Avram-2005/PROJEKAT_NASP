package sstable

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

// FIXME: Delete this after config is done
type LSMConfig struct {
	NumLevels        int
	CompactionFactor int
}

type Level struct {
	levelNum int
	size     uint64
	tables   []*SSTable
}

type LSM struct {
	levels []Level
	config LSMConfig
	sstm   *SSTableManager
}

func NewLSM(lsmConfig LSMConfig, tablesRoot string, sstConfig SSTableConfig, bm *BlockManager.BlockManager) (*LSM, error) {
	sstm, err := SetupSSTableManager(tablesRoot, sstConfig, bm)
	if err != nil {
		return nil, fmt.Errorf("failed to setup SSTableManager: %v", err)
	}
	lsm := LSM{
		levels: make([]Level, lsmConfig.NumLevels),
		config: lsmConfig,
		sstm:   sstm,
	}

	files, err := os.ReadDir(sstm.TablesRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read tables directory: %v", err)
	}

	for _, file := range files {
		sstablePath := filepath.Join(sstm.TablesRoot, file.Name())
		sstable, err := sstm.loadSSTable(sstablePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create SSTable from file %s: %v", sstablePath, err)
		}
		levelNum, err := extractLevelNum(file.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to startup LSM: %v", err)
		}
		lsm.levels[levelNum].tables = append(lsm.levels[levelNum].tables, sstable)
		lsm.levels[levelNum].size += sstable.size
		lsm.levels[levelNum].levelNum = levelNum
	}

	return &lsm, nil
}

func (l *Level) ShouldCompact(compFactor int) bool {
	return len(l.tables) >= compFactor
}

func (lsm *LSM) Compact(levelNum int) error {
	shouldDeleteTombstone := levelNum == lsm.config.NumLevels-1
	newSST, err := lsm.sstm.Merge(lsm.levels[levelNum].tables, levelNum+1, shouldDeleteTombstone)
	if err != nil {
		return fmt.Errorf("failed to merge SSTables for compaction: %v", err)
	}
	err = lsm.levels[levelNum].delete()
	if err != nil {
		return fmt.Errorf("failed to delete old SSTables after compaction: %v", err)
	}
	lsm.levels[levelNum+1].tables = append(lsm.levels[levelNum+1].tables, newSST)
	lsm.levels[levelNum+1].size += newSST.size
	lsm.levels[levelNum+1].levelNum = levelNum + 1
	if levelNum < lsm.config.NumLevels-1 && lsm.levels[levelNum+1].ShouldCompact(lsm.config.CompactionFactor) {
		if err := lsm.Compact(levelNum + 1); err != nil {
			return fmt.Errorf("failed to compact level %d: %v", levelNum+1, err)
		}
	}
	return nil
}

func (l *Level) delete() error {
	for _, sst := range l.tables {
		if err := os.RemoveAll(sst.path); err != nil {
			return fmt.Errorf("failed to delete SSTable file %s: %v", sst.path, err)
		}
	}
	l.tables = nil
	l.size = 0
	return nil
}

func (lsm *LSM) Flush(records []*Record) error {
	sst, err := lsm.sstm.Flush(records)
	if err != nil {
		return fmt.Errorf("failed to flush memtable: %v", err)
	}
	lsm.levels[0].tables = append(lsm.levels[0].tables, sst)
	lsm.levels[0].size += sst.size
	if lsm.levels[0].ShouldCompact(lsm.config.CompactionFactor) {
		if err := lsm.Compact(0); err != nil {
			return fmt.Errorf("failed to compact level 0: %v", err)
		}
	}
	return nil
}

func (lsm *LSM) Get(key string) ([]byte, error) {
	for _, level := range lsm.levels {
		for i := len(level.tables) - 1; i >= 0; i-- {
			sstable := level.tables[i]
			rec, err := lsm.sstm.Get(key, sstable)
			if err != nil {
				return nil, fmt.Errorf("error getting key from SSTable %s: %v", sstable.path, err)
			}
			if rec != nil {
				if rec.Tombstone {
					return nil, fmt.Errorf("key %s has been deleted", key)
				}
				return rec.Value, nil
			}
		}
	}

	return nil, fmt.Errorf("key %s not found in any SSTable", key)
}

func (lsm *LSM) GetNewestRecord() (*Record, error) {
	var newstSST *SSTable
	for _, level := range lsm.levels {
		if len(level.tables) > 0 {
			newstSST = level.tables[len(level.tables)-1]
			break
		}
	}
	if newstSST == nil {
		return nil, nil
	}

	iter, err := lsm.sstm.NewSSTableIterator(newstSST, "", true)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSTable iterator: %v", err)
	}
	defer iter.Close()

	var newest *Record
	for {
		ok, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		if newest == nil || iter.Rec.Timestamp.After(newest.Timestamp) {
			newest = iter.Rec
		}
	}

	return newest, nil
}

// skenira sve nivoe LSM stabla za zadati prefix
func (lsm *LSM) PrefixScan(prefix string) ([]*Record, error) {
	results := make([]*Record, 0)
	for _, level := range lsm.levels {
		for _, sstab := range level.tables {
			if sstab.summary == nil || len(sstab.summary.entries) == 0 {
				continue
			}
			if sstab.filter == nil {
				continue
			}
			if sstab.summary.lastKey < prefix { //ako je poslednji kljuc manji od prefixa, sigurno ga nema u datom nivou
				continue
			}
			if sstab.summary.firstKey > prefix && !hasPrefix(sstab.summary.firstKey, prefix) {
				continue
			}
			sstabRecords, err := lsm.sstm.PrefixScan(sstab, prefix)
			if err != nil {
				return nil, err
			}
			for i := range sstabRecords {
				results = append(results, &sstabRecords[i])
			}
		}
	}
	return results, nil
}

// helper funkcija
func hasPrefix(key, prefix string) bool {
	if len(key) < len(prefix) {
		return false
	}
	return key[:len(prefix)] == prefix
}

// skenira sve nivoe LSM stala za zadati opseg
func (lsm *LSM) RangeScan(startKey, endKey string) ([]*Record, error) {
	results := make([]*Record, 0)
	for _, level := range lsm.levels {
		for _, sstab := range level.tables {
			if sstab.summary.lastKey < startKey || sstab.summary.firstKey > endKey { //provera preklapanja opsega, ako se ne preklapaju preskace
				continue
			}
			sstabRecords, err := lsm.sstm.RangeScan(sstab, startKey, endKey)
			if err != nil {
				return nil, err
			}
			for i := range sstabRecords {
				results = append(results, &sstabRecords[i])
			}
		}
	}
	return results, nil
}

type SSTableInfo struct {
	Level int
	Path  string
	Table *SSTable
}

func (lsm *LSM) GetAllSSTables() []SSTableInfo {
	var result []SSTableInfo
	for _, level := range lsm.levels {
		for _, sst := range level.tables {
			result = append(result, SSTableInfo{
				Level: level.levelNum,
				Path:  sst.path,
				Table: sst,
			})
		}
	}
	return result
}

func (lsm *LSM) ValidateSSTable(sst *SSTable) (bool, []Record, error) {
	return lsm.sstm.ValidateSSTable(sst)
}
