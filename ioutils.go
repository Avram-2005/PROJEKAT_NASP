package main

import (
	"encoding/binary"
	"time"
)

const (
	CRC_L          = 4
	TIMESTAMP_L    = 8
	TOMBSTONE_L    = 1
	KEY_SIZE_L     = 4
	VALUE_SIZE_L   = 4
	OFFSET_L       = 8
	DATA_HEADER_L  = CRC_L + TIMESTAMP_L + TOMBSTONE_L + KEY_SIZE_L + VALUE_SIZE_L
	INDEX_HEADER_L = KEY_SIZE_L + OFFSET_L
	FOOTER_L       = 3 * OFFSET_L
)

type BufferWriter struct {
	buf []byte
	pos int
}

type BufferReader struct {
	buf []byte
	pos int
}

func NewBufferWriter(size int) *BufferWriter {
	return &BufferWriter{
		buf: make([]byte, size),
		pos: 0,
	}
}

func NewBufferReader(buf []byte) *BufferReader {
	return &BufferReader{
		buf: buf,
		pos: 0,
	}
}

func (w *BufferWriter) Seek(offset int) {
	w.pos = offset
}

func (w *BufferWriter) WriteCRC(crc uint32) {
	binary.BigEndian.PutUint32(w.buf[w.pos:], crc)
	w.pos += CRC_L
}

func (w *BufferWriter) WriteTimestamp(t time.Time) {
	binary.BigEndian.PutUint64(w.buf[w.pos:], uint64(t.UnixNano()))
	w.pos += TIMESTAMP_L
}

func (w *BufferWriter) WriteTombstone(tombstone bool) {
	if tombstone {
		w.buf[w.pos] = 1
	} else {
		w.buf[w.pos] = 0
	}
	w.pos += TOMBSTONE_L
}

func (w *BufferWriter) WriteKeySize(size int) {
	binary.BigEndian.PutUint32(w.buf[w.pos:], uint32(size))
	w.pos += KEY_SIZE_L
}

func (w *BufferWriter) WriteValueSize(size int) {
	binary.BigEndian.PutUint32(w.buf[w.pos:], uint32(size))
	w.pos += VALUE_SIZE_L
}

func (w *BufferWriter) WriteOffset(offset uint64) {
	binary.BigEndian.PutUint64(w.buf[w.pos:], offset)
	w.pos += OFFSET_L
}

func (w *BufferWriter) WriteBytes(data []byte) {
	copy(w.buf[w.pos:], data)
	w.pos += len(data)
}

func (r *BufferReader) ReadCRC() uint32 {
	crc := binary.BigEndian.Uint32(r.buf[r.pos:])
	r.pos += CRC_L
	return crc
}

func (r *BufferReader) ReadTimestamp() time.Time {
	timestamp := binary.BigEndian.Uint64(r.buf[r.pos:])
	r.pos += TIMESTAMP_L
	return time.Unix(0, int64(timestamp))
}

func (r *BufferReader) ReadTombstone() bool {
	tombstone := r.buf[r.pos] == 1
	r.pos += TOMBSTONE_L
	return tombstone
}

func (r *BufferReader) ReadKeySize() int {
	size := binary.BigEndian.Uint32(r.buf[r.pos:])
	r.pos += KEY_SIZE_L
	return int(size)
}

func (r *BufferReader) ReadValueSize() int {
	size := binary.BigEndian.Uint32(r.buf[r.pos:])
	r.pos += VALUE_SIZE_L
	return int(size)
}

func (r *BufferReader) ReadOffset() uint64 {
	offset := binary.BigEndian.Uint64(r.buf[r.pos:])
	r.pos += OFFSET_L
	return offset
}

func (r *BufferReader) ReadBytes(size int) []byte {
	data := make([]byte, size)
	copy(data, r.buf[r.pos:r.pos+size])
	r.pos += size
	return data
}
