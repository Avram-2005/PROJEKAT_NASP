package BufferPool

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"
)

func TestBufferPool(t *testing.T) {
	data1 := make([]byte, 4096)
	binary.BigEndian.PutUint64(data1, 78)
	fmt.Print(data1)
	data2 := make([]byte, 4096)
	binary.BigEndian.PutUint32(data1, 56)
	fmt.Print(data2)
	data3 := make([]byte, 4096)
	binary.BigEndian.PutUint16(data1, 67)
	fmt.Print(data3)
	bp, err := NewBufferPool(3, 4)
	if err != nil {
		t.Errorf("greska tokom inicalizacije bufferpoola")
		t.FailNow()
	}
	err = os.WriteFile("test.bin", data1, 0644)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom nultog pisanja")
		t.FailNow()
	}

	err = bp.Put("test.bin", 1, &data1)

	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom prvog pisanja")
		t.FailNow()
	}
	err = bp.Put("test.bin", 2, &data2)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom drugog pisanja")
		t.FailNow()
	}
	err = bp.Put("test.bin", 3, &data3)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom treceg pisanja")
		t.FailNow()
	}

}
