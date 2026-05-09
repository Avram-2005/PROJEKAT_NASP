package SnapshotManager

import (
	"container/list"
	"fmt"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	sstable "github.com/Avram-2005/PROJEKAT_NASP/LSM"
	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
	snapshot "github.com/Avram-2005/PROJEKAT_NASP/Snapshot"
)

// SnapshotManager je mapa koja na svakom kljucu sadrži listu verzija vrednosti na nekom ključu
// ovaj struct zapravo predstavlja mapu koja skladisti SnapshotManager-ove raznih kljuceva baze
type SnapshotManager struct {
	SnapshotManagerMap map[string]*list.List
}

// Konstruktor ne zahteva ikakve parametre-SnapshotManager objekat stvaramo pri pokretanju,
// pa ga punimo sa kljucevima koje korisnik zeli da prati
// TODO: proveriti da li mozda trebaju neki parametri
func NewSnapshotManager() (*SnapshotManager, error) {
	newMap := make(map[string]*list.List)
	return &SnapshotManager{
		SnapshotManagerMap: newMap,
	}, nil
}

// Funkcija koja ili stvara novi SnapshotManager za odredjen kljuc,
// ili dodaje novu vrednost u postojeci
// key-kljuc pod koji dodajemo
// value-niz bajtova, najnovija vrednost
func (sp *SnapshotManager) Add(key string, value snapshot.SnapshotInterface) error {
	foundList, ok := sp.SnapshotManagerMap[key]
	// ako kljuc nije pronadjen, stvaramo nov snapshot
	if !ok {
		newList := list.New()
		newList.PushFront(value)
		sp.SnapshotManagerMap[key] = newList
		return nil
	}
	// ako kljuc vec postoji, dodajemo novu vrednost na kraj njegovog SnapshotManager-a
	foundList.PushFront(value)
	elem := foundList.Front()
	elem.Next()
	return nil
}

// prototype for add many function
func (sp *SnapshotManager) AddMany(key string, memtables *[]*memtable.MemtableAdapter, sstables *[]*sstable.SSTable, sstableManager *sstable.SSTableManager) error {
	memLen := len(*memtables)
	for i := 0; i < memLen; i++ {
		newSnapshot, err := snapshot.NewSnapshotMemtable(key, (*memtables)[i])
		if err != nil {
			continue
		}
		err = sp.Add(key, newSnapshot)
		if err != nil {
			return err
		}
	}
	sstLen := len(*sstables)
	for i := 0; i < sstLen; i++ {
		newSnapshot, err := snapshot.NewSnapshotSSTable(key, (*sstables)[i], sstableManager)
		if err != nil {
			continue
		}
		err = sp.Add(key, newSnapshot)
		if err != nil {
			return err
		}
	}
	return nil
}

// Funkcija koja dobavlja odredjenu verziju naseg podatka
// key-kljuc koji se trazi
// version-verzija kljuca koju trazimo, gde je prva dodata vrednost nulta verzija,
// i svaka naredna je veca za jedan
func (sp *SnapshotManager) Get(key string, version int) (snapshot.SnapshotInterface, error) {
	foundList, ok := sp.SnapshotManagerMap[key]
	//error ako kljuc koji trazimo ne postoji
	if !ok || foundList.Len() == 0 {
		return nil, fmt.Errorf("key not found in SnapshotManager")
	}
	counter := 0
	for elem := foundList.Back(); elem != nil; elem = elem.Prev() {
		// da bi pronasli odgovarajucu verziju koristimo brojac i petlju
		if counter == version {
			return elem.Value.(snapshot.SnapshotInterface), nil
		}
		counter += 1
	}
	return nil, fmt.Errorf("version number not found for the specified key")
}

func (sp *SnapshotManager) GetList(key string) (*list.List, error) {
	value, ok := sp.SnapshotManagerMap[key]
	if ok {
		return value, nil
	}
	return nil, fmt.Errorf("no list at that key")
}

// Funkcija koja dobavlja odredjenu verziju naseg podatka
// key-kljuc koji se trazi
// timestamp-timestamp verzije koju trazimo
func (sp *SnapshotManager) GetByTimestamp(key string, timestamp time.Time) (snapshot.SnapshotInterface, error) {
	foundList, ok := sp.SnapshotManagerMap[key]
	//error ako kljuc koji trazimo ne postoji
	if !ok || foundList.Len() == 0 {
		return nil, fmt.Errorf("key not found in SnapshotManager")
	}
	for elem := foundList.Back(); elem != nil; elem = elem.Prev() {
		// da bi pronasli odgovarajucu verziju koristimo brojac i petlju
		snapshot := elem.Value.(snapshot.SnapshotInterface)
		if snapshot.GetTimestamp().Equal(timestamp) {
			return snapshot, nil
		}
	}
	return nil, fmt.Errorf("version number not found for the specified key")
}

// Funkcija koja dobavlja vrednost specificne verzije snapshot-a
// key-kljuc
// version-verzija(po redu dodavanja)
// bm-blockmanager koji ce izvuci podatke sa diska
func (sp *SnapshotManager) GetValue(key string, version int, bm *BlockManager.BlockManager) (*[]byte, error) {
	snapshot, err := sp.Get(key, version)
	if err != nil {
		return nil, err
	}
	value, err := snapshot.GetValue()
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Funkcija koja dobavlja vrednost specificne verzije snapshot-a
// key-kljuc
// timestamp-timestamp podatka koji trazimo
// bm-blockmanager koji ce izvuci podatke sa diska
func (sp *SnapshotManager) GetValueByTimestamp(key string, timestamp time.Time, bm *BlockManager.BlockManager) (*[]byte, error) {
	snapshot, err := sp.GetByTimestamp(key, timestamp)
	if err != nil {
		return nil, err
	}
	value, err := snapshot.GetValue()
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Funkcija koja dobavlja nultu verziju nase vrednosti
// key-kljuc pod kojim se trazi
func (sp *SnapshotManager) GetFirst(key string) (snapshot.SnapshotInterface, error) {
	foundList, ok := sp.SnapshotManagerMap[key]
	if !ok || foundList.Len() == 0 {
		return nil, fmt.Errorf("key not found in SnapshotManager")
	}
	return foundList.Back().Value.(snapshot.SnapshotInterface), nil
}

// Funkcija dobavlja poslednju verziju nase vrednosti
// key-kljuc pod kojim trazimo
func (sp *SnapshotManager) GetLatest(key string) (snapshot.SnapshotInterface, error) {
	foundList, ok := sp.SnapshotManagerMap[key]
	if !ok || foundList.Len() == 0 {
		return nil, fmt.Errorf("key not found in SnapshotManager")
	}
	returnValue := foundList.Front().Value.(snapshot.SnapshotInterface)
	return returnValue, nil
}

// Funkcija dobavlja prvu vrednost odredjenog kljuca
// key-kljuc
// bm-blockmanager koji ce izvuci podatke sa diska
func (sp *SnapshotManager) GetValueFirst(key string, bm *BlockManager.BlockManager) (*[]byte, error) {
	snapshot, err := sp.GetFirst(key)
	if err != nil {
		return nil, err
	}
	value, err := snapshot.GetValue()
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Funkcija dobavlja poslednju vrednost odredjenog kljuca
// key-kljuc
// bm-blockmanager koji ce izvuci podatke sa diska
func (sp *SnapshotManager) GetValueLatest(key string, bm *BlockManager.BlockManager) (*[]byte, error) {
	snapshot, err := sp.GetLatest(key)
	if err != nil {
		return nil, err
	}
	value, err := snapshot.GetValue()
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Funkcija dobavlja broj razlicitih verzija koje se skladiste za odredjen kljuc
func (sp *SnapshotManager) GetVersionCount(key string) (int, error) {
	foundList, ok := sp.SnapshotManagerMap[key]
	if !ok {
		return 0, fmt.Errorf("key is not tracked in this SnapshotManager")
	}
	return foundList.Len(), nil
}

// Izbacuje SnapshotManager za specifican kljuc iz radne memorije
func (sp *SnapshotManager) Free(key string) error {
	_, ok := sp.SnapshotManagerMap[key]
	if !ok {
		return fmt.Errorf("cannot free the memory of a nonexistent SnapshotManager")
	}
	delete(sp.SnapshotManagerMap, key)
	return nil
}

func (sp *SnapshotManager) ChangeInterfaceType(key string, version int, newVersion snapshot.SnapshotInterface) error {
	foundList, ok := sp.SnapshotManagerMap[key]
	//error ako kljuc koji trazimo ne postoji
	if !ok || foundList.Len() == 0 {
		return fmt.Errorf("key not found in SnapshotManager")
	}
	counter := 0
	for elem := foundList.Back(); elem != nil; elem = elem.Prev() {
		// da bi pronasli odgovarajucu verziju koristimo brojac i petlju
		if counter == version {
			elem.Value = newVersion
			return nil
		}
		counter += 1
	}

	return fmt.Errorf("key version not found in SnapshotManager")
}

func (sp *SnapshotManager) ChangeInterfaceTypeByTimestamp(key string, timestamp time.Time, newVersion snapshot.SnapshotInterface) error {
	foundList, ok := sp.SnapshotManagerMap[key]
	//error ako kljuc koji trazimo ne postoji
	if !ok || foundList.Len() == 0 {
		return fmt.Errorf("key not found in SnapshotManager")
	}
	for elem := foundList.Back(); elem != nil; elem = elem.Prev() {
		// da bi pronasli odgovarajucu verziju koristimo brojac i petlju
		if time.Time.Equal(elem.Value.(snapshot.SnapshotInterface).GetTimestamp(), timestamp) {
			elem.Value = newVersion
			return nil
		}
	}
	return fmt.Errorf("key version not found in SnapshotManager")
}
