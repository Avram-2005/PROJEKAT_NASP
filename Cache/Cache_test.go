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
		t.Errorf("greska tokom stvaranja kesa")
		t.FailNow()
	}
	sst := NewSSTable()
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
	sst.Put("d1", data2)
	sst.Put("d2", data3)
	sst.Put("d3", data4)
	sst.Put("d4", data5)
	sst.Put("d5", data6)
	sst.Put("d6", data1)
	//nove informacije idu direktno u cache
	ch.Put("d1", &data1)
	ch.Put("d2", &data2)
	ch.Put("d3", &data3)
	ch.Put("d4", &data4)
	ch.Put("d5", &data5)
	ch.Put("d6", &data6)
	//proveravamo ispravnost informacija
	read1, err := ch.Get("d1", sst)
	if err != nil {
		fmt.Print(err)

		t.Errorf("greska tokom prvog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, (*read1)) {
		fmt.Print(data1, (*read1))
		t.Errorf("data1 nije isti pri citanju")
	}
	read2, err := ch.Get("d2", sst)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data2, (*read2)) {
		t.Errorf("data2 nije isti pri citanju")
	}
	read3, err := ch.Get("d3", sst)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data3, (*read3)) {
		t.Errorf("data3 nije isti pri citanju")
	}
	read4, err := ch.Get("d4", sst)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom cetvrtog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data4, (*read4)) {
		t.Errorf("data4 nije isti pri citanju")
	}
	read5, err := ch.Get("d5", sst)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom petog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data5, (*read5)) {
		t.Errorf("data5 nije isti pri citanju")
	}
	read6, err := ch.Get("d6", sst)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom sestog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data6, (*read6)) {
		t.Errorf("data6 nije isti pri citanju")
	}
	if ch.currentSize != 6 {
		fmt.Print(ch.currentSize)
		t.Errorf("cache is wrong size")
	}
}

func TestCacheTable(t *testing.T) {
	//proveravamo da li cache ispravno dobavlja vrednosti iz sstable strukture
	ch, err := NewCache(8)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom stvaranja kesa")
		t.FailNow()
	}
	sst := NewSSTable()
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

	sst.Put("d1", data6)
	sst.Put("d2", data2)
	sst.Put("d3", data3)
	sst.Put("d4", data4)
	sst.Put("d5", data5)
	sst.Put("d6", data6)

	ch.Put("d1", &data1)
	read0, err := ch.Get("d1", sst)
	if err != nil {
		fmt.Print(err)

		t.Errorf("greska tokom prvog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, (*read0)) {
		fmt.Print(data1, read0)
		t.Errorf("data1 nije isti pri citanju")
	}

	read1, err := ch.Get("d1", sst)
	if err != nil {
		fmt.Print(err)

		t.Errorf("greska tokom prvog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, (*read1)) {
		fmt.Print(data1, (*read1))
		t.Errorf("data1 nije isti pri citanju")
	}
	read2, err := ch.Get("d2", sst)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data2, (*read2)) {
		t.Errorf("data2 nije isti pri citanju")
	}
	read3, err := ch.Get("d3", sst)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data3, (*read3)) {
		t.Errorf("data3 nije isti pri citanju")
	}
	read4, err := ch.Get("d4", sst)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom cetvrtog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data4, (*read4)) {
		t.Errorf("data4 nije isti pri citanju")
	}
	read5, err := ch.Get("d5", sst)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom petog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data5, (*read5)) {
		t.Errorf("data5 nije isti pri citanju")
	}
	read6, err := ch.Get("d6", sst)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom sestog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data6, (*read6)) {
		t.Errorf("data6 nije isti pri citanju")
	}
	//sada cache dobavlja iz sopstvene cache mape vrednost-proveravamo ispravnost i velicinu
	read6, err = ch.Get("d6", sst)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom sestog citanja")
		t.FailNow()
	}
	if !reflect.DeepEqual(data6, (*read6)) {
		t.Errorf("data6 nije isti pri citanju")
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
		t.Errorf("greska tokom stvaranja kesa")
		t.FailNow()
	}
	sst := NewSSTable()
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
	sst.Put("d1", data6)
	read1, err := ch.Get("d1", sst)
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	if !reflect.DeepEqual(data6, (*read1)) {
		fmt.Print(data6, (*read1))
		t.Errorf("data6 nije isti pri citanju")
		t.FailNow()
	}
}
