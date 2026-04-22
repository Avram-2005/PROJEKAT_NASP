package sstable

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

// FIXME: Delete this after config is done
type LSMConfig struct {
	NumLevels      int
	NumFilesLevel0 int
}

type Level struct {
	levelNum int
	size     uint64
	tables   []*SSTable
}

type LSM struct {
	levels []*Level
	config LSMConfig
	sstm   *SSTableManager
}

func NewLSM(lsmConfig LSMConfig, tablesRoot string, sstConfig SSTableConfig, bm *BlockManager.BlockManager) (*LSM, error) {
	sstm, err := SetupSSTableManager(tablesRoot, sstConfig, bm)
	if err != nil {
		return nil, fmt.Errorf("failed to setup SSTableManager: %v", err)
	}
	lsm := LSM{
		levels: make([]*Level, lsmConfig.NumLevels),
		config: lsmConfig,
		sstm:   sstm,
	}
	lsm.levels[0] = &Level{
		levelNum: 0,
		size:     0,
		tables:   []*SSTable{},
	}

	files, err := os.ReadDir(sstm.TablesRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read tables directory: %v", err)
	}

	for _, file := range files {
		sstablePath := filepath.Join(sstm.TablesRoot, file.Name())
		sstable, err := sstm.createSSTable(sstablePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create SSTable from file %s: %v", sstablePath, err)
		}
		levelNum, err := extractLevelNum(file.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to startup LSM: %v", err)
		}
		if lsm.levels[levelNum] == nil {
			lsm.levels[levelNum] = &Level{
				levelNum: levelNum,
				size:     0,
				tables:   []*SSTable{},
			}
		}
		lsm.levels[levelNum].tables = append(lsm.levels[levelNum].tables, sstable)
		lsm.levels[levelNum].size += sstable.size
		lsm.levels[levelNum].levelNum = levelNum
	}

	return &lsm, nil
}

func (l *Level) ShouldCompact() bool {
	return l.size > uint64(math.Pow(10, float64(l.levelNum))*1024*1024)
}

func (l *Level) ShouldCompactL0(numFilesLevel0 int) bool {
	if l.levelNum != 0 {
		return false
	}
	return len(l.tables) > numFilesLevel0
}

func (lsm *LSM) Compact() error {
	return nil
}

func (lsm *LSM) Flush(mem Memtable) error {
	sst, err := lsm.sstm.Flush(mem, len(lsm.levels[0].tables))
	if err != nil {
		return fmt.Errorf("failed to flush memtable: %v", err)
	}
	lsm.levels[0].tables = append(lsm.levels[0].tables, sst)
	lsm.levels[0].size += sst.size
	if lsm.levels[0].ShouldCompactL0(lsm.config.NumFilesLevel0) {
		if err := lsm.Compact(); err != nil {
			return fmt.Errorf("failed to perform compaction: %v", err)
		}
	}
	return nil
}

func (lsm *LSM) Get(key string) ([]byte, error) {
	for _, level := range lsm.levels {
		for _, sstable := range level.tables {
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
