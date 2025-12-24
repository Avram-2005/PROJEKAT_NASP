package main

import (
	"crypto/md5"
	"fmt"
	"strings"
)

func splitString(s string) []string {
	sa := strings.Split(strings.ToLower(s), " ")
	for _, i := range sa {
		fmt.Println(i)
	}
	return sa
}
func MD5Hash(data []string) [][16]byte {
	hashes := make([][16]byte, len(data))

	for i, s := range data {
		hashes[i] = md5.Sum([]byte(s))
	}

	return hashes
}

func HammingDistance(hashes [][16]byte) [16]byte {
	var result [16]byte

	for i := 0; i < 16; i++ {
		for j := 0; j < 8; j++ {
			sum := 0
			mask := byte(1 << (7 - j))

			for _, h := range hashes {
				if h[i]&mask != 0 {
					sum++
				} else {
					sum--
				}
			}

			if sum > 0 {
				result[i] |= mask
			}
		}
	}

	return result
}

// Sve se jos testira
// DONT JUDGE
func main() {
	text := "Ovo je neki test tekst"

	words := splitString(text)
	hashes := MD5Hash(words)
	center := HammingDistance(hashes)
	for i := 0; i < len(hashes); i++ {
		fmt.Println(hashes[i])
	}
	fmt.Println(center)
}
