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
		t.Errorf("greska tokom stvaranja fajla")
		t.FailNow()
	}
	//prvo pravimo cetiri niza bajtova, koje cemo da ubacujemo u fajl
	data1 := make([]byte, 4096)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 4096)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 4096)
	binary.BigEndian.PutUint16(data3, 67)
	//inicijalizacija bufferpool-a, ciji kes je duzine 3, a velicna blokova 4
	bm, err := NewBlockManager(3, 4)
	if err != nil {
		t.Errorf("greska tokom inicalizacije blockmanagera")
		t.FailNow()
	}
	//ubacujemo prvi niz bajtova u prvi blok fajla test.bin
	err = bm.Put(file, 0, &data1)

	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog pisanja")
		t.FailNow()
	}
	//citamo prvi blok fajla test.bin
	readData1, err := bm.Get(file, 0)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog citanja")
		t.FailNow()
	}
	//proveravamo da li je procitana vrednost jednaka sa upisanom vrednoscu
	if !reflect.DeepEqual(data1, (*readData1)) {
		fmt.Print(data1[:10])
		fmt.Print((*readData1)[:10])
		t.Errorf("data1 nije isto pre i posle citanja")
		t.FailNow()
	}
	//ubacujemo drugi niz bajtova
	err = bm.Put(file, 1, &data2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog pisanja")
		t.FailNow()
	}
	//citamo drugi blok fajla
	readData2, err := bm.Get(file, 1)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog citanja")
		t.FailNow()
	}
	//proveravamo da li je upisano i procitano isto
	if !reflect.DeepEqual(data2, (*readData2)) {
		fmt.Print(data2[:10])
		fmt.Print((*readData2)[:10])
		t.Errorf("data2 nije isto pre i posle citanja")
		t.FailNow()
	}
	//upisujemo podatke u treci blok
	err = bm.Put(file, 2, &data3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg pisanja")
		t.FailNow()
	}
	//citamo treci blok
	readData3, err := bm.Get(file, 2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg citanja")
		t.FailNow()
	}
	//proveravamo da li je upisano i procitano isto
	if !reflect.DeepEqual(data3, (*readData3)) {
		fmt.Print(data3[:10])
		fmt.Print((*readData3)[:10])
		t.Errorf("data3 nije isto pre i posle citanja")
		t.FailNow()
	}
	data4 := make([]byte, 4096)
	binary.BigEndian.PutUint32(data4, 45)
	//velicina buffer-a je 3, a sad upisujemo cetvrtu stvar-prva stvar u bufferu bi trebala da se izbaci!
	err = bm.Put(file, 3, &data4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom cetvrtog pisanja")
		t.FailNow()
	}

	readData4, err := bm.Get(file, 3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom cetvrtog citanja")
		t.FailNow()
	}
	//proveravamo da li je upisano i procitano isto
	if !reflect.DeepEqual(data4, (*readData4)) {
		fmt.Print(data4[:10])
		fmt.Print((*readData4)[:10])
		t.Errorf("data4 nije isto pre i posle citanja")
		t.FailNow()
	}
	readData1, err = bm.Get(file, 0)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom petog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, (*readData1)) {
		fmt.Print(data1[:10])
		fmt.Print((*readData1)[:10])
		t.Errorf("data1 nije isto pre i posle drugog citanja")
		t.FailNow()
	}
	err = file.Close()
	if err != nil {
		fmt.Print(err)
		t.Errorf("Greska tokom zatvaranja fajla")
		t.FailNow()
	}

	err = os.Remove("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("Greska tokom brisanja fajla")
		t.FailNow()
	}
}

func TestSpecific(t *testing.T) {
	file, err := os.Create("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom stvaranja fajla")
		t.FailNow()
	}
	//prvo pravimo cetiri niza bajtova, koje cemo da ubacujemo u fajl
	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 100)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 100)
	binary.BigEndian.PutUint16(data3, 67)
	//inicijalizacija bufferpool-a, ciji kes je duzine 3, a velicna blokova 4
	bm, err := NewBlockManager(2, 4)
	if err != nil {
		t.Errorf("greska tokom inicalizacije blockmanagera")
		t.FailNow()
	}
	err = bm.PutSpecific(file, 0, 0, 100, &data1)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog upisivanja")
		t.FailNow()
	}
	err = bm.PutSpecific(file, 0, 100, 100, &data2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog upisivanja")
		t.FailNow()
	}
	err = bm.PutSpecific(file, 0, 200, 100, &data3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg upisivanja")
		t.FailNow()
	}

	readData1, err := bm.GetSpecific(file, 0, 0, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog citanjaa")
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, (*readData1)) {
		fmt.Print(data1)
		fmt.Print(readData1)
		t.Errorf("data1 nije isti pre i posle pisanja")
		t.FailNow()
	}

	readData2, err := bm.GetSpecific(file, 0, 100, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog citanjaa")
		t.FailNow()
	}
	if !reflect.DeepEqual(data2, (*readData2)) {
		t.Errorf("data2 nije isti pre i posle pisanja")
		t.FailNow()
	}

	readData3, err := bm.GetSpecific(file, 0, 200, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg citanjaa")
		t.FailNow()
	}
	if !reflect.DeepEqual(data3, (*readData3)) {
		t.Errorf("data3 nije isti pre i posle pisanja")
		t.FailNow()
	}

	data4 := make([]byte, 100)
	binary.BigEndian.PutUint16(data4, 12)
	// Inicijalizacija novog blockmanager-a, cime simuliramo prestanak rada aplikacije i ponovno pokretanje
	bm, err = NewBlockManager(2, 4)
	if err != nil {
		t.Errorf("greska tokom inicalizacije drugog blockmanagera")
		t.FailNow()
	}

	err = bm.PutSpecific(file, 0, 150, 100, &data4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom cetvrtog upisivanja")
		t.FailNow()
	}

	data5 := make([]byte, 100)
	binary.BigEndian.PutUint16(data5, 89)

	err = bm.PutSpecific(file, 0, 250, 100, &data5)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom petog upisivanja")
		t.FailNow()
	}

	readData4, err := bm.GetSpecific(file, 0, 150, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom cetvrtog citanjaa")
		t.FailNow()
	}
	if !reflect.DeepEqual(data4, (*readData4)) {
		t.Errorf("data4 nije isti pre i posle pisanja")
		t.FailNow()
	}

	readData5, err := bm.GetSpecific(file, 0, 250, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom petog citanjaa")
		t.FailNow()
	}
	if !reflect.DeepEqual(data5, (*readData5)) {
		t.Errorf("data5 nije isti pre i posle pisanja")
		t.FailNow()
	}

	readData2, err = bm.GetSpecific(file, 0, 100, 100)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom sestog citanjaa")
		t.FailNow()
	}
	if reflect.DeepEqual(data2, (*readData2)) {
		t.Errorf("data2 je isti pre i posle pisanja, a ne bi trebao da bude")
		t.FailNow()
	}
	err = file.Close()
	if err != nil {
		fmt.Print(err)
		t.Errorf("Greska tokom zatvaranja fajla")
		t.FailNow()
	}

	err = os.Remove("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("Greska tokom brisanja fajla")
		t.FailNow()
	}
}

func TestAddBuffer(t *testing.T) {
	file, err := os.Create("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom stvaranja fajla")
		t.FailNow()
	}
	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 100)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 100)
	binary.BigEndian.PutUint16(data3, 67)
	//inicijalizacija bufferpool-a, ciji kes je duzine 3, a velicna blokova 4
	bm, err := NewBlockManager(2, 4)
	if err != nil {
		t.Errorf("greska tokom inicalizacije blockmanagera")
		t.FailNow()
	}
	err = bm.PutSpecific(file, 0, 0, 100, &data1)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog upisivanja")
		t.FailNow()
	}
	err = bm.PutSpecific(file, 0, 100, 100, &data2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog upisivanja")
		t.FailNow()
	}
	err = bm.PutSpecific(file, 0, 200, 100, &data3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg upisivanja")
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
		t.Errorf("greska tokom dodavanja bafera")
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, (*data4)) {
		fmt.Print(data1)
		fmt.Print("DRUGA VREDNOST:")
		fmt.Print(data4)
		t.Errorf("vrednost je neocekivana")
		t.FailNow()
	}
	err = file.Close()
	if err != nil {
		fmt.Print(err)
		t.Errorf("Greska tokom zatvaranja fajla")
		t.FailNow()
	}

	err = os.Remove("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("Greska tokom brisanja fajla")
		t.FailNow()
	}
}

func TestGetBlockSize(t *testing.T) {
	bm1, err := NewBlockManager(2, 4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom inicijalizacije prvog blockmanager-a")
		t.FailNow()
	}
	bm2, err := NewBlockManager(2, 8)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom inicijalizacije drugog blockmanager-a")
		t.FailNow()
	}
	bm3, err := NewBlockManager(2, 16)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom inicijalizacije trećeg blockmanager-a")
		t.FailNow()
	}
	size1 := bm1.GetBlockSize()
	size2 := bm2.GetBlockSize()
	size3 := bm3.GetBlockSize()
	if size1 != 4096 {
		t.Errorf("Velicina blokova prvog blockmanager-a je neocekivana")
		t.FailNow()
	}
	if size2 != 8192 {
		t.Errorf("Velicina blokova drugog blockmanager-a je neocekivana")
		t.FailNow()
	}
	if size3 != 16384 {
		t.Errorf("Velicina blokova trećeg blockmanager-a je neocekivana")
		t.FailNow()
	}
}
