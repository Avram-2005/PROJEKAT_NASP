package sstable

import (
	"encoding/binary"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

type dataBlockWriter struct {
	block        []byte
	currBlockNum int
	currByte     int
	filename     string
}

func newDataBlockWriter(filename string, blockSize int) *dataBlockWriter {
	return &dataBlockWriter{
		// TODO: Do with make and append
		block:        make([]byte, blockSize),
		currBlockNum: 1,
		currByte:     0,
		filename:     filename,
	}
}

const (
	CRC_L        = 4
	TIMESTAMP_L  = 8
	TOMBSTONE_L  = 1
	KEY_SIZE_L   = 4
	VALUE_SIZE_L = 4
	HEADER_L     = CRC_L + TIMESTAMP_L + TOMBSTONE_L + KEY_SIZE_L + VALUE_SIZE_L
)

func (dbw *dataBlockWriter) reset() {
	dbw.currBlockNum += 1
	dbw.currByte = 0
}

func (dbw *dataBlockWriter) Write(entry KeyValue, bm *BlockManager.BlockManager) {
	if dbw.currByte+HEADER_L > cap(dbw.block) {
		bm.Put(dbw.filename, dbw.currBlockNum, &dbw.block)
		dbw.reset()
	}

	crc := uint32(0) // TODO: Calculate CRC
	binary.LittleEndian.PutUint32(dbw.block[dbw.currByte:], crc)
	dbw.currByte += CRC_L

	timestamp := time.Now().UnixNano()
	binary.LittleEndian.PutUint64(dbw.block[dbw.currByte:], uint64(timestamp))
	dbw.currByte += TIMESTAMP_L

	dbw.block[dbw.currByte] = 0
	dbw.currByte += TOMBSTONE_L

	binary.LittleEndian.PutUint32(dbw.block[dbw.currByte:], uint32(len(entry.Key)))
	dbw.currByte += KEY_SIZE_L

	binary.LittleEndian.PutUint32(dbw.block[dbw.currByte:], uint32(len(entry.Value)))
	dbw.currByte += VALUE_SIZE_L

	toWrite := len(entry.Key) + len(entry.Value)
	for toWrite != 0 {
		if len(entry.Key) > 0 {
			n := copy(dbw.block[dbw.currByte:], entry.Key)
			dbw.currByte += n
			entry.Key = entry.Key[n:]
			toWrite -= n
		} else if len(entry.Value) > 0 {
			n := copy(dbw.block[dbw.currByte:], entry.Value)
			dbw.currByte += n
			entry.Value = entry.Value[n:]
			toWrite -= n
		}

		if dbw.currByte == cap(dbw.block) && toWrite > 0 {
			bm.Put(dbw.filename, dbw.currBlockNum, &dbw.block)
			dbw.reset()
		}
	}

	if len(entry.Key) > 0 {
		n := copy(dbw.block[dbw.currByte:], entry.Key)
		dbw.currByte += n
	}
}

// TODO: Perhaps cut off dbw.block to currByte size?
func (dbw *dataBlockWriter) Finalize(bm *BlockManager.BlockManager) {
	if dbw.currByte > 0 {
		bm.Put(dbw.filename, dbw.currBlockNum, &dbw.block)
	}
}
