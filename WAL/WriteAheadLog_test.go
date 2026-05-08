package wal

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	BlockManager "github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
)

// Pravi pomoćni Memtable
func getTestMemtableManager() (*memtable.MemtableManager, error) {
	conf := memtable.MemtableConfig{
		Type:           "hashmap",
		MaxSizeEntries: 100,
	}
	return memtable.NewMemtableManager(2, conf, nil, nil)
}

// Priprema testa: briše stare i pravi nove čiste foldere
func setupTest(t *testing.T) {
	cleanupTest()
	err := os.MkdirAll(FILE_PATH, 0755)
	if err != nil {
		t.Fatalf("Nije moguće kreirati test folder: %v", err)
	}
}

// Briše ceo WAL folder sa diska
func cleanupTest() {
	os.RemoveAll("./WAL")
}

// Provera da li WAL može da se napravi, ugasi, pa ponovo otvori
func TestCreateAndReopen(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	w, _ := CreatNewWAL(16, 4, FILE_PATH, 10)
	bm, err := BlockManager.NewBlockManager(2, 4)
	if err != nil {
		t.Fatalf("Nije moguće kreirati BlockManager: %v", err)
	}
	w.SetBlockManager(bm)
	w.AddRecord("test", []byte("data"))
	w.Close()

	w2, err := CreatNewWAL(16, 4, FILE_PATH, 10)
	w2.SetBlockManager(bm)
	if err != nil {
		t.Fatalf("Neuspešno ponovno otvaranje: %v", err)
	}
	if len(w2.segmentList) != 1 {
		t.Errorf("Očekivan 1 segment, dobijeno %d", len(w2.segmentList))
	}
	w2.Close()
}

// Provera da li sistem uspešno vraća podatke iz fajla u memoriju
func TestFullRecoveryCycle(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	w, _ := CreatNewWAL(16, 4, FILE_PATH, 10)
	bm, err := BlockManager.NewBlockManager(2, 4)
	if err != nil {
		t.Fatalf("Nije moguće kreirati BlockManager: %v", err)
	}
	w.SetBlockManager(bm)
	w.AddRecord("mali", []byte("v"))
	bigVal := []byte("vrednost_koja_se_fragmentise")
	w.AddRecord("veliki", bigVal)

	w.DeleteRecord("obrisan")
	w.Close()

	w2, _ := CreatNewWAL(16, 4, FILE_PATH, 10)
	w2.SetBlockManager(bm)
	mm, _ := getTestMemtableManager()

	if err := w2.Recovery(mm, time.Time{}); err != nil {
		t.Fatalf("Recovery puko: %v", err)
	}

	val, found, _ := mm.Get("mali")
	if !found || string(val) != "v" {
		t.Errorf("Mali zapis nije vraćen kako treba")
	}

	val2, found2, _ := mm.Get("veliki")
	if !found2 || !bytes.Equal(val2, bigVal) {
		t.Errorf("Veliki (fragmentisani) zapis nije ispravno spojen")
	}

	w2.Close()
}

// Provera šta se dešava kada header zapisa udari u samu ivicu bloka
func TestHeaderBoundaryEdgeCase(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	w, _ := CreatNewWAL(16, 4, FILE_PATH, 10)
	bm, err := BlockManager.NewBlockManager(2, 4)
	if err != nil {
		t.Fatalf("Nije moguće kreirati BlockManager: %v", err)
	}
	w.SetBlockManager(bm)
	w.AddRecord("k1", []byte("v1"))
	bigKey := "kljuc_posle_skoka"
	w.AddRecord(bigKey, []byte("podatak"))
	w.Close()

	w2, _ := CreatNewWAL(16, 4, FILE_PATH, 10)
	w2.SetBlockManager(bm)
	mm, _ := getTestMemtableManager()
	if err := w2.Recovery(mm, time.Time{}); err != nil {
		t.Fatalf("Recovery greška na granici bloka: %v", err)
	}

	_, found, _ := mm.Get(bigKey)
	if !found {
		t.Error("Zapis nije pronađen")
	}
	w2.Close()
}

// Provera da li WAL automatski pravi nove fajlove i briše stare
func TestRotationAndFlush(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	w, _ := CreatNewWAL(16, 4, FILE_PATH, 10)
	bm, err := BlockManager.NewBlockManager(2, 4)
	if err != nil {
		t.Fatalf("Nije moguće kreirati BlockManager: %v", err)
	}
	w.SetBlockManager(bm)
	for i := 0; i < 1600; i++ {
		w.AddRecord(fmt.Sprintf("key%d", i), []byte("duzi_podatak_za_rotaciju"))
	}

	if len(w.segmentList) < 2 {
		t.Errorf("Rotacija nije odradila posao")
	}

	if len(w.segmentList) >= 2 {
		w.lowWatermarks = []string{w.segmentList[1]}
		w.FlushWAL()
	}
	w.Close()
}

// Provera da li recovery preživljava ako su neki podaci u fajlu pokvareni
func TestCorruptedChunk(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	w, _ := CreatNewWAL(16, 4, FILE_PATH, 10)
	bm, err := BlockManager.NewBlockManager(2, 4)
	if err != nil {
		t.Fatalf("Nije moguće kreirati BlockManager: %v", err)
	}
	w.SetBlockManager(bm)
	w.AddRecord("validan", []byte("podatak"))
	w.Close()

	//Namerno kvarimo fajl dodavanjem smeća na kraj
	f, _ := os.OpenFile(w.segmentList[0], os.O_WRONLY|os.O_APPEND, 0644)
	f.Write([]byte{99, 0, 0, 0})
	f.Close()

	mm, _ := getTestMemtableManager()
	w2, _ := CreatNewWAL(16, 4, FILE_PATH, 10)
	w2.SetBlockManager(bm)

	w2.Recovery(mm, time.Time{})

	val, found, _ := mm.Get("validan")
	if !found || string(val) != "podatak" {
		t.Error("Dobar podatak pre kvara je morao biti sačuvan")
	}
	w2.Close()
}
