package sstable

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	merkleTree "github.com/Avram-2005/PROJEKAT_NASP/MerkleTree"
	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func (sstm *SSTableManager) validateOneFile(filename string) (bool, []Record, error) {
	f, err := os.Open(filename)
	if err != nil {
		return false, nil, fmt.Errorf("failed to open SSTable file: %v", err)
	}
	defer f.Close()

	footer, err := sstm.loadOneFileFooter(f)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read SSTable footer: %v", err)
	}

	metadataReader := newBlockReader(f, sstm.bm, footer.MetadataStart)

	stat, _ := f.Stat()
	footerStart := uint64(stat.Size()) - FOOTER_L
	metadataSize := footerStart - footer.MetadataStart

	metadataData := make([]byte, metadataSize)
	_, err = metadataReader.Read(metadataData)
	if err != nil && err != io.EOF {
		return false, nil, err
	}

	originalTree := merkleTree.Deserialize(metadataData)
	if originalTree == nil {
		return false, nil, fmt.Errorf("failed to deserialize merkle tree")
	}

	dataReader := newBlockReader(f, sstm.bm, 0)

	var currentRecords []Record
	for {
		currentOffset := dataReader.CurrOffset()
		if currentOffset >= footer.IndexStart {
			break
		}

		rec, err := sstm.parseData(dataReader, false)
		if err != nil {
			break
		}
		currentRecords = append(currentRecords, *rec)
	}

	currentTree, err := merkleTree.NewMerkleTree(currentRecords)
	if err != nil {
		return false, nil, err
	}

	if originalTree.Verify(currentTree.RootHash()) {
		return true, nil, nil
	}

	diffs := merkleTree.FindDifference(originalTree.Root(), currentTree.Root())

	return false, diffs, nil
}

func (sstm *SSTableManager) validateMultipleFiles(sstablePath string) (bool, []Record, error) {
	metadataFilename := sstableFilenameMultFile(sstablePath, "Metadata")
	metadataFile, err := os.Open(metadataFilename)
	if err != nil {
		return false, nil, fmt.Errorf("failed to open metadata file: %v", err)
	}
	defer metadataFile.Close()

	metadataReader := newBlockReader(metadataFile, sstm.bm, 0)

	sizeHeader := make([]byte, 4)
	_, err = metadataReader.Read(sizeHeader)
	if err != nil {
		return false, nil, fmt.Errorf("failed to read size header: %v", err)
	}
	treeSize := binary.BigEndian.Uint32(sizeHeader)

	if treeSize == 0 {
		return false, nil, fmt.Errorf("invalid tree size: %d", treeSize)
	}

	metadataData := make([]byte, treeSize)
	_, err = metadataReader.Read(metadataData)
	if err != nil {
		return false, nil, err
	}

	originalTree := merkleTree.Deserialize(metadataData)
	if originalTree == nil {
		return false, nil, fmt.Errorf("failed to deserialize merkle tree")
	}

	dataFilename := sstableFilenameMultFile(sstablePath, "Data")
	dataFile, err := os.Open(dataFilename)
	if err != nil {
		return false, nil, fmt.Errorf("failed to open data file: %v", err)
	}
	defer dataFile.Close()

	stat, err := dataFile.Stat()
	if err != nil {
		return false, nil, err
	}
	fileSize := stat.Size()

	dataReader := newBlockReader(dataFile, sstm.bm, 0)

	var currentRecords []Record
	for {
		if int64(dataReader.CurrOffset()) >= fileSize {
			break
		}

		rec, err := sstm.parseData(dataReader, false)
		if err != nil {
			break
		}
		currentRecords = append(currentRecords, *rec)
	}

	currentTree, err := merkleTree.NewMerkleTree(currentRecords)
	if err != nil {
		return false, nil, err
	}

	if originalTree.Verify(currentTree.RootHash()) {
		return true, nil, nil
	}

	diffs := merkleTree.FindDifference(originalTree.Root(), currentTree.Root())

	return false, diffs, nil
}
