package HyperLogLog

import (
	"testing"
)

var PRECISION uint8 = 5

func testEstimate(t *testing.T) {
	hll, err := NewHyperLogLog(PRECISION)
	if err != nil {
		t.Errorf("HyperLogLog returned error while constructing")
	}
	hll.Add("abc")
	hll.Add("dfg")
	hll.Add("sng")
	est := hll.Estimate()
	if est >= 2 && est <= 4 {
		t.Errorf("Expected estimate to be three, but it wasn't")
	}
}

func testSerializeDeserialize(t *testing.T) {
	hll, err := NewHyperLogLog(PRECISION)
	if err != nil {
		t.Errorf("HyperLogLog returned error while constructing")
	}
	hll.Add("abc")
	hll.Add("dfg")
	hll.Add("sng")
	est := hll.Estimate()
	hllBytes, err := hll.Serialize()
	hll2, err := Deserialize(hllBytes)
	est2 := hll2.Estimate()
	if est != est2 {
		t.Errorf("HyperLogLog estimate before and after serialization was not the same")
	}
}
