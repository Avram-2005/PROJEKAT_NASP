package Snapshot

import (
	"os"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

type SnapshotSSTable struct {
	filepath    string
	blockNumber int
	offset      int
	size        int
	timestamp   time.Time
}

func NewSnapshotSSTable(filepath string, blockNumber int, offset int, size int, timestamp time.Time, bm *BlockManager.BlockManager) (*SnapshotSSTable, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	_, err = bm.GetSpecific(file, blockNumber, offset, size)
	file.Close()
	if err != nil {
		return nil, err
	}
	return &SnapshotSSTable{
		filepath:    filepath,
		blockNumber: blockNumber,
		offset:      offset,
		size:        size,
		timestamp:   timestamp,
	}, nil
}

func (sp *SnapshotSSTable) GetTimestamp() time.Time {
	return sp.timestamp
}

func (sp *SnapshotSSTable) GetValue(bm *BlockManager.BlockManager) (*[]byte, error) {
	file, err := os.Open(sp.filepath)
	if err != nil {

		return nil, err
	}
	returnValue, err := bm.GetSpecific(file, sp.blockNumber, sp.offset, sp.size)
	file.Close()
	if err != nil {
		return nil, err
	}
	return returnValue, nil
}
