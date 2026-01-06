package SimHash

import (
	"crypto/md5"
	"math/bits"
	"strings"
)

// Pretvara string u lowercase i deli u pod reci
// Vraca mapu   rec : broj pojavljivanja
func splitString(s string) map[string]int {
	parts := strings.Split(strings.ToLower(s), " ")
	wordCount := make(map[string]int)

	for _, p := range parts {
		if p == "" {
			continue
		}
		wordCount[p]++
	}
	return wordCount
}

// md5Hash racuna MD5 hash za svaku rec i cuva njen hesh u 16 byte-a
func md5Hash(data map[string]int) ([][16]byte, []int) {
	hashes := make([][16]byte, len(data))
	weights := make([]int, len(data))

	i := 0
	for key, value := range data {
		hashes[i] = md5.Sum([]byte(key))
		weights[i] = value
		i++
	}
	return hashes, weights
}

// Racunava 128-bitni SimHash koristeci ponderisano sabiranje bitova
func computeSimHash(hashes [][16]byte, weights []int) [16]byte {
	var result [16]byte

	for i := 0; i < 16; i++ {
		for j := 0; j < 8; j++ {
			sum := 0
			mask := byte(1 << (7 - j))

			for k := 0; k < len(hashes); k++ {
				if hashes[k][i]&mask != 0 {
					sum += weights[k]
				} else {
					sum -= weights[k]
				}
			}

			if sum > 0 {
				result[i] |= mask
			}
		}
	}
	return result
}

// GetSimHashValue vraca SimHash vrednost za dati string
func GetSimHashValue(text string) [16]byte {
	words := splitString(text)
	hashes, weights := md5Hash(words)
	return computeSimHash(hashes, weights)
}

// Racuna Hemingovu udaljenost izmedju dva stringa koristeci SimHash vrednosti
func HammingDistance(text1, text2 string) int {
	hash1 := GetSimHashValue(text1)
	hash2 := GetSimHashValue(text2)

	distance := 0
	for i := 0; i < 16; i++ {
		distance += bits.OnesCount8(hash1[i] ^ hash2[i])
	}
	return distance
}
