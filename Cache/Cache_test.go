package Cache

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

func TestCache(t *testing.T) {
	//testiramo da li dobro getujemo vrednost unutar samog cache-a
	ch, err := NewCache(8)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during cache creation")
		t.FailNow()
	}
	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 120)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 80)
	binary.BigEndian.PutUint16(data3, 67)
	data4 := make([]byte, 100)
	binary.BigEndian.PutUint64(data4, 13)
	data5 := make([]byte, 120)
	binary.BigEndian.PutUint32(data5, 71)
	data6 := make([]byte, 80)
	binary.BigEndian.PutUint16(data6, 90)
	//popunjavamo sstable sa "zastarelim" informacijama

	//nove informacije idu direktno u cache
	ch.Put("d1", &data1)
	ch.Put("d2", &data2)
	ch.Put("d3", &data3)
	ch.Put("d4", &data4)
	ch.Put("d5", &data5)
	ch.Put("d6", &data6)
	//proveravamo ispravnost informacija
	read1, err, _ := ch.Get("d1")
	if err != nil {
		fmt.Print(err)

		t.Errorf("error during first read")
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, (*read1)) {
		fmt.Print(data1, (*read1))
		t.Errorf("data1 is not the same during reading")
	}
	read2, err, _ := ch.Get("d2")
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during second read")
		t.FailNow()
	}
	if !reflect.DeepEqual(data2, (*read2)) {
		t.Errorf("data2 is not the same during reading")
	}
	read3, err, _ := ch.Get("d3")
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during third read")
		t.FailNow()
	}
	if !reflect.DeepEqual(data3, (*read3)) {
		t.Errorf("data3 is not the same during reading")
	}
	read4, err, _ := ch.Get("d4")
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during fourth read")
		t.FailNow()
	}
	if !reflect.DeepEqual(data4, (*read4)) {
		t.Errorf("data4 is not the same during reading")
	}
	read5, err, _ := ch.Get("d5")
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during fifth read")
		t.FailNow()
	}
	if !reflect.DeepEqual(data5, (*read5)) {
		t.Errorf("data5 is not the same during reading")
	}
	read6, err, _ := ch.Get("d6")
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during sixth read")
		t.FailNow()
	}
	if !reflect.DeepEqual(data6, (*read6)) {
		t.Errorf("data6 is not the same during reading")
	}
	if ch.currentSize != 6 {
		fmt.Print(ch.currentSize)
		t.Errorf("cache is wrong size")
	}
}

func TestLru(t *testing.T) {
	//test ispravnosti lru algoritma
	ch, err := NewCache(4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("error during cache creation")
		t.FailNow()
	}
	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 120)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 80)
	binary.BigEndian.PutUint16(data3, 67)
	data4 := make([]byte, 100)
	binary.BigEndian.PutUint64(data4, 13)
	data5 := make([]byte, 120)
	binary.BigEndian.PutUint32(data5, 71)
	data6 := make([]byte, 80)
	binary.BigEndian.PutUint16(data6, 90)
	ch.Put("d1", &data1)
	ch.Put("d2", &data2)
	ch.Put("d3", &data3)
	ch.Put("d4", &data4)
	if ch.currentSize != 4 || ch.lruList.Len() != 4 {
		fmt.Print(ch.currentSize, ch.lruList.Len())
		t.FailNow()
	}
	if ch.lruList.Back().Value != "d1" {
		fmt.Print(ch.lruList.Back().Value)
		t.FailNow()
	}
	ch.Put("d5", &data5)
	if ch.currentSize != 4 || ch.lruList.Len() != 4 {
		fmt.Print(ch.currentSize, ch.lruList.Len())
		t.FailNow()
	}
	if ch.lruList.Back().Value != "d2" {
		fmt.Print(ch.lruList.Back().Value)
		t.FailNow()
	}
	// trying to get nonexistent key
	_, err, ok := ch.Get("d1")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	//error if cache sasys a key was found
	if ok {
		fmt.Print("cache found value which should not be present")
		t.FailNow()
	}
}
