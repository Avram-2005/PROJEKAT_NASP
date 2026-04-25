package sstable

import (
	"io"
	"os"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

type blockWriter struct {
	block        []byte
	currBlockNum int
	currByte     int
	file         *os.File
	bm           *BlockManager.BlockManager
}

func newBlockWriter(file *os.File, bm *BlockManager.BlockManager) *blockWriter {
	return &blockWriter{
		block:        make([]byte, bm.GetBlockSize()),
		currBlockNum: 0,
		currByte:     0,
		file:         file,
		bm:           bm,
	}
}

func (bw *blockWriter) CurrOffset() uint64 {
	return uint64(bw.currBlockNum*cap(bw.block) + bw.currByte)
}

func (bw *blockWriter) flush() {
	bw.bm.Put(bw.file, bw.currBlockNum, &bw.block)
	bw.currBlockNum += 1
	bw.currByte = 0
	bw.block = make([]byte, cap(bw.block))
}

func (bw *blockWriter) copyToBlock(data []byte, offset int) int {
	availableSpace := cap(bw.block) - bw.currByte
	if availableSpace == 0 {
		bw.flush()
		availableSpace = cap(bw.block)
	}
	n := copy(bw.block[bw.currByte:], data[offset:])
	bw.currByte += n
	return n
}

// TODO: Consider adding padding
func (bw *blockWriter) Write(data []byte) int {
	totalWritten := 0
	toWrite := len(data)

	for toWrite > 0 {
		n := bw.copyToBlock(data, totalWritten)
		totalWritten += n
		toWrite -= n
	}

	return totalWritten
}

func (bw *blockWriter) Finalize() {
	if bw.currByte > 0 {
		block := bw.block[:bw.currByte]
		bw.bm.Put(bw.file, bw.currBlockNum, &block)
	}
}

func (bw *blockWriter) Seek(offset int) {
	bw.currBlockNum = offset / cap(bw.block)
	bw.currByte = offset % cap(bw.block)
}

type blockReader struct {
	block        []byte
	currBlockNum int
	currByte     int
	file         *os.File
	bm           *BlockManager.BlockManager
}

func newBlockReader(file *os.File, bm *BlockManager.BlockManager, offset uint64) *blockReader {
	br := blockReader{
		block:        make([]byte, bm.GetBlockSize()),
		currBlockNum: int(offset / uint64(bm.GetBlockSize())),
		currByte:     int(offset % uint64(bm.GetBlockSize())),
		file:         file,
		bm:           bm,
	}
	br.readBlock()
	return &br
}

func (br *blockReader) CurrOffset() uint64 {
	return uint64(br.currBlockNum*cap(br.block) + br.currByte)
}

func (br *blockReader) readBlock() error {
	block, err := br.bm.Get(br.file, br.currBlockNum)
	if err != nil {
		return err
	}
	br.block = *block
	return nil
}

// TODO: Consider adding padding
func (br *blockReader) Read(dest []byte) (int, error) {
	totalRead := 0
	toRead := len(dest)

	for toRead > 0 {
		if br.currByte >= len(br.block) {
			br.currBlockNum += 1
			err := br.readBlock()
			if err == io.EOF {
				break
			} else if err != nil {
				return totalRead, err
			}
			br.currByte = 0
		}

		n := copy(dest[totalRead:], br.block[br.currByte:])
		br.currByte += n
		totalRead += n
		toRead -= n
	}

	if toRead > 0 {
		if totalRead == 0 {
			return 0, io.EOF
		}
		return totalRead, io.ErrUnexpectedEOF
	}

	return totalRead, nil
}

func (br *blockReader) Skip(offset int) {
	br.currByte += offset
}
