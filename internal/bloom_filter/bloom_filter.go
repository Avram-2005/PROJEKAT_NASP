package bloom_filter

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
	numHashFuncs uint8
	numBits      uint32
	hashFuncs    []HashWithSeed
	bitset       []byte
}

var masks = [8]byte{
	1 << 0, 1 << 1, 1 << 2, 1 << 3, 1 << 4, 1 << 5, 1 << 6, 1 << 7,
}

func (bf *bloomFilter) IsFound(data []byte) bool {
	for _, hashFunc := range bf.hashFuncs {
		i := hashFunc.hash(data) % bf.numBits
		target := bf.bitset[i/8]
		if target&masks[i%8] == 0 {
			return false
		}
	}
	return true
}

func (bf *bloomFilter) Set(data []byte) {
	for _, hashFunc := range bf.hashFuncs {
		i := hashFunc.hash(data) % bf.numBits
		target := bf.bitset[i/8]
		bf.bitset[i/8] = target | masks[i%8]
	}
}

func NewBloomFilter(expectedElements uint, falsePositiveRate float64) (*bloomFilter, error) {
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		return nil, errors.New("falsePositiveRate must be in range (0, 1)")
	}
	numBits := calculateM(expectedElements, falsePositiveRate)
	numHashFuncs := calculateK(expectedElements, numBits)
	bitset := make([]byte, numBits/8+1)
	return &bloomFilter{
		numHashFuncs,
		numBits,
		createHashFunctions(numHashFuncs),
		bitset,
	}, nil
}

const sizeOfNumHashFuncs = 2
const sizeOfNumBits = 4
const sizeOfMetadata = sizeOfNumHashFuncs + sizeOfNumBits
const sizeOfSeed = 4

func (bf *bloomFilter) Dump() []byte {
	sizeOfHashPart := sizeOfSeed * uint(bf.numHashFuncs)
	sizeOfBitsPart := uint(bf.numBits/8 + 1)
	data := make([]byte, sizeOfMetadata+sizeOfHashPart+sizeOfBitsPart)
	binary.BigEndian.PutUint16(data, uint16(bf.numHashFuncs))
	binary.BigEndian.PutUint32(data[sizeOfNumHashFuncs:], bf.numBits)
	for i, seed := range bf.hashFuncs {
		copy(data[sizeOfMetadata+i*sizeOfSeed:], seed.Seed)
	}
	copy(data[sizeOfMetadata+sizeOfHashPart:], bf.bitset)
	return data
}

func LoadBloomFilter(data []byte) *bloomFilter {
	numHashFuncs := binary.BigEndian.Uint16(data)
	numBits := binary.BigEndian.Uint32(data[sizeOfNumHashFuncs:])
	hashFuncs := make([]HashWithSeed, numHashFuncs)
	for i := range numHashFuncs {
		seed := make([]byte, 4)
		copy(seed, data[sizeOfMetadata+i*4:])
		hashFuncs[i] = HashWithSeed{seed}
	}
	bitset := make([]byte, numBits/8+1)
	sizeOfHashPart := sizeOfSeed * numHashFuncs
	copy(bitset, data[sizeOfMetadata+sizeOfHashPart:])
	return &bloomFilter{
		uint8(numHashFuncs),
		numBits,
		hashFuncs,
		bitset,
	}
}

func (h HashWithSeed) hash(data []byte) uint32 {
	fn := md5.New()
	fn.Write(append(data, h.Seed...))
	return binary.BigEndian.Uint32(fn.Sum(nil))
}

func createHashFunctions(k uint8) []HashWithSeed {
	h := make([]HashWithSeed, k)
	ts := uint32(time.Now().Unix())
	for i := range uint32(k) {
		seed := make([]byte, 4)
		binary.BigEndian.PutUint32(seed, ts+i)
		hfn := HashWithSeed{Seed: seed}
		h[i] = hfn
	}
	return h
}

func calculateM(expectedElements uint, falsePositiveRate float64) uint32 {
	return uint32(math.Ceil(float64(expectedElements) * math.Abs(math.Log(falsePositiveRate)) / math.Pow(math.Log(2), float64(2))))
}

func calculateK(expectedElements uint, m uint32) uint8 {
	return uint8(math.Ceil((float64(m) / float64(expectedElements)) * math.Log(2)))
}
