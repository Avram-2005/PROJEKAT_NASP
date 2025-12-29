package projekatnasp

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"
	"os"
)

type hyperLogLog struct {
	m       uint64  //hyperLogLog set size
	p       uint8   //hyperLogLog precision
	buckets []uint8 //hyperLogLog set
}

// HyperLogLog constructor-p(precision) must be between 4 and 16
func NewHyperLogLog(p uint8) (*hyperLogLog, error) {
	//preciznost je ocekivana da bude unutar odredjenog opsega, ako nije, baca error
	if p < HLL_MIN_PRECISION || p > HLL_MAX_PRECISION {
		return nil, fmt.Errorf("hyperLogLog precision should be between 4 and 16")
	}
	//bit shiftujemo da bi smo dobili m = 2 na stepen p
	m := uint64(1 << p)
	buckets := make([]uint8, m)
	return &hyperLogLog{
		m:       m,
		p:       p,
		buckets: buckets,
	}, nil
}

// Funkcija za dodavanje string-a u hyperloglog
func (hll *hyperLogLog) Add(text string) {
	hashValue := hash(text)
	//Trazimo prvih k-bitova za kljuc
	bucket := firstKbits(hashValue, uint64(hll.p))
	//Gledamo koliko nula ima na kraju, i dodajemo jos jedan na to
	value := trailingZeroBits(hashValue) + 1
	if hll.buckets[bucket] < uint8(value) {
		hll.buckets[bucket] = uint8(value)
	}
}

func (hll *hyperLogLog) Serialize(FileName string) error {
	bytes := make([]byte, 10)
	binary.BigEndian.PutUint16(bytes, uint16(hll.p))
	binary.BigEndian.PutUint64(bytes, hll.m)
	bytes = append(bytes, hll.buckets...)

	filePath := fmt.Sprintf(FileName, os.PathSeparator)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("Error with opening file")
	}
	defer file.Close()

	_, err = file.Write(bytes)
	if err != nil {
		return fmt.Errorf("Error with writing to file")
	}
	return nil
}

func Deserialize(FileName string) (*hyperLogLog, error) {
	filePath := fmt.Sprintf(FileName, os.PathSeparator)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("Error with opening file")
	}

	defer file.Close()
	file.Seek(0, 0)
	p := make([]byte, 2)
	_, err = file.Read(p)
	if err != nil {
		panic(err)
	}
	m := make([]byte, 8)
	_, err = file.Read(m)
	if err != nil {
		panic(err)
	}
	_, err = file.Read(m)
	bucketLen := binary.BigEndian.Uint64(m)
	buckets := make([]byte, bucketLen)
	_, err = file.Read(buckets)
	if err != nil {
		panic(err)
	}

	return &hyperLogLog{
		p:       uint8(binary.BigEndian.Uint16(p)),
		m:       bucketLen,
		buckets: buckets,
	}, nil
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

// vrlo prosta privatna hash funkcija
func hash(text string) uint64 {
	fn := md5.New()
	fn.Write([]byte(text))
	return binary.BigEndian.Uint64(fn.Sum(nil))
}
