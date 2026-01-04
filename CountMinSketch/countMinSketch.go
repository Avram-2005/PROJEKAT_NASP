package CountMinSketch

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

func (cms *countMinSketch) Serialize() []byte {
	kbytes := make([]byte, 4)
	binary.BigEndian.PutUint32(kbytes, uint32(cms.k))

	mbytes := make([]byte, 4)
	binary.BigEndian.PutUint32(mbytes, uint32(cms.m))

	bytes := append(kbytes, mbytes...)

	tablebytes := make([]byte, 4)
	for i := uint(0); i < cms.k; i++ {
		for j := uint(0); j < cms.m; j++ {
			binary.BigEndian.PutUint32(tablebytes, uint32(cms.table[i][j]))
			bytes = append(bytes, tablebytes...)
		}
	}

	hashbytes := make([]byte, 4)
	for _, seed := range cms.hashes {
		binary.BigEndian.PutUint32(hashbytes, uint32(len(seed.Seed)))
		bytes = append(bytes, hashbytes...)
		bytes = append(bytes, seed.Seed...)
	}

	return bytes
}

func Deserialize(data []byte) *countMinSketch {
	bytes := 0

	k := uint(binary.BigEndian.Uint32(data[bytes:]))
	bytes += 4

	m := uint(binary.BigEndian.Uint32(data[bytes:]))
	bytes += 4

	table := make([][]int, k)
	for i := uint(0); i < k; i++ {
		table[i] = make([]int, m)
		for j := uint(0); j < m; j++ {
			table[i][j] = int(binary.BigEndian.Uint32(data[bytes:]))
			bytes += 4
		}
	}

	hashes := make([]hashWithSeed, k)
	for i := uint(0); i < k; i++ {
		seedLen := binary.BigEndian.Uint32(data[bytes:])
		bytes += 4

		seed := make([]byte, seedLen)
		copy(seed, data[bytes:bytes+int(seedLen)])
		bytes += int(seedLen)
		hashes[i] = hashWithSeed{Seed: seed}
	}

	return &countMinSketch{
		k:      k,
		m:      m,
		table:  table,
		hashes: hashes,
	}
}

// pomocne funkcije

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
