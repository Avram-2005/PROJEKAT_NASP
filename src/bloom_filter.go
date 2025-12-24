package main

import (
	"crypto/md5"
	"encoding/binary"
	"math"
	"errors"
	"time"
)

type BloomFilter struct {
	// TODO: Reconsider the name
	bitfields []byte
}

func (bf BloomFilter) IsIn(data []byte) (bool, error) {
	return false, errors.New("not implemented")
}

func (bf BloomFilter) Set(data []byte) error {
	return errors.New("not implemented")
}

func New(expectedElements int, falsePositiveRate float64) (BloomFilter, error) {
	return BloomFilter{}, errors.New("not implemented")
}

func Dump(bf BloomFilter, filename string) (error) {
	return errors.New("not implemented")
}

func Load(filename string) (BloomFilter, error) {
	return BloomFilter{}, errors.New("not implemented")
}

// PERF: Save the masks maybe
func (bf BloomFilter) setBit(i uint64) {
	target := bf.bitfields[i/8]
	var mask byte = 1 << (i % 8)
	bf.bitfields[i/8] = target | mask
}

func (bf BloomFilter) getBit(i uint64) bool {
	target := bf.bitfields[i/8]
	var mask byte = 1 << (i % 8)
	return (target & mask) != 0
}

func calculateM(expectedElements int, falsePositiveRate float64) uint {
	return uint(math.Ceil(float64(expectedElements) * math.Abs(math.Log(falsePositiveRate)) / math.Pow(math.Log(2), float64(2))))
}

func calculateK(expectedElements int, m uint) uint {
	return uint(math.Ceil((float64(m) / float64(expectedElements)) * math.Log(2)))
}

type HashWithSeed struct {
	Seed []byte
}

func (h HashWithSeed) hash(data []byte) uint64 {
	fn := md5.New()
	fn.Write(append(data, h.Seed...))
	return binary.BigEndian.Uint64(fn.Sum(nil))
}

func createHashFunctions(k uint32) []HashWithSeed {
	h := make([]HashWithSeed, k)
	ts := uint32(time.Now().Unix())
	for i := range k {
		seed := make([]byte, 4)
		binary.BigEndian.PutUint32(seed, ts+i)
		hfn := HashWithSeed{Seed: seed}
		h[i] = hfn
	}
	return h
}

// TODO: Implement testing as well
