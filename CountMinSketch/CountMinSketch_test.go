package CountMinSketch

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEsimate(t *testing.T) {
	cms, err := NewCountMinSketch(0.01, 0.01)
	if err != nil {
		fmt.Println("Greska pri kreiranju CMS: ", err)
		return
	}

	cms.Add("jabuka")
	cms.Add("jabuka")
	cms.Add("jabuka")
	cms.Add("jabuka")
	cms.Add("banana")
	cms.Add("banana")
	cms.Add("banana")
	cms.Add("kruska")

	fmt.Println("jabuka: ", cms.Estimate("jabuka"))
	fmt.Println("banana: ", cms.Estimate("banana"))
	fmt.Println("kruska: ", cms.Estimate("kruska"))
}

func TestSerializeDeserialize(t *testing.T) {
	cms, err := NewCountMinSketch(0.01, 0.01)
	if err != nil {
		fmt.Println("Greska pri kreiranju CMS ", err)
		return
	}

	cms.Add("jabuka")
	cms.Add("jabuka")
	cms.Add("banana")
	cms.Add("banana")
	cms.Add("kruska")

	data := cms.Serialize()
	cms2 := Deserialize(data)

	if !reflect.DeepEqual(cms, cms2) {
		t.Errorf("CountMinSkecth nije isti pre i posle serijalizacije")
	}

	if cms.Estimate("jabuka") != cms2.Estimate("jabuka") {
		t.Errorf("jabuka nema isti Estimate")
	}
}
