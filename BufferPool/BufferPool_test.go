package BufferPool

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestBufferPool(t *testing.T) {
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
	bp, err := NewBufferPool(3, 4)
	if err != nil {
		t.Errorf("greska tokom inicalizacije bufferpoola")
		t.FailNow()
	}

	//ubacujemo prvi niz bajtova u prvi blok fajla test.bin
	err = bp.Put(file, 0, &data1)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog pisanja")
		t.FailNow()
	}
	//ubacujemo drugi niz bajtova
	err = bp.Put(file, 1, &data2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog pisanja")
		t.FailNow()
	}
	//upisujemo podatke u treci blok
	err = bp.Put(file, 2, &data3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg pisanja")
		t.FailNow()
	}
	//citamo prvi blok fajla test.bin
	readData1, err := bp.Get(file, 0)
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
	//citamo drugi blok fajla
	readData2, err := bp.Get(file, 1)
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
	//citamo treci blok
	readData3, err := bp.Get(file, 2)
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
	err = bp.Put(file, 3, &data4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom cetvrtog pisanja")
		t.FailNow()
	}

	readData4, err := bp.Get(file, 3)
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
	readData1, err = bp.Get(file, 0)
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

func TestLru1(t *testing.T) {
	file, err := os.Create("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom stvaranja fajla")
		t.FailNow()
	}
	//prvo pravimo tri niza bajtova, koje cemo da ubacujemo u fajl
	data1 := make([]byte, 4096)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 4096)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 4096)
	binary.BigEndian.PutUint16(data3, 67)
	//inicijalizacija bufferpool-a, ciji kes je duzine 2, a velicna blokova 4
	bp, err := NewBufferPool(2, 4)
	if err != nil {
		t.Errorf("Greska pri inicijalizaciji bufferpool-a")
		t.FailNow()
	}
	bp.Put(file, 0, &data1)
	bp.Put(file, 1, &data2)
	bp.Put(file, 2, &data3)
	//Put-ovali smo 3 bloka, i zbog toga ocekujemo da back i front liste ne budu null
	check1 := bp.lruList.Back()
	check2 := bp.lruList.Front()

	if check2 == nil {
		t.Errorf("Front nije vratio nista")
		t.FailNow()
	}
	if check1 == nil {
		t.Errorf("Back nije vratio nista")
		t.FailNow()
	}
	//Izvlacimo podatke iz same mape
	//Check1 bi trebao da bude back liste, u ovom
	//slucaju kljuc bloka broj 2 toest data2
	data_from_map1 := bp.cacheMap[(*check1).Value.(string)]
	//ocekujemo da data2 i data izvucen iz mape budu isti, iz tog razloga
	if !reflect.DeepEqual(data2, data_from_map1) {
		fmt.Print(data2)
		fmt.Print((*check1))
		t.Errorf("neocekivana vrednost za back")
		t.FailNow()
	}
	//Izvlacimo podatke iz same mape
	//Check2 bi trebao da bude back liste, u ovom
	//slucaju kljuc bloka broj 3 toest data3
	data_from_map2 := bp.cacheMap[(*check2).Value.(string)]
	if !reflect.DeepEqual(data3, data_from_map2) {
		fmt.Print(data3)
		fmt.Print((*check2))
		t.Errorf("neocekivana vrednost za front")
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

// rad sa dva fajla, i bufferpoolom dovoljno velikim da svi blokovi staju
func TestBigBufferPool(t *testing.T) {
	file, err := os.Create("test.bin")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom stvaranja fajla")
		t.FailNow()
	}

	file2, err := os.Create("test.txt")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom stvaranja drugog fajla")
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
	bp, err := NewBufferPool(10, 4)
	if err != nil {
		t.Errorf("greska tokom inicalizacije bufferpoola")
		t.FailNow()
	}

	//ubacujemo prvi niz bajtova u prvi blok fajla test.bin
	err = bp.Put(file, 0, &data1)

	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog pisanja")
		t.FailNow()
	}
	//citamo prvi blok fajla test.bin
	readData1, err := bp.Get(file, 0)
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
	err = bp.Put(file, 1, &data2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog pisanja")
		t.FailNow()
	}
	//citamo drugi blok fajla
	readData2, err := bp.Get(file, 1)
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
	err = bp.Put(file, 2, &data3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg pisanja")
		t.FailNow()
	}
	//citamo treci blok
	readData3, err := bp.Get(file, 2)
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
	err = bp.Put(file, 3, &data4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom cetvrtog pisanja")
		t.FailNow()
	}

	readData4, err := bp.Get(file, 3)
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
	readData1, err = bp.Get(file, 0)
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
	data1 = []byte{'a', 'b', 'c', 'd', 'e'}
	err = bp.Put(file2, 0, &data1)

	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog pisanja")
		t.FailNow()
	}
	//citamo prvi blok fajla test.txt
	readData1, err = bp.Get(file2, 0)
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

	err = file2.Close()
	if err != nil {
		fmt.Print(err)
		t.Errorf("Greska tokom zatvaranja drugog fajla")
		t.FailNow()
	}

	err = os.Remove("test.txt")
	if err != nil {
		fmt.Print(err)
		t.Errorf("Greska tokom brisanja drugog fajla")
		t.FailNow()
	}
}

func TestLru2(t *testing.T) {
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
	bp, err := NewBufferPool(2, 4)
	if err != nil {
		t.Errorf("Greska pri inicijalizaciji bufferpool-a")
		t.FailNow()
	}
	bp.Put(file, 0, &data1)
	bp.Put(file, 1, &data2)
	bp.Put(file, 2, &data3)
	get_data2, _ := bp.Get(file, 1)
	get_data3, _ := bp.Get(file, 2)
	//Getovali smo block 3 i block 2-ocekujemo da back i front ne budu nil
	check1 := bp.lruList.Back()
	check2 := bp.lruList.Front()

	if check2 == nil {
		t.Errorf("Front nije vratio nista")
		t.FailNow()
	}
	if check1 == nil {
		t.Errorf("Back nije vratio nista")
		t.FailNow()
	}
	//Ocekujemo da podaci izvuceni iz mape budu jednaki drugom bloku-data2
	data_from_map1 := bp.cacheMap[(*check1).Value.(string)]
	if !reflect.DeepEqual((*get_data2), data_from_map1) {
		fmt.Print(get_data2)
		fmt.Print(data_from_map1)
		t.Errorf("neocekivana vrednost za back")
		t.FailNow()
	}
	data_from_map2 := bp.cacheMap[(*check2).Value.(string)]
	if !reflect.DeepEqual((*get_data3), data_from_map2) {
		fmt.Print(get_data3)
		fmt.Print(data_from_map2)
		t.Errorf("neocekivana vrednost za front")
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
