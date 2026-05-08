package BlockManager

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"testing"
)

// Ovi testovi su direktna kopija testova iz bufferpoola,
// koji proveravaju samo da li blockmanager zapravo funkcionise kao interfejs sa bufferpool-om
func TestBlockManager(t *testing.T) {
	file, err := os.Create("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during file creation")
		t.FailNow()
	}

	// prvo pravimo cetiri niza bajtova, koje cemo da ubacujemo u fajl
	data1 := make([]byte, 4096)
	binary.BigEndian.PutUint64(data1, 78)

	data2 := make([]byte, 4096)
	binary.BigEndian.PutUint32(data2, 56)

	data3 := make([]byte, 4096)
	binary.BigEndian.PutUint16(data3, 67)

	// inicijalizacija bufferpool-a, ciji kes je duzine 3, a velicna blokova 4
	bm, err := NewBlockManager(3, 4)
	if err != nil {
		t.Errorf("error during blockmanager initialization")
		t.FailNow()
	}

	// ubacujemo prvi niz bajtova u prvi blok fajla test.bin
	err = bm.Put(file, 0, &data1)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during first write")
		t.FailNow()
	}

	// citamo prvi blok fajla test.bin
	readData1, err := bm.Get(file, 0)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during first read")
		t.FailNow()
	}

	// proveravamo da li je procitana vrednost jednaka sa upisanom vrednoscu
	if !reflect.DeepEqual(data1, (*readData1)) {
		fmt.Print(data1[:10])
		fmt.Print((*readData1)[:10])
		t.Errorf("data1 is not the same before and after reading")
		t.FailNow()
	}

	// ubacujemo drugi niz bajtova
	err = bm.Put(file, 1, &data2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during second write")
		t.FailNow()
	}

	// citamo drugi blok fajla
	readData2, err := bm.Get(file, 1)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during second read")
		t.FailNow()
	}

	// proveravamo da li je upisano i procitano isto
	if !reflect.DeepEqual(data2, (*readData2)) {
		fmt.Print(data2[:10])
		fmt.Print((*readData2)[:10])
		t.Errorf("data2 is not the same before and after reading")
		t.FailNow()
	}

	// upisujemo podatke u treci blok
	err = bm.Put(file, 2, &data3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during third write")
		t.FailNow()
	}

	// citamo treci blok
	readData3, err := bm.Get(file, 2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during third read")
		t.FailNow()
	}

	// proveravamo da li je upisano i procitano isto
	if !reflect.DeepEqual(data3, (*readData3)) {
		fmt.Print(data3[:10])
		fmt.Print((*readData3)[:10])
		t.Errorf("data3 is not the same before and after reading")
		t.FailNow()
	}

	data4 := make([]byte, 4096)
	binary.BigEndian.PutUint32(data4, 45)

	// velicina buffer-a je 3, a sad upisujemo cetvrtu stvar-prva stvar u bufferu bi trebala da se izbaci!
	err = bm.Put(file, 3, &data4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during fourth write")
		t.FailNow()
	}

	readData4, err := bm.Get(file, 3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during fourth read")
		t.FailNow()
	}

	// proveravamo da li je upisano i procitano isto
	if !reflect.DeepEqual(data4, (*readData4)) {
		fmt.Print(data4[:10])
		fmt.Print((*readData4)[:10])
		t.Errorf("data4 is not the same before and after reading")
		t.FailNow()
	}

	readData1, err = bm.Get(file, 0)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during fifth read")
		t.FailNow()
	}

	if !reflect.DeepEqual(data1, (*readData1)) {
		fmt.Print(data1[:10])
		fmt.Print((*readData1)[:10])
		t.Errorf("data1 is not the same before and after second reading")
		t.FailNow()
	}

	err = file.Close()
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during file closing")
		t.FailNow()
	}

	err = os.Remove("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during file deletion")
		t.FailNow()
	}
}

func TestSpecific(t *testing.T) {
	file, err := os.Create("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during file creation")
		t.FailNow()
	}

	// prvo pravimo cetiri niza bajtova, koje cemo da ubacujemo u fajl
	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)

	data2 := make([]byte, 100)
	binary.BigEndian.PutUint32(data2, 56)

	data3 := make([]byte, 100)
	binary.BigEndian.PutUint16(data3, 67)

	// inicijalizacija bufferpool-a, ciji kes je duzine 3, a velicna blokova 4
	bm, err := NewBlockManager(2, 4)
	if err != nil {
		t.Errorf("error during blockmanager initialization")
		t.FailNow()
	}

	err = bm.PutSpecific(file, 0, 0, 100, &data1)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during first write")
		t.FailNow()
	}

	err = bm.PutSpecific(file, 0, 100, 100, &data2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during second write")
		t.FailNow()
	}

	err = bm.PutSpecific(file, 0, 200, 100, &data3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during third write")
		t.FailNow()
	}

	readData1, err := bm.GetSpecific(file, 0, 0, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during first read")
		t.FailNow()
	}

	if !reflect.DeepEqual(data1, (*readData1)) {
		fmt.Print(data1)
		fmt.Print(readData1)
		t.Errorf("data1 is not the same before and after writing")
		t.FailNow()
	}

	readData2, err := bm.GetSpecific(file, 0, 100, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during second read")
		t.FailNow()
	}

	if !reflect.DeepEqual(data2, (*readData2)) {
		t.Errorf("data2 is not the same before and after writing")
		t.FailNow()
	}

	readData3, err := bm.GetSpecific(file, 0, 200, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during third read")
		t.FailNow()
	}

	if !reflect.DeepEqual(data3, (*readData3)) {
		t.Errorf("data3 is not the same before and after writing")
		t.FailNow()
	}

	data4 := make([]byte, 100)
	binary.BigEndian.PutUint16(data4, 12)

	// Inicijalizacija novog blockmanager-a, cime simuliramo prestanak rada aplikacije i ponovno pokretanje
	bm, err = NewBlockManager(2, 4)
	if err != nil {
		t.Errorf("error during second blockmanager initialization")
		t.FailNow()
	}

	err = bm.PutSpecific(file, 0, 150, 100, &data4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during fourth write")
		t.FailNow()
	}

	data5 := make([]byte, 100)
	binary.BigEndian.PutUint16(data5, 89)

	err = bm.PutSpecific(file, 0, 250, 100, &data5)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during fifth write")
		t.FailNow()
	}

	readData4, err := bm.GetSpecific(file, 0, 150, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during fourth read")
		t.FailNow()
	}

	if !reflect.DeepEqual(data4, (*readData4)) {
		t.Errorf("data4 is not the same before and after writing")
		t.FailNow()
	}

	readData5, err := bm.GetSpecific(file, 0, 250, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during fifth read")
		t.FailNow()
	}

	if !reflect.DeepEqual(data5, (*readData5)) {
		t.Errorf("data5 is not the same before and after writing")
		t.FailNow()
	}

	readData2, err = bm.GetSpecific(file, 0, 100, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during sixth read")
		t.FailNow()
	}

	if reflect.DeepEqual(data2, (*readData2)) {
		t.Errorf("data2 is the same before and after writing, but it should not be")
		t.FailNow()
	}

	err = file.Close()
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during file closing")
		t.FailNow()
	}

	err = os.Remove("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during file deletion")
		t.FailNow()
	}
}

func TestAddBuffer(t *testing.T) {
	file, err := os.Create("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during file creation")
		t.FailNow()
	}

	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)

	data2 := make([]byte, 100)
	binary.BigEndian.PutUint32(data2, 56)

	data3 := make([]byte, 100)
	binary.BigEndian.PutUint16(data3, 67)

	// inicijalizacija bufferpool-a, ciji kes je duzine 3, a velicna blokova 4
	bm, err := NewBlockManager(2, 4)
	if err != nil {
		t.Errorf("error during blockmanager initialization")
		t.FailNow()
	}

	err = bm.PutSpecific(file, 0, 0, 100, &data1)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during first write")
		t.FailNow()
	}

	err = bm.PutSpecific(file, 0, 100, 100, &data2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during second write")
		t.FailNow()
	}

	err = bm.PutSpecific(file, 0, 200, 100, &data3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during third write")
		t.FailNow()
	}

	bm.AddBuffer(file, 0)

	remainingZeroes := make([]byte, 3796)
	data3 = append(data3, remainingZeroes...)
	data2 = append(data2, data3...)
	data1 = append(data1, data2...)

	data4, err := bm.Get(file, 0)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during buffer addition")
		t.FailNow()
	}

	if !reflect.DeepEqual(data1, (*data4)) {
		fmt.Print(data1)
		fmt.Print("SECOND VALUE:")
		fmt.Print(data4)
		t.Errorf("unexpected value")
		t.FailNow()
	}

	err = file.Close()
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during file closing")
		t.FailNow()
	}

	err = os.Remove("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during file deletion")
		t.FailNow()
	}
}

func TestGetBlockSize(t *testing.T) {
	bm1, err := NewBlockManager(2, 4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during first blockmanager initialization")
		t.FailNow()
	}

	bm2, err := NewBlockManager(2, 8)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during second blockmanager initialization")
		t.FailNow()
	}

	bm3, err := NewBlockManager(2, 16)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during third blockmanager initialization")
		t.FailNow()
	}

	size1 := bm1.GetBlockSize()
	size2 := bm2.GetBlockSize()
	size3 := bm3.GetBlockSize()

	if size1 != 4096 {
		t.Errorf("unexpected block size for the first blockmanager")
		t.FailNow()
	}

	if size2 != 8192 {
		t.Errorf("unexpected block size for the second blockmanager")
		t.FailNow()
	}

	if size3 != 16384 {
		t.Errorf("unexpected block size for the third blockmanager")
		t.FailNow()
	}
}
