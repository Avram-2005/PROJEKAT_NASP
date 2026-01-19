package merkleTree

import (
	"crypto/sha256" //hash funkcija
	"fmt"
)

var emptyHash = make([]byte, 32) // prazan hash

type MerkleNode struct {
	left   *MerkleNode // pokazivac na levo dete
	right  *MerkleNode // pokazivac na desno dete
	parent *MerkleNode // pokazivac na roditelja
	hash   []byte      //hash vrednost cvora
}

type MerkleTree struct {
	root *MerkleNode
}

func NewMerkleTree(data [][]byte) (*MerkleTree, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("Empty data set")
	}

	var nodes []*MerkleNode

	// listovi
	for _, d := range data {
		h := sha256.Sum256(d)
		nodes = append(nodes, &MerkleNode{hash: h[:]})
	}

	for len(nodes) > 1 { // dok ne ostane jedan cvor(koren)
		var level []*MerkleNode // seldeci nivo

		for i := 0; i < len(nodes); i += 2 {
			left := nodes[i]
			var right *MerkleNode

			// ako postoji desno dete
			if i+1 < len(nodes) {
				right = nodes[i+1]
			} else {
				right = &MerkleNode{
					hash: emptyHash, // prazan hash
				}
			}

			// hash levog i desnog deteta
			combined := append(left.hash, right.hash...)
			h := sha256.Sum256(combined)

			parent := &MerkleNode{
				left:  left,
				right: right,
				hash:  h[:],
			}
			left.parent = parent
			right.parent = parent

			level = append(level, parent)
		}
		nodes = level // prelazak na sledeci nivo stabla
	}

	return &MerkleTree{
		root: nodes[0], //koren
	}, nil
}
