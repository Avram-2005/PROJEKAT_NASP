package record

import (
	"encoding/binary"
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

func (r *Record) SerializeVarInt() []byte {
	key := []byte(r.Key)
	value := r.Value

	payloadWriter := NewBufferWriter(3*binary.MaxVarintLen64 + TOMBSTONE_L + len(key) + len(value))
	payloadLen := 0
	payloadLen += payloadWriter.WriteTimestampVarint(r.Timestamp)
	payloadWriter.WriteTombstone(r.Tombstone)
	payloadLen += TOMBSTONE_L
	payloadLen += payloadWriter.WriteKeySizeVarint(len(key))
	payloadLen += payloadWriter.WriteValueSizeVarint(len(value))
	payloadWriter.WriteBytes(key)
	payloadLen += len(key)
	payloadWriter.WriteBytes(value)
	payloadLen += len(value)

	payload := payloadWriter.Buf[:payloadLen]
	crc := crc32.ChecksumIEEE(payload)

	writer := NewBufferWriter(binary.MaxVarintLen32 + payloadLen)
	totalLen := writer.WriteCRCVarint(crc)
	writer.WriteBytes(payload)
	totalLen += payloadLen

	return writer.Buf[:totalLen]
}

type RecordHeader struct {
	CRC       uint32
	Timestamp time.Time
	Tombstone bool
	KeySize   int
	ValueSize int
}

func DeserializeRecordHeaderVarInt(data []byte) (*RecordHeader, int, int, error) {
	reader := NewBufferReaderReuse(data)

	crc, err := reader.ReadCRCVarint()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read CRC: %v", err)
	}

	crcStart := reader.CurrOffset()

	timestamp, err := reader.ReadTimestampVarint()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read timestamp: %v", err)
	}
	tombstone := reader.ReadTombstone()
	keySize, err := reader.ReadKeySizeVarint()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read key size: %v", err)
	}
	valueSize, err := reader.ReadValueSizeVarint()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read value size: %v", err)
	}

	return &RecordHeader{
		CRC:       crc,
		Timestamp: timestamp,
		Tombstone: tombstone,
		KeySize:   keySize,
		ValueSize: valueSize,
	}, reader.CurrOffset(), crcStart, nil
}

func DeserializeRecord(data []byte) (*Record, error) {
	reader := NewBufferReaderReuse(data)

	crc := reader.ReadCRC()
	realCrc := crc32.ChecksumIEEE(data[CRC_L:])
	if crc != realCrc {
		return nil, fmt.Errorf("CRC mismatch: expected %d, got %d", crc, realCrc)
	}

	timestamp := reader.ReadTimestamp()
	tombstone := reader.ReadTombstone()
	keySize := reader.ReadKeySize()
	valueSize := reader.ReadValueSize()

	key := string(reader.ReadBytes(keySize))
	value := reader.ReadBytes(valueSize)

	return &Record{
		Timestamp: timestamp,
		Tombstone: tombstone,
		Key:       key,
		Value:     value,
	}, nil
}

func (r *Record) Size() int {
	return HEADER_SIZE + len(r.Key) + len(r.Value)
}
