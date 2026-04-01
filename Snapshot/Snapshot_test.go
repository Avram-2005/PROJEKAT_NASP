package snapshot

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

func TestSnapshot(t *testing.T) {
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
	snapshot1, err := NewSnapshot(filepath, 0, 0, 100, time.Now(), bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom incijalizacije snapshota broj 1")
		t.FailNow()
	}
	snapshot2, err := NewSnapshot(filepath, 0, 100, 100, time.Now(), bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom incijalizacije snapshota broj 2")
		t.FailNow()
	}
	snapshot3, err := NewSnapshot(filepath, 0, 200, 100, time.Now(), bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom incijalizacije snapshota broj 3")
		t.FailNow()
	}
	readData1, err := snapshot1.GetValue(bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom dobavljanja vrednosti")
		t.FailNow()
	}
	if reflect.DeepEqual(data1, (*readData1)) != true {
		fmt.Print(data1)
		fmt.Print(*readData1)
		t.Errorf("neocekivana vrednost podataka prvog snapshot-a")
		t.FailNow()
	}
	readData2, err := snapshot2.GetValue(bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom dobavljanja vrednosti")
		t.FailNow()
	}
	if reflect.DeepEqual(data2, (*readData2)) != true {
		t.Errorf("neocekivana vrednost podataka drugog snapshot-a")
		t.FailNow()
	}
	readData3, err := snapshot3.GetValue(bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom dobavljanja vrednosti")
		t.FailNow()
	}
	if reflect.DeepEqual(data3, (*readData3)) != true {
		t.Errorf("neocekivana vrednost podataka treceg snapshot-a")
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
