package BloomFilter

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"math"
	"time"
)

type hashWithSeed struct {
	Seed []byte
}

type BloomFilter struct {
	numHashFuncs uint8
	numBits      uint32
	hashFuncs    []hashWithSeed
	bitset       []byte
}

func (bf *BloomFilter) IsFound(data []byte) bool {
	for _, hashFunc := range bf.hashFuncs {
		i := hashFunc.hash(data) % bf.numBits
		target := bf.bitset[i/8]
		if target&(1<<(i%8)) == 0 {
			return false
		}
	}
	return true
}

func CalculateBloomFilterSize(expectedElements uint, falsePositiveRate float64) uint32 {
	numBits := calculateM(expectedElements, falsePositiveRate)
	numHashFuncs := calculateK(expectedElements, numBits)
	return sizeOfMetadata + uint32(sizeOfSeed)*uint32(numHashFuncs) + uint32(numBits/8+1)
}

func (bf *BloomFilter) GetSize() uint32 {
	return sizeOfMetadata + uint32(sizeOfSeed)*uint32(bf.numHashFuncs) + uint32(bf.numBits/8+1)
}

func (bf *BloomFilter) Set(data []byte) {
	for _, hashFunc := range bf.hashFuncs {
		i := hashFunc.hash(data) % bf.numBits
		target := bf.bitset[i/8]
		bf.bitset[i/8] = target | (1 << (i % 8))
	}
}

func NewBloomFilter(expectedElements uint, falsePositiveRate float64) (*BloomFilter, error) {
	if falsePositiveRate <= 0 || falsePositiveRate >= 1 {
		return nil, errors.New("falsePositiveRate must be in range (0, 1)")
	}
	numBits := calculateM(expectedElements, falsePositiveRate)
	numHashFuncs := calculateK(expectedElements, numBits)
	bitset := make([]byte, numBits/8+1)
	return &BloomFilter{
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

func (bf *BloomFilter) Dump() []byte {
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

func LoadBloomFilter(data []byte) *BloomFilter {
	numHashFuncs := binary.BigEndian.Uint16(data)
	numBits := binary.BigEndian.Uint32(data[sizeOfNumHashFuncs:])
	hashFuncs := make([]hashWithSeed, numHashFuncs)
	for i := range numHashFuncs {
		seed := make([]byte, 4)
		copy(seed, data[sizeOfMetadata+i*4:])
		hashFuncs[i] = hashWithSeed{seed}
	}
	bitset := make([]byte, numBits/8+1)
	sizeOfHashPart := sizeOfSeed * numHashFuncs
	copy(bitset, data[sizeOfMetadata+sizeOfHashPart:])
	return &BloomFilter{
		uint8(numHashFuncs),
		numBits,
		hashFuncs,
		bitset,
	}
}

func (h hashWithSeed) hash(data []byte) uint32 {
	fn := md5.New()
	fn.Write(append(data, h.Seed...))
	return binary.BigEndian.Uint32(fn.Sum(nil))
}

func createHashFunctions(k uint8) []hashWithSeed {
	h := make([]hashWithSeed, k)
	ts := uint32(time.Now().Unix())
	for i := range uint32(k) {
		seed := make([]byte, 4)
		binary.BigEndian.PutUint32(seed, ts+i)
		hfn := hashWithSeed{Seed: seed}
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
