package snapshot

import (
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

type SnapshotMemtable struct {
	value     *[]byte
	timestamp time.Time
}

func NewSnapshotMemtable(value *[]byte, timestamp time.Time) *SnapshotMemtable {
	return &SnapshotMemtable{
		value:     value,
		timestamp: timestamp,
	}
}

func (sp *SnapshotMemtable) GetTimestamp() time.Time {
	return sp.timestamp
}

func (sp *SnapshotMemtable) GetValue(bm *BlockManager.BlockManager) (*[]byte, error) {
	return sp.value, nil
}

func (sp *SnapshotMemtable) GetType() string {
	return "memtable"
}
