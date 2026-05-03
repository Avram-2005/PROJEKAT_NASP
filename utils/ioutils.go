package utils

import (
	"encoding/binary"
	"fmt"
	"time"
)

const (
	CRC_L        = 4
	TIMESTAMP_L  = 8
	TOMBSTONE_L  = 1
	KEY_SIZE_L   = 4
	VALUE_SIZE_L = 4
	OFFSET_L     = 8
)

type BufferWriter struct {
	Buf []byte
	pos int
}

type BufferReader struct {
	Buf []byte
	pos int
}

func NewBufferWriter(size int) *BufferWriter {
	return &BufferWriter{
		Buf: make([]byte, size),
		pos: 0,
	}
}

func NewBufferReader(size int) *BufferReader {
	return &BufferReader{
		Buf: make([]byte, size),
		pos: 0,
	}
}

func NewBufferReaderReuse(buf []byte) *BufferReader {
	return &BufferReader{
		Buf: buf,
		pos: 0,
	}
}

func (w *BufferWriter) Seek(offset int) {
	w.pos = offset
}

func (w *BufferWriter) WriteCRC(crc uint32) {
	binary.BigEndian.PutUint32(w.Buf[w.pos:], crc)
	w.pos += CRC_L
}

func (w *BufferWriter) WriteTimestamp(t time.Time) {
	binary.BigEndian.PutUint64(w.Buf[w.pos:], uint64(t.UnixNano()))
	w.pos += TIMESTAMP_L
}

func (w *BufferWriter) WriteTombstone(tombstone bool) {
	if tombstone {
		w.Buf[w.pos] = 1
	} else {
		w.Buf[w.pos] = 0
	}
	w.pos += TOMBSTONE_L
}

func (w *BufferWriter) WriteKeySize(size int) {
	binary.BigEndian.PutUint32(w.Buf[w.pos:], uint32(size))
	w.pos += KEY_SIZE_L
}

func (w *BufferWriter) WriteValueSize(size int) {
	binary.BigEndian.PutUint32(w.Buf[w.pos:], uint32(size))
	w.pos += VALUE_SIZE_L
}

func (w *BufferWriter) WriteOffset(offset uint64) {
	binary.BigEndian.PutUint64(w.Buf[w.pos:], offset)
	w.pos += OFFSET_L
}

func (w *BufferWriter) WriteUVarint(value uint64) int {
	n := binary.PutUvarint(w.Buf[w.pos:], value)
	w.pos += n
	return n
}

func (w *BufferWriter) WriteVarint(value int64) int {
	n := binary.PutVarint(w.Buf[w.pos:], value)
	w.pos += n
	return n
}

func (w *BufferWriter) WriteCRCVarint(crc uint32) int {
	return w.WriteUVarint(uint64(crc))
}

func (w *BufferWriter) WriteTimestampVarint(t time.Time) int {
	return w.WriteVarint(t.UnixNano())
}

func (w *BufferWriter) WriteKeySizeVarint(size int) int {
	return w.WriteUVarint(uint64(size))
}

func (w *BufferWriter) WriteValueSizeVarint(size int) int {
	return w.WriteUVarint(uint64(size))
}

func (w *BufferWriter) WriteOffsetVarint(offset uint64) int {
	return w.WriteUVarint(offset)
}

func (w *BufferWriter) WriteBytes(data []byte) {
	copy(w.Buf[w.pos:], data)
	w.pos += len(data)
}

func (r *BufferReader) ReadCRC() uint32 {
	crc := binary.BigEndian.Uint32(r.Buf[r.pos:])
	r.pos += CRC_L
	return crc
}

func (r *BufferReader) ReadTimestamp() time.Time {
	timestamp := binary.BigEndian.Uint64(r.Buf[r.pos:])
	r.pos += TIMESTAMP_L
	return time.Unix(0, int64(timestamp))
}

func (r *BufferReader) ReadTombstone() bool {
	tombstone := r.Buf[r.pos] == 1
	r.pos += TOMBSTONE_L
	return tombstone
}

func (r *BufferReader) ReadKeySize() int {
	size := binary.BigEndian.Uint32(r.Buf[r.pos:])
	r.pos += KEY_SIZE_L
	return int(size)
}

func (r *BufferReader) ReadValueSize() int {
	size := binary.BigEndian.Uint32(r.Buf[r.pos:])
	r.pos += VALUE_SIZE_L
	return int(size)
}

func (r *BufferReader) ReadOffset() uint64 {
	offset := binary.BigEndian.Uint64(r.Buf[r.pos:])
	r.pos += OFFSET_L
	return offset
}

func (r *BufferReader) ReadUVarint() (uint64, error) {
	value, n := binary.Uvarint(r.Buf[r.pos:])
	if n == 0 {
		return 0, fmt.Errorf("failed to read uvarint: buffer too small")
	}
	if n < 0 {
		return 0, fmt.Errorf("failed to read uvarint: value overflow")
	}
	r.pos += n
	return value, nil
}

func (r *BufferReader) ReadVarint() (int64, error) {
	value, n := binary.Varint(r.Buf[r.pos:])
	if n == 0 {
		return 0, fmt.Errorf("failed to read varint: buffer too small")
	}
	if n < 0 {
		return 0, fmt.Errorf("failed to read varint: value overflow")
	}
	r.pos += n
	return value, nil
}

func (r *BufferReader) ReadCRCVarint() (uint32, error) {
	value, err := r.ReadUVarint()
	if err != nil {
		return 0, err
	}
	return uint32(value), nil
}

func (r *BufferReader) ReadTimestampVarint() (time.Time, error) {
	value, err := r.ReadVarint()
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, value), nil
}

func (r *BufferReader) ReadKeySizeVarint() (int, error) {
	value, err := r.ReadUVarint()
	if err != nil {
		return 0, err
	}
	return int(value), nil
}

func (r *BufferReader) ReadValueSizeVarint() (int, error) {
	value, err := r.ReadUVarint()
	if err != nil {
		return 0, err
	}
	return int(value), nil
}

func (r *BufferReader) ReadOffsetVarint() (uint64, error) {
	return r.ReadUVarint()
}

func (r *BufferReader) ReadBytes(size int) []byte {
	data := make([]byte, size)
	copy(data, r.Buf[r.pos:r.pos+size])
	r.pos += size
	return data
}
