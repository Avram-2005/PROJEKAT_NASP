package snapshot

import (
	"fmt"
	"time"

	sstable "github.com/Avram-2005/PROJEKAT_NASP/LSM"
)

type SnapshotSSTable struct {
	key            string
	sstable        *sstable.SSTable
	sstableManager *sstable.SSTableManager
	timestamp      time.Time
}

func NewSnapshotSSTable(key string, sstable *sstable.SSTable, sstableManager *sstable.SSTableManager) (*SnapshotSSTable, error) {
	rec, err := sstableManager.Get(key, sstable)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, fmt.Errorf("sstable does not contain key")
	}
	return &SnapshotSSTable{
		key:            key,
		sstableManager: sstableManager,
		sstable:        sstable,
		timestamp:      rec.Timestamp,
	}, nil
}

func (sp *SnapshotSSTable) GetTimestamp() time.Time {
	return sp.timestamp
}

func (sp *SnapshotSSTable) GetValue() (*[]byte, error) {
	value, err := sp.sstableManager.Get(sp.key, sp.sstable)
	if err != nil {
		return nil, err
	}
	return &value.Value, nil
}

func (sp *SnapshotSSTable) GetType() string {
	return "sstable"
}
