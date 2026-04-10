package record

import (
	"fmt"
	. "github.com/Avram-2005/PROJEKAT_NASP/utils"
	"hash/crc32"
	"time"
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
		return nil, fmt.Errorf("Ne moze tombstone da bude false i da value ima vrednost!")
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

type RecordHeader struct {
	CRC       uint32
	Timestamp time.Time
	Tombstone bool
	KeySize   int
	ValueSize int
}

func DeserializeRecordHeader(data []byte) *RecordHeader {
	reader := NewBufferReaderReuse(data)

	crc := reader.ReadCRC()
	timestamp := reader.ReadTimestamp()
	tombstone := reader.ReadTombstone()
	keySize := reader.ReadKeySize()
	valueSize := reader.ReadValueSize()

	return &RecordHeader{
		CRC:       crc,
		Timestamp: timestamp,
		Tombstone: tombstone,
		KeySize:   keySize,
		ValueSize: valueSize,
	}
}

func DeserializeRecord(data []byte) (*Record, error) {
	header := DeserializeRecordHeader(data)

	realCrc := crc32.ChecksumIEEE(data[CRC_L:])
	if realCrc != header.CRC {
		return nil, fmt.Errorf("CRC mismatch: expected %d, got %d", header.CRC, realCrc)
	}

	reader := NewBufferReaderReuse(data[HEADER_SIZE:])

	key := string(reader.ReadBytes(header.KeySize))
	value := reader.ReadBytes(header.ValueSize)

	return &Record{
		Timestamp: header.Timestamp,
		Tombstone: header.Tombstone,
		Key:       key,
		Value:     value,
	}, nil
}

func (r *Record) Size() int {
	return HEADER_SIZE + len(r.Key) + len(r.Value)
}
