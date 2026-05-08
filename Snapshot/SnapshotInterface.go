package snapshot

import "time"

type SnapshotInterface interface {
	GetTimestamp() time.Time
	GetValue() (*[]byte, error)
	GetType() string
}
