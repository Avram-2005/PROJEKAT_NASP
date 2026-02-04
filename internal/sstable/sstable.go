package sstable

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

// FIXME: DELETE AFTER Memtable MERGE /
// ////////////////////////////////////

type KeyValue struct {
	Key       string
	Value     []byte
	Tombstone bool //za brisanje, true ako je obrisan
}

type Memtable interface {
	GetSortedEntries() []KeyValue //povratna vred/ parovi kljuc-vred neophodni za sstable
}

//////////////////////////////////////

// TODO: Compression (1.3[DZ3])

// TODO: Save to multiple files (Cassandra) or in one file (LevelDB) (1.3[DZ2])
type SSTable struct {
	index   index
	summary summary
}

func (sst *SSTable) Get(key string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// FIXME: Delete this after DB structure is done
var tablesRoot string

func SetupDirectory(root string) error {
	tablesRoot = filepath.Join(root, "tables")
	return os.MkdirAll(tablesRoot, os.ModePerm)
}

func Flush(mem Memtable, tableNum int, bm *BlockManager.BlockManager) error {
	filename := filepath.Join(tablesRoot, fmt.Sprintf("usertable-%d-Data.txt", tableNum))
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("file %s already exists", filename)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	dw := newDataBlockWriter(filename, bm.GetBlockSize())
	for _, entry := range mem.GetSortedEntries() {
		dw.Write(entry, bm)
	}
	dw.Finalize(bm)
	return nil
}
