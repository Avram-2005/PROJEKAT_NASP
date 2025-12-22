package projekatnasp

import (
	"fmt"
	"math"
	"math/bits"
)

type hyperLogLog struct {
	m       uint64 //hyperLogLog set size
	p       uint8  //hyperLogLog precision
	buckets []uint8
}

// constructor
func NewHyperLogLog(p uint8) (*hyperLogLog, error) {
	hll := hyperLogLog{}
	//preciznost je ocekivana da bude unutar odredjenog opsega, ako nije, baca error
	if p < HLL_MIN_PRECISION || p > HLL_MAX_PRECISION {
		return &hll, fmt.Errorf("hyperLogLog precision should be between 4 and 16")
	}
	hll.p = p
	hll.m = uint64(math.Pow(2, float64(hll.p)))
	hll.buckets = make([]uint8, 0)
	return &hll, nil
}

// helper functions :)
const (
	HLL_MIN_PRECISION = 4
	HLL_MAX_PRECISION = 16
)

func firstKbits(value, k uint64) uint64 {
	return value >> (64 - k)
}

func trailingZeroBits(value uint64) int {
	return bits.TrailingZeros64(value)
}

func (hll *hyperLogLog) Estimate() float64 {
	sum := 0.0
	for _, val := range hll.buckets {
		sum += math.Pow(math.Pow(2.0, float64(val)), -1)
	}

	alpha := 0.7213 / (1.0 + 1.079/float64(hll.m))
	estimation := alpha * math.Pow(float64(hll.m), 2.0) / sum
	emptyRegs := hll.emptyCount()
	if estimation <= 2.5*float64(hll.m) { // do small range correction
		if emptyRegs > 0 {
			estimation = float64(hll.m) * math.Log(float64(hll.m)/float64(emptyRegs))
		}
	} else if estimation > 1/30.0*math.Pow(2.0, 32.0) { // do large range correction
		estimation = -math.Pow(2.0, 32.0) * math.Log(1.0-estimation/math.Pow(2.0, 32.0))
	}
	return estimation
}

func (hll *hyperLogLog) emptyCount() int {
	sum := 0
	for _, val := range hll.buckets {
		if val == 0 {
			sum++
		}
	}
	return sum
}
