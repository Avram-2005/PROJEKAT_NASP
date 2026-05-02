package MerkleTree

import (
	"testing"
	"time"

	. "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func TestDetectChange(t *testing.T) {
	// test 1 - izmenjen podatak pod kljucem b
	ts := time.Now()

	r1, _ := NewRecord("a", []byte("value1"), false, ts)
	r2, _ := NewRecord("b", []byte("value2"), false, ts)
	r3, _ := NewRecord("c", []byte("value3"), false, ts)

	r1Modified, _ := NewRecord("a", []byte("value1"), false, ts)
	r2Modified, _ := NewRecord("b", []byte("a"), false, ts)
	r3Modified, _ := NewRecord("c", []byte("value3"), false, ts)

	records1 := []Record{*r1, *r2, *r3}
	records2 := []Record{*r1Modified, *r2Modified, *r3Modified}

	m1, err1 := NewMerkleTree(records1)
	if err1 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err1)
	}

	m2, err2 := NewMerkleTree(records2)
	if err2 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err2)
	}

	if m1.Verify(m2.RootHash()) {
		t.Fatal("Podaci su isti")
	}

	diffs := FindDifference(m1.Root(), m2.Root())
	for _, d := range diffs {
		t.Logf("Izmenjen podatak pod kljucem: %s, sa podatkom: %s, sa timestamp: %s, sa tombstone: %t", string(d.Key), string(d.Value), d.Timestamp.String(), d.Tombstone)
	}

	// test 2 - isti podaci

	r4, _ := NewRecord("a", []byte("value1"), false, ts)
	r5, _ := NewRecord("b", []byte("value2"), false, ts)
	r6, _ := NewRecord("c", []byte("value3"), false, ts)

	r4Modified, _ := NewRecord("a", []byte("value1"), false, ts)
	r5Modified, _ := NewRecord("b", []byte("value2"), false, ts)
	r6Modified, _ := NewRecord("c", []byte("value3"), false, ts)

	records3 := []Record{*r4, *r5, *r6}
	records4 := []Record{*r4Modified, *r5Modified, *r6Modified}

	m3, err3 := NewMerkleTree(records3)
	if err3 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err3)
	}

	m4, err4 := NewMerkleTree(records4)
	if err4 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err4)
	}

	if m3.Verify(m4.RootHash()) {
		t.Log("Podaci su isti")
	}

	diffs2 := FindDifference(m3.Root(), m4.Root())
	for _, d := range diffs2 {
		t.Fatalf("Izmenjen podatak pod kljucem: %s, sa podatkom: %s, sa timestamp: %s, sa tombstone: %t", string(d.Key), string(d.Value), d.Timestamp.String(), d.Tombstone)
	}

	// test 3 - izmenjeni podaci pod kljucevima a i c

	r7, _ := NewRecord("a", []byte("value1"), false, ts)
	r8, _ := NewRecord("b", []byte("value2"), false, ts)
	r9, _ := NewRecord("c", []byte("value3"), false, ts)

	r7Modified, _ := NewRecord("a", []byte("a"), false, ts)
	r8Modified, _ := NewRecord("b", []byte("value2"), false, ts)
	r9Modified, _ := NewRecord("c", []byte("value4"), false, ts)

	records5 := []Record{*r7, *r8, *r9}
	records6 := []Record{*r7Modified, *r8Modified, *r9Modified}

	m5, err5 := NewMerkleTree(records5)
	if err5 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err5)
	}

	m6, err6 := NewMerkleTree(records6)
	if err6 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err6)
	}

	if m5.Verify(m6.RootHash()) {
		t.Fatal("Podaci su isti")
	}

	diffs3 := FindDifference(m5.Root(), m6.Root())
	for _, d := range diffs3 {
		t.Logf("Izmenjen podatak pod kljucem: %s, sa podatkom: %s, sa timestamp: %s, sa tombstone: %t", string(d.Key), string(d.Value), d.Timestamp.String(), d.Tombstone)
	}

	// test 4 - izmenjen kljuc a i Timestamp pod kljucem c

	r10, _ := NewRecord("a", []byte("value1"), false, ts)
	r11, _ := NewRecord("b", []byte("value2"), false, ts)
	r12, _ := NewRecord("c", []byte("value3"), false, ts)

	r10Modified, _ := NewRecord("aa", []byte("value1"), false, ts)
	r11Modified, _ := NewRecord("b", []byte("value2"), false, ts)
	r12Modified, _ := NewRecord("c", []byte("value3"), false, time.Now().Add(time.Hour))

	records7 := []Record{*r10, *r11, *r12}
	records8 := []Record{*r10Modified, *r11Modified, *r12Modified}

	m7, err7 := NewMerkleTree(records7)
	if err7 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err7)
	}

	m8, err8 := NewMerkleTree(records8)
	if err8 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err8)
	}

	if m7.Verify(m8.RootHash()) {
		t.Fatal("Podaci su isti")
	}

	diffs4 := FindDifference(m7.Root(), m8.Root())
	for _, d := range diffs4 {
		t.Logf("Izmenjen podatak pod kljucem: %s, sa podatkom: %s, sa timestamp: %s, sa tombstone: %t", string(d.Key), string(d.Value), d.Timestamp.String(), d.Tombstone)
	}
}

func TestSerializeDeserialize(t *testing.T) {
	ts := time.Now()

	r1, _ := NewRecord("a", []byte("value1"), false, ts)
	r2, _ := NewRecord("b", []byte("value2"), false, ts)
	r3, _ := NewRecord("c", []byte("value3"), false, ts)

	records1 := []Record{*r1, *r2, *r3}

	m1, err1 := NewMerkleTree(records1)
	if err1 != nil {
		t.Fatal("Greska pri kreiranju merkle stabla ", err1)
	}

	data := m1.Serialize()
	m2 := Deserialize(data)

	if !m1.Verify(m2.RootHash()) {
		t.Fatal("Podaci su razliciti")
	}

	diffs := FindDifference(m1.Root(), m2.Root())
	for _, d := range diffs {
		t.Fatalf("Izmenjen podatak pod kljucem: %s, sa podatkom: %s, sa timestamp: %s, sa tombstone: %t", string(d.Key), string(d.Value), d.Timestamp.String(), d.Tombstone)
	}
}
