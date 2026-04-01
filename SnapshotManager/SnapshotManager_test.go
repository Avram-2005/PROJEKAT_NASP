package SnapshotManager

import (
	"encoding/binary"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	snapshot "github.com/Avram-2005/PROJEKAT_NASP/Snapshot"
)

func TestSnapshotManager(t *testing.T) {
	sp, err := NewSnapshotManager()
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom incijalizacije SnapshotManager-a")
		t.FailNow()
	}
	//Dodavanje svih vrednosti za testiranje
	filepath := "test.bin"
	file, err := os.Create(filepath)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom stvaranja fajla")
		t.FailNow()
	}
	bm, err := BlockManager.NewBlockManager(4, 4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("treba da se prijavi greska, ali nije prijavljena")
		t.FailNow()
	}
	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)
	bm.PutSpecific(file, 0, 0, 100, &data1)
	bm.PutSpecific(file, 0, 100, 100, &data1)
	bm.PutSpecific(file, 0, 200, 100, &data1)
	bm.PutSpecific(file, 0, 300, 100, &data1)

	timestamp1 := time.Now()
	snapshot1, err := snapshot.NewSnapshot(filepath, 0, 0, 100, timestamp1, bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("pogresna 1 inicijalizacija")
		t.FailNow()
	}

	timestamp2 := time.Now()
	snapshot2, err := snapshot.NewSnapshot(filepath, 0, 100, 100, timestamp2, bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("pogresna 2 inicijalizacija")
		t.FailNow()
	}
	timestamp3 := time.Now()
	snapshot3, err := snapshot.NewSnapshot(filepath, 0, 200, 100, timestamp3, bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("pogresna 3 inicijalizacija")
		t.FailNow()
	}

	timestamp4 := time.Now()
	snapshot4, err := snapshot.NewSnapshot(filepath, 0, 200, 100, timestamp4, bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("pogresna 4 inicijalizacija")
		t.FailNow()
	}

	sp.Add("key1", snapshot1)
	sp.Add("key1", snapshot2)
	sp.Add("key1", snapshot3)
	sp.Add("key2", snapshot4)

	// Provera broja verzija posle dodavanja
	number, err := sp.GetVersionCount("key1")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja broja verzija prvog kljuca")
		t.FailNow()
	}
	if number != 3 {
		fmt.Print(err)
		t.Errorf("pogresan broj verzija prvog kljuca")
		t.FailNow()
	}

	// Provera broja verzija posle dodavanja
	number, err = sp.GetVersionCount("key2")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja broja verzija drugog kljuca")
		t.FailNow()
	}
	if number != 1 {
		fmt.Print(err)
		t.Errorf("pogresan broj verzija drugog kljuca")
		t.FailNow()
	}

	//proveravanje metode getlatest
	compare1, err := sp.GetLatest("key1")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja poslednje verzije prvog kljuca")
		t.FailNow()
	}
	if !reflect.DeepEqual(snapshot3, compare1) {
		fmt.Print(snapshot3, compare1)
		t.Errorf("pogresna poslednja verzija prvog kljuca")
		t.FailNow()
	}

	//Proveravanje metode getfirst
	compare2, err := sp.GetFirst("key1")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja prve verzije prvog kljuca")
		t.FailNow()
	}
	if !reflect.DeepEqual(snapshot1, compare2) {
		fmt.Print(snapshot1, compare2)
		t.Errorf("pogresna poslednja verzija prvog kljuca")
		t.FailNow()
	}

	compare3, err := sp.Get("key1", 1)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja druge verzije prvog kljuca")
		t.FailNow()
	}
	if !reflect.DeepEqual(snapshot2, compare3) {
		fmt.Print(snapshot2, compare3)
		t.Errorf("pogresna druga verzija prvog kljuca")
		t.FailNow()
	}

	compare4, err := sp.Get("key2", 0)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja prve verzije verzije prvog kljuca")
		t.FailNow()
	}
	if !reflect.DeepEqual(snapshot4, compare4) {
		fmt.Print(snapshot4, compare4)
		t.Errorf("pogresna prva verzija drugog kljuca")
		t.FailNow()
	}

	compare4, err = sp.GetFirst("key2")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja prve verzije verzije drugog kljuca")
		t.FailNow()
	}
	if !reflect.DeepEqual(snapshot4, compare4) {
		fmt.Print(snapshot4, compare4)
		t.Errorf("pogresna prva verzija drugog kljuca")
		t.FailNow()
	}

	compare4, err = sp.GetLatest("key2")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom trazenja poslednje verzije verzije drugog kljuca")
		t.FailNow()
	}
	if !reflect.DeepEqual(snapshot4, compare4) {
		fmt.Print(snapshot4, compare4)
		t.Errorf("pogresna poslednja verzija drugog kljuca")
		t.FailNow()
	}

	err = sp.Free("key1")
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom oslobadjanja kljuca")
		t.FailNow()
	}
	// posle oslobadjanja memorije za key1, ocekujemo error ako pokusamo da ga nadjemo
	_, err = sp.GetVersionCount("key1")
	if err == nil {
		fmt.Print(err)
		t.Errorf("treba da se prijavi greska, ali nije prijavljena")
		t.FailNow()
	}

	file.Close()
	err = os.Remove(filepath)
	if err != nil {
		fmt.Print(err)
		t.Errorf("zatvaranje fajla onemoguceno")
		t.FailNow()
	}

}

// testiramo sposobnost snapshotmanagera da dobavi vrednosti na koje snapshot-ovi pokazuju
func TestConnectionWithSnapshot(t *testing.T) {
	sp, err := NewSnapshotManager()
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom incijalizacije SnapshotManager-a")
		t.FailNow()
	}
	//Dodavanje svih vrednosti za testiranje
	filepath := "test.bin"
	file, err := os.Create(filepath)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom stvaranja fajla")
		t.FailNow()
	}
	bm, err := BlockManager.NewBlockManager(4, 4)
	if err != nil {
		fmt.Print(err)
		t.Errorf("treba da se prijavi greska, ali nije prijavljena")
		t.FailNow()
	}
	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 100)
	binary.BigEndian.PutUint64(data2, 56)
	data3 := make([]byte, 100)
	binary.BigEndian.PutUint64(data3, 90)

	bm.PutSpecific(file, 0, 0, 100, &data1)
	bm.PutSpecific(file, 0, 100, 100, &data2)
	bm.PutSpecific(file, 0, 200, 100, &data3)

	timestamp1 := time.Now()
	snapshot1, err := snapshot.NewSnapshot(filepath, 0, 0, 100, timestamp1, bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("pogresna 1 inicijalizacija")
		t.FailNow()
	}

	timestamp2 := time.Now()
	snapshot2, err := snapshot.NewSnapshot(filepath, 0, 100, 100, timestamp2, bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("pogresna 2 inicijalizacija")
		t.FailNow()
	}
	timestamp3 := time.Now()
	snapshot3, err := snapshot.NewSnapshot(filepath, 0, 200, 100, timestamp3, bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("pogresna 3 inicijalizacija")
		t.FailNow()
	}

	sp.Add("key1", snapshot1)
	sp.Add("key1", snapshot2)
	sp.Add("key2", snapshot3)

	//testiramo dobavljanje vrednosti po verziji
	readValue, err := sp.GetValue("key1", 0, bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom dobavljanja podatka")
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, (*readValue)) {
		fmt.Print(data1, (*readValue))
		t.Errorf("procitani pogresni podaci")
		t.FailNow()
	}
	//testiramo dobavljanje vrednosti po timestamp-u
	readValue, err = sp.GetValueByTimestamp("key1", timestamp2, bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom dobavljanja podataka")
		t.FailNow()
	}
	if !reflect.DeepEqual(data2, (*readValue)) {
		fmt.Print(data2, (*readValue))
		t.Errorf("procitani pogresni podaci")
		t.FailNow()
	}
	//testiramo getlatest i getfirst
	readValue, err = sp.GetValueLatest("key2", bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom dobavljanja podatka")
		t.FailNow()
	}
	if !reflect.DeepEqual(data3, (*readValue)) {
		fmt.Print(data3, (*readValue))
		t.Errorf("procitani pogresni podaci")
		t.FailNow()
	}
	readValue, err = sp.GetValueFirst("key2", bm)
	if err != nil {
		fmt.Print(err)
		t.Errorf("greska tokom dobavljanja podatka")
		t.FailNow()
	}
	if !reflect.DeepEqual(data3, (*readValue)) {
		fmt.Print(data3, (*readValue))
		t.Errorf("procitani pogresni podaci")
		t.FailNow()
	}
	file.Close()
	err = os.Remove(filepath)
	if err != nil {
		fmt.Print(err)
		t.Errorf("zatvaranje fajla onemoguceno")
		t.FailNow()
	}

}
