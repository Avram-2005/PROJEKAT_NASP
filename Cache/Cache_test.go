package Cache

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

func TestCache(t *testing.T) {
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

	sst.Put("d1", data1)
	sst.Put("d2", data2)
	sst.Put("d3", data3)
	sst.Put("d4", data4)
	sst.Put("d5", data5)
	sst.Put("d6", data6)

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
}
