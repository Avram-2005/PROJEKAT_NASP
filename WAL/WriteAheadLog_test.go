package wal

import (
	"os"
	"testing"
)

func setupTest(t *testing.T) string {
	err := os.MkdirAll(FILE_PATH, 0755)
	if err != nil {
		t.Fatalf("Nije moguće kreirati test folder: %v", err)
	}
	return FILE_PATH
}

func cleanupTest() {
	os.RemoveAll("./WAL")
}

func TestCreateNewWAL(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	segmentSize := 1024
	blockSize := 16

	walObject, err := CreatNewWAL(segmentSize, blockSize)
	if err != nil {
		t.Fatalf("Greška pri kreiranju WAL-a: %v", err)
	}

	if walObject == nil {
		t.Fatal("WAL je nil, došlo je do greške pri kreiranju")
	} else {
		defer walObject.Close()
	}
}

func TestWriteAndRead(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	w, err := CreatNewWAL(64, 16)
	if err != nil {
		t.Fatalf("Kreiranje nije uspelo: %v", err)
	}
	defer w.Close()

	key := "kljuc1"
	value := []byte("neka vrednost")

	err2 := w.AddRecord(key, value)
	if err2 != nil {
		t.Fatalf("Greška tokom upisa: %v", err2)
	}

	w.ReadAll()
}

func TestSegmentSplitting(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	sSize := 4048
	bSize := 4
	myWal, err := CreatNewWAL(sSize, bSize)
	if err != nil {
		t.Fatalf("Greška: %v", err)
	}
	defer myWal.Close()

	for i := 0; i < 1000; i++ {
		_ = myWal.AddRecord("kljuc", []byte("vivaldijeva kuca na drvetu"))
	}

	if len(myWal.segmentList) < 2 {
		t.Logf("Broj segmenata je: %d", len(myWal.segmentList))
	}
}

func TestBlockPadding(t *testing.T) {
	setupTest(t)
	defer cleanupTest()

	blockKB := 4
	segmentKB := 16

	testWal, err := CreatNewWAL(segmentKB, blockKB)
	if err != nil {
		t.Fatalf("Još jedna greška: %v", err)
	}
	defer testWal.Close()

	blockBytes := 4 * 1024

	longKey := "noa_je_mnogo_zgodan_decko_i_ima_dugacak_kljuc"
	bigValue := make([]byte, 2000)

	err1 := testWal.AddRecord(longKey, bigValue)
	if err1 != nil {
		t.Fatal("Prvi upis nije prošao")
	}

	err2 := testWal.AddRecord(longKey, bigValue)
	if err2 != nil {
		t.Fatal("Drugi upis nije prošao")
	}

	if testWal.currentWritePosition%blockBytes == 0 {
		t.Log("Uspeh: zapis je upisan na početak novog bloka.")
	}
}
