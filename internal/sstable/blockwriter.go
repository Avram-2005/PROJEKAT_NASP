package sstable

import "github.com/Avram-2005/PROJEKAT_NASP/BlockManager"

type blockWriter struct {
	block        []byte
	currBlockNum int
	currByte     int
	filename     string
	bm           *BlockManager.BlockManager
}

func newBlockWriter(filename string, bm *BlockManager.BlockManager) *blockWriter {
	return &blockWriter{
		block:        make([]byte, bm.GetBlockSize()),
		currBlockNum: 1,
		currByte:     0,
		filename:     filename,
		bm:           bm,
	}
}

func (bw *blockWriter) flush() {
	bw.bm.Put(bw.filename, bw.currBlockNum, &bw.block)
	bw.currBlockNum += 1
	bw.currByte = 0
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
		bw.bm.Put(bw.filename, bw.currBlockNum, &block)
	}
}
