package snapshot

import (
	"time"

	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
)

type SnapshotMemtable struct {
	key              string
	memtableInstance *memtable.MemtableAdapter
	timestamp        time.Time
}

func NewSnapshotMemtable(key string, memtableInstance *memtable.MemtableAdapter) (*SnapshotMemtable, error) {
	rec, _, err := memtableInstance.GetRecord(key)
	if err != nil {
		return nil, err
	}

	return &SnapshotMemtable{
		key:              key,
		memtableInstance: memtableInstance,
		timestamp:        rec.Timestamp,
	}, nil
}

func (sp *SnapshotMemtable) GetTimestamp() time.Time {
	return sp.timestamp
}

func (sp *SnapshotMemtable) GetValue() (*[]byte, error) {
	value, ok, err := sp.memtableInstance.Get(sp.key)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &value, nil
}

func (sp *SnapshotMemtable) GetType() string {
	return "memtable"
}
