package MerkleTree

import (
	"bytes"
	"crypto/sha256" //hash funkcija
	"encoding/binary"
	"fmt"
)

var emptyHash = make([]byte, 32) // prazan hash

type MerkleNode struct {
	left   *MerkleNode // pokazivac na levo dete
	right  *MerkleNode // pokazivac na desno dete
	parent *MerkleNode // pokazivac na roditelja
	hash   []byte      // hash vrednost cvora
	data   []byte      // podaci
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
		nodes = append(nodes, &MerkleNode{hash: h[:], data: d})
	}

	for len(nodes) > 1 { // dok ne ostane jedan cvor(koren)
		var level []*MerkleNode // sledeci nivo

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

func (m *MerkleTree) Verify(expectedRoot []byte) bool {
	if m == nil || m.root == nil {
		return false
	}
	return bytes.Equal(m.root.hash, expectedRoot) // da li su hash-evi isti
}

func FindDifference(a, b *MerkleNode) [][]byte {
	if a == nil || b == nil {
		return nil
	}

	// hash-evi su isti
	if bytes.Equal(a.hash, b.hash) {
		return nil
	}

	// hash se razlikuje, a dosli smo do lista stabla, znaci dosli smo do izmene
	if a.left == nil && a.right == nil {
		return [][]byte{a.data}
	}

	// cuvace sve pronadjene razlike u listovima
	var diffs [][]byte

	// rekurzivno trazimo razlike
	diffs = append(diffs, FindDifference(a.left, b.left)...)
	diffs = append(diffs, FindDifference(a.right, b.right)...)

	return diffs
}

func (m *MerkleTree) Serialize() []byte {
	if m.root == nil {
		return nil
	}
	return serializeNode(m.root)
}

func serializeNode(n *MerkleNode) []byte {
	if n == nil {
		return nil
	}

	var result []byte

	// dodajemo flag da li je list(1-da, 0-ne)
	if n.left == nil && n.right == nil {
		result = append(result, byte(1)) // list
	} else {
		result = append(result, byte(0)) // unutrasnji cvor
	}

	// dodajemo hash (32 bajta za sha256)
	result = append(result, n.hash...)

	// ako je list, dodajemo duzinu podataka i podatke
	if n.left == nil && n.right == nil && n.data != nil {
		dataLen := make([]byte, 4)
		binary.BigEndian.PutUint32(dataLen, uint32(len(n.data)))
		result = append(result, dataLen...)
		result = append(result, n.data...)
	}

	// rekurzivno dodajemo decu
	result = append(result, serializeNode(n.left)...)
	result = append(result, serializeNode(n.right)...)

	return result
}

func Deserialize(data []byte) *MerkleTree {
	root, _ := deserializeNode(data, 0)
	return &MerkleTree{root: root}
}

func deserializeNode(data []byte, offset int) (*MerkleNode, int) {
	n := &MerkleNode{}

	// flag
	var isLeaf bool
	if data[offset] == 1 {
		isLeaf = true
	} else {
		isLeaf = false
	}
	offset++

	// hash
	n.hash = data[offset : offset+32]
	offset += 32

	// ako je list, procitaj podatke
	if isLeaf {
		dataLen := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		n.data = data[offset : offset+int(dataLen)]
		offset += int(dataLen)
	} else { // rekurzivna deserijalizacija levog i desnog deteta
		n.left, offset = deserializeNode(data, offset)
		n.right, offset = deserializeNode(data, offset)

		if n.left != nil {
			n.left.parent = n
		}
		if n.right != nil {
			n.right.parent = n
		}
	}

	return n, offset
}
