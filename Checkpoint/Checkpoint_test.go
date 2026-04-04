package checkpoint

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

// provera funkcionalnosti samih hard linkova-ocekivano ponasanje je da
// pri upisu u originalni fajl hard link bude izmenjen, i obratno
func TestHardLink(t *testing.T) {
	file, err := os.Create("sstables/test.bin")
	if err != nil {
		fmt.Print("error creating file", err)
		t.FailNow()
	}
	data1 := make([]byte, 100)
	bm, err := BlockManager.NewBlockManager(4, 4)
	if err != nil {
		fmt.Print("error creating bm ", err)
		t.FailNow()
	}
	bm.Put(file, 0, &data1)

	err = os.Mkdir("checkpoints/dir1", 0755)
	if err != nil {
		fmt.Print("error creating directory ", err)
		t.FailNow()
	}

	fileName := file.Name()
	err = CreateHardLink(fileName, "dir1")
	if err != nil {
		fmt.Print("error creating link ", fileName, err)
		t.FailNow()
	}

	link, err := os.OpenFile("checkpoints/dir1/test.bin", os.O_RDWR, 0644)
	if err != nil {
		fmt.Print("error opening link file", err)
		t.FailNow()
	}
	binary.BigEndian.PutUint16(data1, 45)
	bm.Put(file, 0, &data1)

	readData, err := bm.Get(link, 0)
	if err != nil {
		fmt.Print("error getting from link file ", err)
		t.FailNow()
	}

	if !reflect.DeepEqual(data1, (*readData)) {
		if err != nil {
			fmt.Print("error on first compare")
			fmt.Print(data1, (*readData))
			t.FailNow()
		}
	}

	binary.BigEndian.PutUint16(data1, 67)
	err = bm.Put(link, 0, &data1)
	if err != nil {
		fmt.Print("error putting to link file", err)
		t.FailNow()
	}

	readData, err = bm.Get(file, 0)
	if err != nil {
		fmt.Print("error getting from link file", err)
		t.FailNow()
	}

	if !reflect.DeepEqual(data1, (*readData)) {
		if err != nil {
			fmt.Print("error on second compare")
			fmt.Print(data1, (*readData))
			t.FailNow()
		}
	}

	err = file.Close()
	if err != nil {
		fmt.Print("error closing file ", err)
		t.FailNow()
	}
	err = link.Close()
	if err != nil {
		fmt.Print("error closing link ", err)
		t.FailNow()
	}
	err = os.Remove("sstables/test.bin")
	if err != nil {
		fmt.Print("error deleting file ", err)
		t.FailNow()
	}
	err = os.Remove("checkpoints/dir1/test.bin")
	if err != nil {
		fmt.Print("error deleting file ", err)
		t.FailNow()
	}
	err = os.Remove("checkpoints/dir1")
	if err != nil {
		fmt.Print("error deleting directory ", err)
		t.FailNow()
	}

}

// testiranje kreiranja checkpoint-a za odredjen direktorijum
// zelimo da za svaki fajl unutar odabranog direktorijuma nastane hard link unutar checkpoints-a
func TestCheckpoint(t *testing.T) {
	firstTable, err := os.Create("sstables/test1.bin")
	if err != nil {
		fmt.Print("error creating file", err)
		t.FailNow()
	}
	secondTable, err := os.Create("sstables/test2.bin")
	if err != nil {
		fmt.Print("error creating file", err)
		t.FailNow()
	}
	thirdTable, err := os.Create("sstables/test3.bin")
	if err != nil {
		fmt.Print("error creating file", err)
		t.FailNow()
	}

	bm, err := BlockManager.NewBlockManager(4, 4)
	if err != nil {
		fmt.Print("error creating bm ", err)
		t.FailNow()
	}
	data1 := make([]byte, 100)
	binary.BigEndian.PutUint16(data1, 45)
	bm.Put(firstTable, 0, &data1)
	data2 := make([]byte, 100)
	binary.BigEndian.PutUint16(data2, 67)
	bm.Put(secondTable, 0, &data2)
	data3 := make([]byte, 100)
	binary.BigEndian.PutUint16(data3, 89)
	bm.Put(thirdTable, 0, &data3)

	err = CreateCheckpoint("sstables")
	if err != nil {
		fmt.Print("error creating link ", err)
		t.FailNow()
	}

	firstLink, err := os.OpenFile("checkpoints/sstables/test1.bin", os.O_RDWR, 0644)
	if err != nil {
		fmt.Print("error opening link file", err)
		t.FailNow()
	}
	secondLink, err := os.OpenFile("checkpoints/sstables/test1.bin", os.O_RDWR, 0644)
	if err != nil {
		fmt.Print("error opening link file", err)
		t.FailNow()
	}
	thirdLink, err := os.OpenFile("checkpoints/sstables/test1.bin", os.O_RDWR, 0644)
	if err != nil {
		fmt.Print("error opening link file", err)
		t.FailNow()
	}

	readData, err := bm.Get(firstLink, 0)
	if err != nil {
		fmt.Print("error getting from link file ", err)
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, (*readData)) {
		if err != nil {
			fmt.Print("error on first compare")
			fmt.Print(data1, (*readData))
			t.FailNow()
		}
	}
	readData, err = bm.Get(secondLink, 0)
	if err != nil {
		fmt.Print("error getting from link file ", err)
		t.FailNow()
	}
	if !reflect.DeepEqual(data2, (*readData)) {
		if err != nil {
			fmt.Print("error on second compare")
			fmt.Print(data2, (*readData))
			t.FailNow()
		}
	}
	readData, err = bm.Get(thirdLink, 0)
	if err != nil {
		fmt.Print("error getting from link file ", err)
		t.FailNow()
	}
	if !reflect.DeepEqual(data3, (*readData)) {
		if err != nil {
			fmt.Print("error on third compare")
			fmt.Print(data3, (*readData))
			t.FailNow()
		}
	}

	err = firstTable.Close()
	if err != nil {
		fmt.Print("error closing file ", err)
		t.FailNow()
	}
	err = secondTable.Close()
	if err != nil {
		fmt.Print("error closing file ", err)
		t.FailNow()
	}
	err = thirdTable.Close()
	if err != nil {
		fmt.Print("error closing file ", err)
		t.FailNow()
	}
	err = firstLink.Close()
	if err != nil {
		fmt.Print("error closing link ", err)
		t.FailNow()
	}
	err = secondLink.Close()
	if err != nil {
		fmt.Print("error closing link ", err)
		t.FailNow()
	}
	err = thirdLink.Close()
	if err != nil {
		fmt.Print("error closing link ", err)
		t.FailNow()
	}
	err = os.Remove("sstables/test1.bin")
	if err != nil {
		fmt.Print("error deleting file ", err)
		t.FailNow()
	}
	err = os.Remove("sstables/test2.bin")
	if err != nil {
		fmt.Print("error deleting file ", err)
		t.FailNow()
	}
	err = os.Remove("sstables/test3.bin")
	if err != nil {
		fmt.Print("error deleting file ", err)
		t.FailNow()
	}
	err = os.Remove("checkpoints/sstables/test1.bin")
	if err != nil {
		fmt.Print("error deleting file ", err)
		t.FailNow()
	}
	err = os.Remove("checkpoints/sstables/test2.bin")
	if err != nil {
		fmt.Print("error deleting file ", err)
		t.FailNow()
	}
	err = os.Remove("checkpoints/sstables/test3.bin")
	if err != nil {
		fmt.Print("error deleting file ", err)
		t.FailNow()
	}
	err = os.Remove("checkpoints/sstables")
	if err != nil {
		fmt.Print("error deleting directory ", err)
		t.FailNow()
	}

}
