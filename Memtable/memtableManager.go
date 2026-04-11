package memtable

import "fmt"

//Memtable menadzer upravlja sa N instamci memtable adaptera
//od kojih je 1 read-write, ostale su read-only
type MemtableManager struct {
	instances  []*MemtableAdapter
	maxCount   int
	baseConfig MemtableConfig
	Flush      func(entries []KeyValue) error
}

func NewMemtableManager(maxCount int, config MemtableConfig, Flush func([]KeyValue) error) (*MemtableManager, error) {
	if maxCount < 1 {
		return nil, fmt.Errorf("Number of instances of memtable must be greater or equal to 1")
	}
	first, err := NewMemtableAdapter(config)
	if err != nil {
		return nil, err
	}
	return &MemtableManager{
		instances:  []*MemtableAdapter{first},
		maxCount:   maxCount,
		baseConfig: config,
		Flush:      Flush,
	}, nil
}

//Vraca trenutno aktivnu read-write tabelu
func (mm *MemtableManager) activeTable() *MemtableAdapter {
	return mm.instances[len(mm.instances)-1]
}

//helper funkcija
//rotira tabele kada se jedna napuni
//ako su sve pune, najstarija se flushuje i aktivira novu
func (mm *MemtableManager) rotateAndFlushIfNecessary() error {
	if len(mm.instances) >= mm.maxCount {
		oldest := mm.instances[0]
		if mm.Flush != nil {
			if err := mm.Flush(oldest.GetSortedEntries()); err != nil {
				return fmt.Errorf("Error while flushing: %w", err)
			}
		}
		mm.instances = mm.instances[1:]
	}
	newTable, err := NewMemtableAdapter(mm.baseConfig)
	if err != nil {
		return err
	}
	mm.instances = append(mm.instances, newTable)
	return nil
}

//Upisuje par kljuc-vrednost u aktivnu tabelu
//Pri dostizanju N tabela, najstarija se flushuje
func (mm *MemtableManager) Put(key string, value []byte) error {
	actTab := mm.activeTable()
	if err := actTab.Put(key, value); err != nil {
		return err
	}
	if actTab.IsFull() {
		if err := mm.rotateAndFlushIfNecessary(); err != nil {
			return err
		}
	}
	return nil
}

//Upisuje tombstone u aktivnu tabelu
func (mm *MemtableManager) Delete(key string) error {
	active := mm.activeTable()
	if err := active.Put(key, nil); err != nil {
		if err2 := active.Put(key, []byte{}); err2 != nil {
			return err2
		}
	}
	if active.IsFull() {
		if err := mm.rotateAndFlushIfNecessary(); err != nil {
			return err
		}
	}
	return nil
}

//pretrazuje od najnovije ka najstarijoj tabeli, LRU princip
func (mm *MemtableManager) Get(key string) ([]byte, bool, error) {
	for i := len(mm.instances) - 1; i >= 0; i-- {
		val, found, err := mm.instances[i].Get(key)
		if err != nil {
			return nil, false, err
		}
		if found {
			return val, true, nil
		}
	}
	return nil, false, nil
}

//vraca trenutni broj instanci u memoriji
func (mm *MemtableManager) InstanceCount() int {
	return len(mm.instances)
}

//vraca ukupan broj razlicitih kljuceva u svim instancama
func (mm *MemtableManager) TotalSize() int {
	total := 0
	for _, inst := range mm.instances {
		total += inst.Size()
	}
	return total
}
