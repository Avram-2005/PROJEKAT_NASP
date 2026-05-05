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

	var currentRecords []Record
	iter, err := sstm.NewSSTableIterator(&SSTable{
		path:        filename,
		isMultFiles: false,
		footer:      footer,
	}, "", false)
	if err != nil {
		return false, nil, fmt.Errorf("failed to create iterator: %v", err)
	}
	defer iter.Close()
	for iter.Rec != nil {
		currentRecords = append(currentRecords, *iter.Rec)
		_, err := iter.Next()
		if err != nil {
			break
		}
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

	var currentRecords []Record
	iter, err := sstm.NewSSTableIterator(&SSTable{
		path:        sstablePath,
		isMultFiles: true,
	}, "", false)
	if err != nil {
		return false, nil, fmt.Errorf("failed to create iterator: %v", err)
	}
	defer iter.Close()
	for iter.Rec != nil {
		currentRecords = append(currentRecords, *iter.Rec)
		_, err := iter.Next()
		if err != nil {
			break
		}
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
