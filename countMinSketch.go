package projekatnasp

import (
	"crypto/md5"
	"encoding/binary"
	"math"
	"time"
)

type CountMinSketch struct {
	k     uint    // broj hash funkcija(redovi)
	m     uint    // broj kolona
	table [][]int // k x m matrica
}

// kreiranje nove instance
func NewCountMinSketch(epsilon, delta float64) *CountMinSketch {
	m := CalculateM(epsilon)
	k := CalculateK(delta)

	table := make([][]int, k)
	for i := range table {
		table[i] = make([]int, m)
	}

	return &CountMinSketch{
		k:     k,
		m:     m,
		table: table,
	}
}

func (cms *CountMinSketch) Add(key string) {
	// pretvaranje kljuca u niz bajtova
	data := []byte(key)

	// kreiranje k hash funkcija
	hashes := CreateHashFunctions(cms.k)

	for i := uint(0); i < cms.k; i++ {
		j := hashes[i].Hash(data) % uint64(cms.m)
		cms.table[i][j]++
	}
}

// helper functions

func CalculateM(epsilon float64) uint {
	return uint(math.Ceil(math.E / epsilon))
}

func CalculateK(delta float64) uint {
	return uint(math.Ceil(math.Log(math.E / delta)))
}

type HashWithSeed struct {
	Seed []byte
}

func (h HashWithSeed) Hash(data []byte) uint64 {
	fn := md5.New()
	fn.Write(append(data, h.Seed...))
	return binary.BigEndian.Uint64(fn.Sum(nil))
}

func CreateHashFunctions(k uint) []HashWithSeed {
	h := make([]HashWithSeed, k)
	ts := uint(time.Now().Unix())
	for i := uint(0); i < k; i++ {
		seed := make([]byte, 32)
		binary.BigEndian.PutUint32(seed, uint32(ts+i))
		hfn := HashWithSeed{Seed: seed}
		h[i] = hfn
	}
	return h
}
