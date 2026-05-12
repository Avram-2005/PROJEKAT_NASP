package record

import (
	"fmt"
	"hash/crc32"
	"time"

	. "github.com/Avram-2005/PROJEKAT_NASP/utils"
)

type Record struct {
	Timestamp time.Time
	Tombstone bool
	Key       string
	Value     []byte
}

const HEADER_SIZE = CRC_L + TIMESTAMP_L + TOMBSTONE_L + KEY_SIZE_L + VALUE_SIZE_L

func NewRecord(key string, value []byte, tombstone bool, timestamp time.Time) (*Record, error) {
	if tombstone && len(value) != 0 {
		return nil, fmt.Errorf("tombstone cannot be true when value is not empty")
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	return &Record{
		Timestamp: timestamp,
		Tombstone: tombstone,
		Key:       key,
		Value:     value,
	}, nil
}

func (r *Record) Serialize() []byte {
	keySize := len(r.Key)
	valueSize := len(r.Value)
	totalSize := HEADER_SIZE + keySize + valueSize

	writer := NewBufferWriter(totalSize)

	writer.Seek(CRC_L)
	writer.WriteTimestamp(r.Timestamp)
	writer.WriteTombstone(r.Tombstone)
	writer.WriteKeySize(keySize)
	writer.WriteValueSize(valueSize)
	writer.WriteBytes([]byte(r.Key))
	writer.WriteBytes(r.Value)

	CRC := crc32.ChecksumIEEE(writer.Buf[CRC_L:])
	writer.Seek(0)
	writer.WriteCRC(CRC)

	return writer.Buf
}

func DeserializeRecord(data []byte) (*Record, int, error) {
	if len(data) < HEADER_SIZE {
		return nil, 0, fmt.Errorf("insufficient data for header")
	}

	reader := NewBufferReaderReuse(data)

	crc := reader.ReadCRC()
	ts := reader.ReadTimestamp()
	tomb := reader.ReadTombstone()
	kSize := int(reader.ReadKeySize())
	vSize := int(reader.ReadValueSize())

	totalSize := HEADER_SIZE + kSize + vSize

	if len(data) < totalSize {
		return nil, 0, fmt.Errorf("not enough data: expected %d, got %d", totalSize, len(data))
	}

	payloadForChecksum := data[4:totalSize]
	actualCRC := crc32.ChecksumIEEE(payloadForChecksum)

	if crc != actualCRC {
		return nil, 0, fmt.Errorf("CRC invalid: expected %08x, got %08x", crc, actualCRC)
	}

	key := string(reader.ReadBytes(kSize))
	value := reader.ReadBytes(vSize)

	rec := &Record{Timestamp: ts, Tombstone: tomb, Key: key, Value: value}

	return rec, totalSize, nil
}

func (r *Record) Size() int {
	return HEADER_SIZE + len(r.Key) + len(r.Value)
}
