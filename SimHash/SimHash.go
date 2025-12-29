package SimHash

import (
	"crypto/md5"
	"fmt"
	"math/bits"
	"strings"
)

func splitString(s string) map[string]int {
	sa := strings.Split(strings.ToLower(s), " ")
	stringMap := make(map[string]int)
	for _, i := range sa {
		fmt.Println(i)
		_, ok := stringMap[i]
		if ok {
			stringMap[i] += 1
		} else {
			stringMap[i] = 1
		}

	}
	return stringMap
}
func md5Hash(data map[string]int) ([][16]byte, []int) {
	hashes := make([][16]byte, len(data))
	multiply := make([]int, len(data))
	i := 0
	for key, value := range data {
		hashes[i] = md5.Sum([]byte(key))
		multiply[i] = value
		i++
	}

	return hashes, multiply
}

func hammingDistance(hashes [][16]byte, multiply []int) ([16]byte, int) {
	var result [16]byte

	for i := 0; i < 16; i++ {
		for j := 0; j < 8; j++ {
			sum := 0
			mask := byte(1 << (7 - j))

			for k := 0; k < len(hashes); k++ {
				if hashes[k][i]&mask != 0 {
					sum += 1 * multiply[k]
				} else {
					sum--
				}
			}

			if sum > 0 {
				result[i] |= mask
			}
		}
	}

	count := 0
	for _, b := range result {
		count += bits.OnesCount8(b)
	}
	return result, count
}

func GetSimHashValue(text string) int {
	words := splitString(text)
	hashes, multiply := md5Hash(words)
	_, value := hammingDistance(hashes, multiply)
	return value
}
