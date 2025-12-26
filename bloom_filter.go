package main

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"math"
	"time"
)

type HashWithSeed struct {
	Seed []byte
}

type bloomFilter struct {
	bitset    []byte
	hashFuncs []HashWithSeed
	m         uint
	k         uint64
}

var masks = [8]byte{
	1 << 0, 1 << 1, 1 << 2, 1 << 3, 1 << 4, 1 << 5, 1 << 6, 1 << 7,
}

func (bf bloomFilter) IsFound(data []byte) bool {
	for _, hashFunc := range bf.hashFuncs {
		i := (hashFunc.hash(data)) % bf.k
		target := bf.bitset[i/8]
		if target&masks[i%8] == 0 {
			return false
		}
	}
	return true
}

func (bf bloomFilter) Set(data []byte) {
	for _, hashFunc := range bf.hashFuncs {
		i := (hashFunc.hash(data)) % bf.k
		target := bf.bitset[i/8]
		bf.bitset[i/8] = target | masks[i%8]
	}
}

func NewBloomFilter(expectedElements uint, falsePositiveRate float64) (*bloomFilter, error) {
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		return nil, errors.New("falsePositiveRate must be in range (0, 1)")
	}
	m := calculateM(expectedElements, falsePositiveRate)
	k := calculateK(expectedElements, m)
	bitset := make([]byte, k/8)
	return &bloomFilter{
		bitset,
		createHashFunctions(k),
		m,
		k,
	}, nil
}

func DumpBloomFilter(bf bloomFilter, filename string) error {
	return errors.New("not implemented")
}

func LoadBloomFilter(filename string) (bloomFilter, error) {
	return bloomFilter{}, errors.New("not implemented")
}

func calculateM(expectedElements uint, falsePositiveRate float64) uint {
	return uint(math.Ceil(float64(expectedElements) * math.Abs(math.Log(falsePositiveRate)) / math.Pow(math.Log(2), float64(2))))
}

func calculateK(expectedElements uint, m uint) uint64 {
	return uint64(math.Ceil((float64(m) / float64(expectedElements)) * math.Log(2)))
}

func (h HashWithSeed) hash(data []byte) uint64 {
	fn := md5.New()
	fn.Write(append(data, h.Seed...))
	return binary.BigEndian.Uint64(fn.Sum(nil))
}

func createHashFunctions(k uint64) []HashWithSeed {
	h := make([]HashWithSeed, k)
	ts := uint64(time.Now().Unix())
	for i := range k {
		seed := make([]byte, 4)
		binary.BigEndian.PutUint64(seed, ts+i)
		hfn := HashWithSeed{Seed: seed}
		h[i] = hfn
	}
	return h
}

// TODO: Implement testing as well
