package sstable

import (
	"fmt"
	"math"
)

// FIXME: Delete this after config is done
var isLevelTiered bool

type Level struct {
	levelNum int
	size     int64
	tables   []SSTable
}

// TODO: 1.4[DZ1] Size-Tiered Compaction
type LSM struct {
	levels []Level
}

func (l *Level) ShouldCompact() bool {
	return l.size > int64(math.Pow(10, float64(l.levelNum))*1024*1024)
}

func (lsm *LSM) Compact() error {
	return fmt.Errorf("Compaction not implemented yet")
}

func (lsm *LSM) Flush(mem Memtable) error {
	return fmt.Errorf("Flush not implemented yet")
}

func (lsm *LSM) Get(key string) ([]byte, error) {
	return nil, fmt.Errorf("Get not implemented yet")
}
