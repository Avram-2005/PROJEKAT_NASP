package BufferPool

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

func TestBufferPool(t *testing.T) {
	//prvo pravimo cetiri niza bajtova, koje cemo da ubacujemo u fajl
	data1 := make([]byte, 4096)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 4096)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 4096)
	binary.BigEndian.PutUint16(data3, 67)
	//inicijalizacija bufferpool-a, ciji kes je duzine 3, a velicna blokova 4
	bp, err := NewBufferPool(3, 4)
	if err != nil {
		t.Errorf("greska tokom inicalizacije bufferpoola")
		t.FailNow()
	}
	//ubacujemo prvi niz bajtova u prvi blok fajla test.bin
	err = bp.Put("test.bin", 1, &data1)

	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog pisanja")
		t.FailNow()
	}
	//citamo prvi blok fajla test.bin
	readData1, err := bp.Get("test.bin", 1)
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
	err = bp.Put("test.bin", 2, &data2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog pisanja")
		t.FailNow()
	}
	//citamo drugi blok fajla
	readData2, err := bp.Get("test.bin", 2)
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
	err = bp.Put("test.bin", 3, &data3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg pisanja")
		t.FailNow()
	}
	//citamo treci blok
	readData3, err := bp.Get("test.bin", 3)
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
	err = bp.Put("test.bin", 4, &data4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom cetvrtog pisanja")
		t.FailNow()
	}

	readData4, err := bp.Get("test.bin", 4)
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
	readData1, err = bp.Get("test.bin", 1)
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
}
