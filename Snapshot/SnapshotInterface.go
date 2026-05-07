package snapshot

import (
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

type SnapshotInterface interface {
	GetTimestamp() time.Time
	GetValue(bm *BlockManager.BlockManager) (*[]byte, error)
	GetType() string
}
