package SnapshotManager

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	engine "github.com/Avram-2005/PROJEKAT_NASP/Engine"
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
		MaxSizeBytes:      200,
		MaxSizeEntries:    10,
		SkipListMaxHeight: 8,
		BPlusTreeDegree:   2,
	}
	memtableManager, err := memtable.NewMemtableManager(5, config, func(entries []*record.Record) error {
		return nil
	})
	if err != nil {
		fmt.Print("error initializing memtable")
		t.FailNow()
	}

	for i := 0; i < 3; i++ {
		temp := make([]byte, 10)
		random := uint32(rand.Intn(100))
		binary.BigEndian.PutUint32(temp, random)
		key := "key" + strconv.Itoa(i)
		memtableManager.Put(key, temp)
	}
	for i := 0; i < 3; i++ {
		temp := make([]byte, 10)
		random := uint32(rand.Intn(100))
		binary.BigEndian.PutUint32(temp, random)
		key := "key" + strconv.Itoa(i)
		memtableManager.Put(key, temp)
	}
	memtables := memtableManager.GetMemtables()

	sstables := make([]*sstable.SSTable, 3)
	sstableManager := &sstable.SSTableManager{}
	sp.AddMany("key0", &memtables, &sstables, sstableManager)
	sp.AddMany("key1", &memtables, &sstables, sstableManager)
	sp.AddMany("key2", &memtables, &sstables, sstableManager)
	versions, err := sp.GetVersionCount("key0")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	fmt.Print(versions)
	versions, err = sp.GetVersionCount("key1")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	fmt.Print(versions)
	versions, err = sp.GetVersionCount("key2")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	fmt.Print(versions)
}

func TestSnapshotWithEngine(t *testing.T) {
	sp, err := NewSnapshotManager()
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom incijalizacije SnapshotManager-a")
		t.FailNow()
	}
	eng, err := engine.NewEngine("snapshot.yaml", "TestDataBase/walDATA", "TestDataBase/sstable")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	for i := 0; i < 3; i++ {
		temp := make([]byte, 10)
		random := uint32(rand.Intn(100))
		binary.BigEndian.PutUint32(temp, random)
		key := "key" + strconv.Itoa(i)
		eng.Put(key, temp)
	}
	for i := 0; i < 3; i++ {
		temp := make([]byte, 10)
		random := uint32(rand.Intn(100))
		binary.BigEndian.PutUint32(temp, random)
		key := "key" + strconv.Itoa(i)
		eng.Put(key, temp)
	}
	for i := 0; i < 3; i++ {
		temp := make([]byte, 10)
		random := uint32(rand.Intn(100))
		binary.BigEndian.PutUint32(temp, random)
		key := "key" + strconv.Itoa(i)
		eng.Put(key, temp)
	}
	for i := 0; i < 3; i++ {
		temp := make([]byte, 10)
		random := uint32(rand.Intn(100))
		binary.BigEndian.PutUint32(temp, random)
		key := "key" + strconv.Itoa(i)
		eng.Put(key, temp)
	}
	memtables := eng.GetAllMemtables()
	sstables, err := eng.GetAllSSTablesForSnapshot()
	if err != nil {
		fmt.Print("error getting sstables")
		fmt.Print(err)
		t.FailNow()
	}
	sstableManager := eng.GetSSTableManager()
	sp.AddMany("key0", &memtables, &sstables, &sstableManager)
	sp.AddMany("key1", &memtables, &sstables, &sstableManager)
	sp.AddMany("key2", &memtables, &sstables, &sstableManager)
	versions, err := sp.GetVersionCount("key0")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	if versions != 4 {
		fmt.Println("wrong version amount for key0")
		fmt.Println(versions)
		t.FailNow()
	}
	fmt.Println(versions)
	versions, err = sp.GetVersionCount("key1")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	if versions != 4 {
		fmt.Println("wrong version amount for key1")
		fmt.Println(versions)
		t.FailNow()
	}
	fmt.Println(versions)
	versions, err = sp.GetVersionCount("key2")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	if versions != 4 {
		fmt.Println("wrong version amount for key2")
		fmt.Println(versions)
		t.FailNow()
	}
	fmt.Println(versions)

	eng.ShutDown()
	err = os.RemoveAll("TestDataBase")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
}
