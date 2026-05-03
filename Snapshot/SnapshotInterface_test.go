package Snapshot

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

func TestSnapshotSSTable(t *testing.T) {
	filepath := "test.bin"
	file, err := os.Create(filepath)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom stvaranja fajla")
		t.FailNow()
	}
	bm, err := BlockManager.NewBlockManager(4, 4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("treba da se prijavi greska, ali nije prijavljena")
		t.FailNow()
	}
	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 100)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 100)
	binary.BigEndian.PutUint16(data3, 67)
	bm.PutSpecific(file, 0, 0, 100, &data1)
	bm.PutSpecific(file, 0, 100, 100, &data2)
	bm.PutSpecific(file, 0, 200, 100, &data3)
	SnapshotSSTable1, err := NewSnapshotSSTable(filepath, 0, 0, 100, time.Now(), bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom incijalizacije SnapshotSSTablea broj 1")
		t.FailNow()
	}
	SnapshotSSTable2, err := NewSnapshotSSTable(filepath, 0, 100, 100, time.Now(), bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom incijalizacije SnapshotSSTablea broj 2")
		t.FailNow()
	}
	SnapshotSSTable3, err := NewSnapshotSSTable(filepath, 0, 200, 100, time.Now(), bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom incijalizacije SnapshotSSTablea broj 3")
		t.FailNow()
	}
	readData1, err := SnapshotSSTable1.GetValue(bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom dobavljanja vrednosti")
		t.FailNow()
	}
	if reflect.DeepEqual(data1, (*readData1)) != true {
		fmt.Print(data1)
		fmt.Print(*readData1)
		t.Errorf("neocekivana vrednost podataka prvog SnapshotSSTable-a")
		t.FailNow()
	}
	readData2, err := SnapshotSSTable2.GetValue(bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom dobavljanja vrednosti")
		t.FailNow()
	}
	if reflect.DeepEqual(data2, (*readData2)) != true {
		t.Errorf("neocekivana vrednost podataka drugog SnapshotSSTable-a")
		t.FailNow()
	}
	readData3, err := SnapshotSSTable3.GetValue(bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom dobavljanja vrednosti")
		t.FailNow()
	}
	if reflect.DeepEqual(data3, (*readData3)) != true {
		t.Errorf("neocekivana vrednost podataka treceg SnapshotSSTable-a")
		t.FailNow()
	}
	file.Close()
	err = os.Remove(filepath)
	if err != nil {
		fmt.Print(err)
		t.Errorf("zatvaranje fajla onemoguceno")
		t.FailNow()
	}
}

func TestSnapshotInterface(t *testing.T) {
	bm, err := BlockManager.NewBlockManager(4, 4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("treba da se prijavi greska, ali nije prijavljena")
		t.FailNow()
	}
	snapshotList := list.New()
	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 100)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 100)
	binary.BigEndian.PutUint16(data3, 67)
	memSnapshot1 := NewSnapshotMemtable(&data1, time.Now())
	memSnapshot2 := NewSnapshotMemtable(&data2, time.Now())
	memSnapshot3 := NewSnapshotMemtable(&data3, time.Now())
	snapshotList.PushBack(memSnapshot1)
	snapshotList.PushBack(memSnapshot2)
	snapshotList.PushBack(memSnapshot3)
	compare1, err := snapshotList.Front().Value.(SnapshotInterface).GetValue(bm)
	if err != nil {
		t.FailNow()
		fmt.Print("something went wrong")
	}
	if !reflect.DeepEqual(data1, (*compare1)) {
		t.FailNow()
		fmt.Print(data1, (*compare1))
	}
}
