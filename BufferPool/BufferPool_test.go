package BufferPool

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

func TestBufferPool(t *testing.T) {
	data1 := make([]byte, 4096)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 4096)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 4096)
	binary.BigEndian.PutUint16(data3, 67)
	bp, err := NewBufferPool(3, 4)
	if err != nil {
		t.Errorf("greska tokom inicalizacije bufferpoola")
		t.FailNow()
	}

	err = bp.Put("test.bin", 1, &data1)

	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog pisanja")
		t.FailNow()
	}

	readData1, err := bp.Get("test.bin", 1)
	fmt.Print(readData1)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog citanja")
		t.FailNow()
	}

	if !reflect.DeepEqual(data1, (*readData1)) {
		fmt.Print(data1[:10])
		fmt.Print((*readData1)[:10])
		t.Errorf("data1 nije isto pre i posle citanja")
		t.FailNow()
	}

	err = bp.Put("test.bin", 2, &data2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog pisanja")
		t.FailNow()
	}

	readData2, err := bp.Get("test.bin", 2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog citanja")
		t.FailNow()
	}

	if !reflect.DeepEqual(data2, (*readData2)) {
		fmt.Print(data2[:10])
		fmt.Print((*readData2)[:10])
		t.Errorf("data2 nije isto pre i posle citanja")
		t.FailNow()
	}

	err = bp.Put("test.bin", 3, &data3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg pisanja")
		t.FailNow()
	}

	readData3, err := bp.Get("test.bin", 3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg citanja")
		t.FailNow()
	}

	if !reflect.DeepEqual(data3, (*readData3)) {
		fmt.Print(data3[:10])
		fmt.Print((*readData3)[:10])
		t.Errorf("data3 nije isto pre i posle citanja")
		t.FailNow()
	}

}
