package projekatnasp

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

type countMinSketch struct {
	k      uint           // broj hash funkcija(redovi)
	m      uint           // broj kolona
	table  [][]int        // k x m matrica
	hashes []hashWithSeed // hash funkcije
}

// kreiranje nove instance
func NewCountMinSketch(epsilon, delta float64) (*countMinSketch, error) {
	if epsilon <= 0 || epsilon > 1 {
		return nil, fmt.Errorf("Epsilon must be lesser than 1 and greater than 0")
	}
	if delta <= 0 || delta > 1 {
		return nil, fmt.Errorf("Delta must be lesser than 1 and greater than 0")
	}

	m := calculateM(epsilon)
	k := calculateK(delta)

	table := make([][]int, k)
	for i := range table {
		table[i] = make([]int, m)
	}

	return &countMinSketch{
		k:      k,
		m:      m,
		table:  table,
		hashes: createHashFunctions(k),
	}, nil
}

func (cms *countMinSketch) Add(key string) {
	// pretvaranje kljuca u niz bajtova
	data := []byte(key)

	for i := uint(0); i < cms.k; i++ {
		j := cms.hashes[i].hash(data) % uint64(cms.m)
		cms.table[i][j]++
	}
}

func (cms *countMinSketch) Estimate(key string) int {
	data := []byte(key)
	r := make([]int, cms.k) // niz duzine k

	for i := uint(0); i < cms.k; i++ {
		j := cms.hashes[i].hash(data) % uint64(cms.m)
		r[i] = cms.table[i][j]
	}

	// trazi minimum
	min := r[0]
	for i := 1; i < len(r); i++ {
		if r[i] < min {
			min = r[i]
		}
	}

	return min
}

func (cms *countMinSketch) Clear() {
	for i := uint(0); i < cms.k; i++ {
		for j := uint(0); j < cms.m; j++ {
			cms.table[i][j] = 0
		}
	}
}

// helper functions

func calculateM(epsilon float64) uint {
	return uint(math.Ceil(math.E / epsilon))
}

func calculateK(delta float64) uint {
	return uint(math.Ceil(math.Log(math.E / delta)))
}

type hashWithSeed struct {
	Seed []byte
}

func (h hashWithSeed) hash(data []byte) uint64 {
	fn := md5.New()
	fn.Write(append(data, h.Seed...))
	return binary.BigEndian.Uint64(fn.Sum(nil))
}

func createHashFunctions(k uint) []hashWithSeed {
	h := make([]hashWithSeed, k)
	ts := uint(time.Now().Unix())
	for i := uint(0); i < k; i++ {
		seed := make([]byte, 32)
		binary.BigEndian.PutUint32(seed, uint32(ts+i))
		hfn := hashWithSeed{Seed: seed}
		h[i] = hfn
	}
	return h
}
