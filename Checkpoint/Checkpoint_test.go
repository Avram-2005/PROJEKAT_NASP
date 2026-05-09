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
	err := os.Mkdir("sstables", 0755)
	if err != nil {
		fmt.Print("error creating directory ", err)
		t.FailNow()
	}
	err = os.Mkdir("checkpoints", 0755)
	if err != nil {
		fmt.Print("error creating directory ", err)
		t.FailNow()
	}

	file1, err := os.Create("sstables/test.bin")
	if err != nil {
		fmt.Print("error creating file", err)
		t.FailNow()
	}

	err = os.Mkdir("sstables/sst1", 0755)
	if err != nil {
		fmt.Print("error creating directory ", err)
		t.FailNow()
	}

	file2, err := os.Create("sstables/sst1/test1.bin")
	if err != nil {
		fmt.Print("error creating file 2", err)
		t.FailNow()
	}

	bm, err := BlockManager.NewBlockManager(4, 4)
	if err != nil {
		fmt.Print("error creating bm ", err)
		t.FailNow()
	}

	data1 := make([]byte, 100)
	bm.Put(file1, 0, &data1)

	data2 := make([]byte, 100)
	binary.BigEndian.PutUint16(data2, 56)
	bm.Put(file2, 0, &data2)

	err = os.Mkdir("checkpoints/dir1", 0755)
	if err != nil {
		fmt.Print("error creating directory ", err)
		t.FailNow()
	}

	fileName := file1.Name()
	err = CreateHardLink(fileName, "dir1")
	if err != nil {
		fmt.Print("error creating link ", fileName, err)
		t.FailNow()
	}
	fileName = file2.Name()
	err = CreateHardLink(fileName, "dir1")
	if err != nil {
		fmt.Print("error creating link ", fileName, err)
		t.FailNow()
	}

	link1, err := os.OpenFile("checkpoints/dir1/test.bin", os.O_RDWR, 0644)
	if err != nil {
		fmt.Print("error opening link file", err)
		t.FailNow()
	}
	binary.BigEndian.PutUint16(data1, 45)
	bm.Put(file1, 0, &data1)

	readData, err := bm.Get(link1, 0)
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

	link2, err := os.OpenFile("checkpoints/dir1/sst1/test1.bin", os.O_RDWR, 0644)
	if err != nil {
		fmt.Print("error opening link file", err)
		t.FailNow()
	}

	readData, err = bm.Get(link2, 0)
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

	binary.BigEndian.PutUint16(data1, 67)
	err = bm.Put(link1, 0, &data1)
	if err != nil {
		fmt.Print("error putting to link file", err)
		t.FailNow()
	}

	readData, err = bm.Get(file1, 0)
	if err != nil {
		fmt.Print("error getting from link file", err)
		t.FailNow()
	}

	if !reflect.DeepEqual(data1, (*readData)) {
		if err != nil {
			fmt.Print("error on third compare")
			fmt.Print(data1, (*readData))
			t.FailNow()
		}
	}

	err = file1.Close()
	if err != nil {
		fmt.Print("error closing file ", err)
		t.FailNow()
	}
	err = file2.Close()
	if err != nil {
		fmt.Print("error closing file ", err)
		t.FailNow()
	}
	err = link1.Close()
	if err != nil {
		fmt.Print("error closing link ", err)
		t.FailNow()
	}
	err = link2.Close()
	if err != nil {
		fmt.Print("error closing link ", err)
		t.FailNow()
	}
	err = os.Remove("sstables/test.bin")
	if err != nil {
		fmt.Print("error deleting file ", err)
		t.FailNow()
	}
	err = os.RemoveAll("sstables/sst1")
	if err != nil {
		fmt.Print("error deleting directory ", err)
		t.FailNow()
	}
	err = DeleteCheckpointDirectory("dir1")
	if err != nil {
		fmt.Print("error deleting checkpoint", err)
		t.FailNow()
	}
	err = os.Remove("checkpoints")
	if err != nil {
		fmt.Print("error deleting checkpoint", err)
		t.FailNow()
	}
	err = os.Remove("sstables")
	if err != nil {
		fmt.Print("error deleting checkpoint", err)
		t.FailNow()
	}
}

// testiranje kreiranja checkpoint-a za odredjen direktorijum
// zelimo da za svaki fajl unutar odabranog direktorijuma nastane hard link unutar checkpoints-a
func TestCheckpoint(t *testing.T) {
	err := os.Mkdir("sstables", 0755)
	if err != nil {
		fmt.Print("error creating directory ", err)
		t.FailNow()
	}
	err = os.Mkdir("checkpoints", 0755)
	if err != nil {
		fmt.Print("error creating directory ", err)
		t.FailNow()
	}

	firstTable, err := os.Create("sstables/test1.bin")
	if err != nil {
		fmt.Print("error creating file 1", err)
		t.FailNow()
	}
	err = os.Mkdir("sstables/sst1", 0755)
	if err != nil {
		fmt.Print("error creating directory 1", err)
		t.FailNow()
	}
	secondTable, err := os.Create("sstables/sst1/test2.bin")
	if err != nil {
		fmt.Print("error creating file 2", err)
		t.FailNow()
	}
	err = os.Mkdir("sstables/sst1/sst2", 0755)
	if err != nil {
		fmt.Print("error creating directory 2", err)
		t.FailNow()
	}
	thirdTable, err := os.Create("sstables/sst1/sst2/test3.bin")
	if err != nil {
		fmt.Print("error creating file 3", err)
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

	checkpoint, err := CreateCheckpoint("sstables", "checkpoint1")
	if err != nil {
		fmt.Print("error creating checkpoint 1", err)
		t.FailNow()
	}

	firstLink, err := checkpoint.OpenFileReadWrite("test1.bin")
	if err != nil {
		fmt.Print("error opening link file 1", err)
		t.FailNow()
	}
	secondLink, err := checkpoint.OpenFileReadWrite("sst1/test2.bin")
	if err != nil {
		fmt.Print("error opening link file 2", err)
		t.FailNow()
	}
	thirdLink, err := checkpoint.OpenFileReadWrite("sst1/sst2/test3.bin")
	if err != nil {
		fmt.Print("error opening link file 3", err)
		t.FailNow()
	}

	readData, err := bm.Get(firstLink, 0)
	if err != nil {
		fmt.Print("error getting from link file 1", err)
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
		fmt.Print("error getting from link file 2", err)
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
		fmt.Print("error getting from link file 3", err)
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
		fmt.Print("error closing file 1", err)
		t.FailNow()
	}
	err = secondTable.Close()
	if err != nil {
		fmt.Print("error closing file 2", err)
		t.FailNow()
	}
	err = thirdTable.Close()
	if err != nil {
		fmt.Print("error closing file 3", err)
		t.FailNow()
	}
	err = firstLink.Close()
	if err != nil {
		fmt.Print("error closing link 1", err)
		t.FailNow()
	}
	err = secondLink.Close()
	if err != nil {
		fmt.Print("error closing link 2", err)
		t.FailNow()
	}
	err = thirdLink.Close()
	if err != nil {
		fmt.Print("error closing link 3", err)
		t.FailNow()
	}
	err = os.RemoveAll("sstables")
	if err != nil {
		fmt.Print("error emptying directory", err)
		t.FailNow()
	}
	err = os.Mkdir("sstables", 0755)
	if err != nil {
		fmt.Print("error creating directory sstables", err)
		t.FailNow()
	}
	checkpoint.Delete()

	err = os.Remove("checkpoints")
	if err != nil {
		fmt.Print("error deleting checkpoint", err)
		t.FailNow()
	}
	err = os.Remove("sstables")
	if err != nil {
		fmt.Print("error deleting checkpoint", err)
		t.FailNow()
	}
}

// testiranje funkcionalnosti CheckpointManager klase
func TestCheckpointManager(t *testing.T) {
	err := os.Mkdir("sstables", 0755)
	if err != nil {
		fmt.Print("error creating directory ", err)
		t.FailNow()
	}
	err = os.Mkdir("checkpoints", 0755)
	if err != nil {
		fmt.Print("error creating directory ", err)
		t.FailNow()
	}

	err = os.Mkdir("checkpoints/ch1", 0755)
	if err != nil {
		fmt.Print("error creating directory 1", err)
		t.FailNow()
	}
	ch1, err := os.Create("checkpoints/ch1/test1.bin")
	if err != nil {
		fmt.Print("error creating file 1", err)
		t.FailNow()
	}
	err = os.Mkdir("checkpoints/ch1/sst1", 0755)
	if err != nil {
		fmt.Print("error creating directory 1", err)
		t.FailNow()
	}
	ch2, err := os.Create("checkpoints/ch1/sst1/test2.bin")
	if err != nil {
		fmt.Print("error creating file 2", err)
		t.FailNow()
	}
	err = os.Mkdir("checkpoints/ch1/sst1/sst2", 0755)
	if err != nil {
		fmt.Print("error creating directory 2", err)
		t.FailNow()
	}
	ch3, err := os.Create("checkpoints/ch1/sst1/sst2/test3.bin")
	if err != nil {
		fmt.Print("error creating file 3", err)
		t.FailNow()
	}
	err = os.Mkdir("sstables/sst3", 0755)
	if err != nil {
		fmt.Print("error creating directory 4", err)
		t.FailNow()
	}
	firstTable, err := os.Create("sstables/sst3/test1.bin")
	if err != nil {
		fmt.Print("error creating file 3", err)
		t.FailNow()
	}
	// instanciramo checkpointmanager,
	// sa ocekivanjem da ce pratiti sta je stavljeno u checkpoint folder
	checkpointManager, err := NewCheckpointManager()
	if err != nil {
		fmt.Print("error creating manager ", err)
		t.FailNow()
	}
	// ocekujemo da manager prepozna postojanje checkpoint-a
	Checkpoint, err := checkpointManager.GetCheckpoint("ch1")
	if err != nil {
		fmt.Print("error getting from manager ", err)
		t.FailNow()
	}
	// dodajemo checkpoint-ocekujemo da se kreiraju direktorijumi
	err = checkpointManager.AddCheckpoint("sstables/sst3", "ch2")
	if err != nil {
		fmt.Print("error creating checkpoint ", err)
		t.FailNow()
	}
	// nov checkpoint se getuje
	_, err = checkpointManager.GetCheckpoint("ch2")
	if err != nil {
		fmt.Print("error getting checkpoint 2", err)
		t.FailNow()
	}

	checkpointList := checkpointManager.GetCheckpointList()
	if checkpointList.Len() != 2 {
		fmt.Print("error getting checkpoint list ")
		t.FailNow()
	}

	err = firstTable.Close()
	if err != nil {
		fmt.Print("error closing file ", err)
		t.FailNow()
	}
	err = ch1.Close()
	if err != nil {
		fmt.Print("error closing checkpoint 1 ", err)
		t.FailNow()
	}
	err = ch2.Close()
	if err != nil {
		fmt.Print("error closing checkpoint 2 ", err)
		t.FailNow()
	}
	err = ch3.Close()
	if err != nil {
		fmt.Print("error closing checkpoint 3 ", err)
		t.FailNow()
	}
	err = os.RemoveAll("sstables/sst3")
	if err != nil {
		fmt.Print("error resetting directory ", err)
		t.FailNow()
	}
	// brisanje oba checkpoint-a, ocekujemo da radi kako treba ako checkpoinmanager adekvatno prati
	err = Checkpoint.Delete()
	if err != nil {
		fmt.Print("error deleting checkpoint 1", err)
		t.FailNow()
	}
	err = checkpointManager.DeleteCheckpoint("ch2")
	if err != nil {
		fmt.Print("error deleting checkpoint 2", err)
		t.FailNow()
	}

	err = os.Remove("checkpoints")
	if err != nil {
		fmt.Print("error deleting checkpoint", err)
		t.FailNow()
	}
	err = os.Remove("sstables")
	if err != nil {
		fmt.Print("error deleting checkpoint", err)
		t.FailNow()
	}
}
