package main

import "time"

type Record struct {
	CRC       uint32
	Timestamp time.Time
	Tombstone bool
	Key       string
	Value     []byte
}

const HEADER_SIZE = CRC_L + TIMESTAMP_L + TOMBSTONE_L + KEY_SIZE_L + VALUE_SIZE_L

func (r *Record) Serialize() []byte {
	keySize := len(r.Key)
	valueSize := len(r.Value)
	totalSize := HEADER_SIZE + keySize + valueSize

	writer := newBufferWriter(totalSize)

	writer.WriteCRC(r.CRC)
	writer.WriteTimestamp(r.Timestamp)
	writer.WriteTombstone(r.Tombstone)
	writer.WriteKeySize(keySize)
	writer.WriteValueSize(valueSize)
	writer.WriteBytes([]byte(r.Key))
	writer.WriteBytes(r.Value)

	return writer.buf
}

func DeserializeRecord(data []byte) (*Record, error) {
	reader := newBufferReader(data)

	crc := reader.ReadCRC()
	timestamp := reader.ReadTimestamp()
	tombstone := reader.ReadTombstone()
	keySize := reader.ReadKeySize()
	valueSize := reader.ReadValueSize()

	key := string(reader.ReadBytes(keySize))
	value := reader.ReadBytes(valueSize)

	return &Record{
		CRC:       crc,
		Timestamp: timestamp,
		Tombstone: tombstone,
		Key:       key,
		Value:     value,
	}, nil
}

func (r *Record) Size() int {
	return HEADER_SIZE + len(r.Key) + len(r.Value)
}
