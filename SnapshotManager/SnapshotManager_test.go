package SnapshotManager

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	sstable "github.com/Avram-2005/PROJEKAT_NASP/LSM"
	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func TestSnapshotManager(t *testing.T) {
	sp, err := NewSnapshotManager()
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom incijalizacije SnapshotManager-a")
		t.FailNow()
	}
	config := memtable.MemtableConfig{
		Type:              "hashmap",
		MaxSizeEntries:    10,
		SkipListMaxHeight: 8,
		BPlusTreeDegree:   2,
	}
	memtableManager, err := memtable.NewMemtableManager(3, config, func(entries []*record.Record) error {
		return nil
	})
	if err != nil {
		fmt.Print("error initializing memtable")
		t.FailNow()
	}
	memtables := memtableManager.GetMemtables()
	for i := 0; i < 3; i++ {
		temp := make([]byte, 10)
		random := uint32(rand.Intn(100))
		binary.BigEndian.PutUint32(temp, random)
		key := "key" + strconv.Itoa(i)
		memtables[0].Put(key, temp)
	}
	for i := 0; i < 2; i++ {
		temp := make([]byte, 100)
		random := uint32(rand.Intn(100))
		binary.BigEndian.PutUint32(temp, random)
		key := "key" + strconv.Itoa(i)
		memtables[1].Put(key, temp)
	}
	for i := 0; i < 3; i++ {
		temp := make([]byte, 100)
		random := uint32(rand.Intn(100))
		binary.BigEndian.PutUint32(temp, random)
		key := "key" + strconv.Itoa(i)
		memtables[2].Put(key, temp)
	}
	sstables := make([]sstable.SSTable, 3)
	sstableManager := &sstable.SSTableManager{}
	sp.AddMany("key0", &memtables, &sstables, sstableManager)
	sp.AddMany("key1", &memtables, &sstables, sstableManager)
	sp.AddMany("key2", &memtables, &sstables, sstableManager)

}
