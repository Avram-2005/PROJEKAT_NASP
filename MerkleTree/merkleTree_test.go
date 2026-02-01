package MerkleTree

import (
	"testing"
)

func TestDetectChange(t *testing.T) {
	data1 := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}

	data2 := [][]byte{
		[]byte("a"),
		[]byte("d"),
		[]byte("c"),
	}

	m1, err1 := NewMerkleTree(data1)
	if err1 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err1)
	}

	m2, err2 := NewMerkleTree(data2)
	if err2 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err2)
	}

	if m1.Verify(m2.root.hash) {
		t.Fatal("Podaci su isti")
	}

	diff := FindDifference(m1.root, m2.root)
	if diff != nil {
		t.Log("Pronadjena izmena u listu sa hash-om", diff.hash)
	}

}
