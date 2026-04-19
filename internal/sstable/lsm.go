package sstable

import (
	"fmt"
	"math"

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
	tables   []SSTable
}

type LSM struct {
	levels []Level
	bm     *BlockManager.BlockManager
	config LSMConfig
}

func NewLSM(bm *BlockManager.BlockManager, config LSMConfig) *LSM {
	return &LSM{
		levels: make([]Level, config.NumLevels),
		bm:     bm,
		config: config,
	}
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
	return fmt.Errorf("Compaction not implemented yet")
}

func (lsm *LSM) Flush(mem Memtable) error {
	sst, err := FlushSSTable(mem, len(lsm.levels[0].tables), lsm.bm)
	if err != nil {
		return fmt.Errorf("failed to flush memtable: %v", err)
	}
	lsm.levels[0].tables = append(lsm.levels[0].tables, *sst)
	lsm.levels[0].size += sst.size
	if lsm.levels[0].ShouldCompactL0(lsm.config.NumFilesLevel0) {
		if err := lsm.Compact(); err != nil {
			return fmt.Errorf("failed to perform minor compaction: %v", err)
		}
	}
	return nil
}

func (lsm *LSM) Get(key string) ([]byte, error) {
	return nil, fmt.Errorf("Get not implemented yet")
}
