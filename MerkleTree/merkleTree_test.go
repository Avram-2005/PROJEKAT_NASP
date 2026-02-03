package MerkleTree

import (
	"reflect"
	"testing"
)

func TestDetectChange(t *testing.T) {
	// test 1 - izmenjen podatak b
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

	diffs := FindDifference(m1.root, m2.root)
	for _, d := range diffs {
		t.Log("Izmenjen podatak: ", string(d))
	}

	// test 2 - isti podaci

	data3 := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}

	data4 := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}

	m3, err3 := NewMerkleTree(data3)
	if err3 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err3)
	}

	m4, err4 := NewMerkleTree(data4)
	if err4 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err4)
	}

	if m3.Verify(m4.root.hash) {
		t.Log("Podaci su isti")
	}

	diffs2 := FindDifference(m3.root, m4.root)
	for _, d := range diffs2 {
		t.Fatal("Izmenjen podatak: ", string(d))
	}

	// test 3 - izmenjeni podaci b i d

	data5 := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
		[]byte("d"),
	}

	data6 := [][]byte{
		[]byte("a"),
		[]byte("d"),
		[]byte("c"),
		[]byte("e"),
	}

	m5, err5 := NewMerkleTree(data5)
	if err5 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err5)
	}

	m6, err6 := NewMerkleTree(data6)
	if err6 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err6)
	}

	if m5.Verify(m6.root.hash) {
		t.Fatal("Podaci su isti")
	}

	diffs3 := FindDifference(m5.root, m6.root)
	for _, d := range diffs3 {
		t.Log("Izmenjen podatak: ", string(d))
	}
}

func TestSerializeDeserialize(t *testing.T) {
	data1 := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
		[]byte("d"),
	}
	m1, err1 := NewMerkleTree(data1)
	if err1 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err1)
	}

	data := m1.Serialize()
	m2 := Deserialize(data)

	if !reflect.DeepEqual(m1, m2) {
		t.Errorf("Merkle stablo nije isti pre i posle serijalizacije")
	}

	if m1.Verify(m2.root.hash) {
		t.Log("Podaci su isti")
	}

	diffs := FindDifference(m1.root, m2.root)
	for _, d := range diffs {
		t.Fatal("Izmenjen podatak: ", string(d))
	}
}
