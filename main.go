package main

import (
	"fmt"

	h "github.com/Avram-2005/PROJEKAT_NASP/SimHash"
)

// Sve se jos testira
// DONT JUDGE
func main() {
	text := "Probabilistic data structures are fun, and fun is to learn them"
	value := h.GetSimHashValue(text)
	fmt.Println(value)

}
