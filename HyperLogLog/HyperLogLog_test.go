package HyperLogLog

import (
	"reflect"
	"testing"
)

var PRECISION uint8 = 5

func TestEstimate(t *testing.T) {
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

func TestSerializeDeserialize(t *testing.T) {
	hll, err := NewHyperLogLog(PRECISION)
	if err != nil {
		t.Errorf("HyperLogLog returned error while constructing")
	}
	hll.Add("abc")
	hll.Add("dfg")
	hll.Add("sng")

	hllBytes, err := hll.Serialize()
	if err != nil {
		t.Errorf("Error while serializing")
	}
	hll2, err := Deserialize(hllBytes)
	if err != nil {
		t.Errorf("Error while deserializing")
	}
	if !reflect.DeepEqual(hll, hll2) {
		t.Errorf("HyperLogLog is not equal before and after serialization")
	}
}
