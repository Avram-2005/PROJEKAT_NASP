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
	newSST, err := lsm.sstm.Merge(lsm.levels[levelNum].tables, levelNum+1)
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
	for levelNum, level := range lsm.levels[:len(lsm.levels)-1] {
		if !level.ShouldCompact(lsm.config.CompactionFactor) {
			break
		}
		if err := lsm.Compact(levelNum); err != nil {
			return fmt.Errorf("failed to compact level %d: %v", level.levelNum, err)
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
		return nil, fmt.Errorf("no SSTables found in LSM")
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
