package bloom_filter

import (
	"encoding/binary"
	"reflect"
	"testing"
)

// TODO: Consider testing out BloomFilters with varying lengths
//
//	See: https://github.com/google/leveldb/blob/main/util/bloom_test.cc
var LENGTH uint = 1000
var FALSE_POSITIVE_RATE = 0.05

func TestNotFound(t *testing.T) {
	bf, err := NewBloomFilter(LENGTH, FALSE_POSITIVE_RATE)
	if err != nil {
		t.Errorf("NewBloomFilter returned an error: %s", err)
	}
	bf.Set([]byte("abc"))
	bf.Set([]byte("123"))
	bf.Set([]byte("xyz"))
	if bf.IsFound([]byte("pqr")) {
		t.Errorf("Didn't set key 'pqr' but found")
	}
	if bf.IsFound([]byte("abd")) {
		t.Errorf("Didn't set key 'abd' but found")
	}
	if bf.IsFound([]byte("021")) {
		t.Errorf("Didn't set key '021' but found")
	}
}

func TestFound(t *testing.T) {
	data := make([]byte, 4)
	for range LENGTH {
		bf, _ := NewBloomFilter(LENGTH, FALSE_POSITIVE_RATE)
		for i := range uint32(LENGTH) {
			binary.BigEndian.PutUint32(data, i)
			bf.Set(data)
		}
		false_positives := 0
		for i := range uint32(LENGTH) {
			binary.BigEndian.PutUint32(data, i+10000)
			if bf.IsFound(data) {
				false_positives++
			}
		}
		// NOTE: As a probabilistic structure, there is always variance, using 1.6 to account for that
		if false_positives > int(float64(LENGTH)*FALSE_POSITIVE_RATE*1.6) {
			t.Errorf("There are %d false-positives", false_positives)
		}
	}
}

func TestSerialize(t *testing.T) {
	bf, _ := NewBloomFilter(LENGTH, FALSE_POSITIVE_RATE)
	data := bf.Dump()
	loaded := LoadBloomFilter(data)
	if !reflect.DeepEqual(bf, loaded) {
		t.Error("Serialized then deserialized is not the same as the starting BloomFilter. Expected", bf, "but was", loaded)
	}
}
