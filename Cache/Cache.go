package Cache

import (
	"container/list"
	"fmt"
)

type Cache struct {
	maxSize     int               //maksimalna velicina cache-a
	currentSize int               //trenutna velicina cache-a
	lruList     *list.List        //lista za lru algoritam
	cacheMap    map[string][]byte //mapa koja sadrzi parove kljuc vrednost
}

// konstruktor za cache-mora samo biti veci od 0
func NewCache(size int) (*Cache, error) {
	if size == 0 {
		return nil, fmt.Errorf("Cache cannot be size 0")
	}
	cacheMap := make(map[string][]byte, size)
	return &Cache{
		maxSize:     size,
		currentSize: 0,
		lruList:     list.New(),
		cacheMap:    cacheMap,
	}, nil
}

// parametri su kljuc tipa string i niz bajtova proizvoljne velicine za vrednost
func (ch *Cache) Put(key string, value *[]byte) error {
	//proveravamo jel kljuc vec tu
	_, ok := ch.cacheMap[key]
	if !ok {
		//ako nije ubacujemo ga, proveravajuci current size
		if ch.currentSize >= ch.maxSize {
			//mehanizam brisanja
			err := ch.evictOldest()
			if err != nil {
				return err
			}
		}
		//dodajemo nov el i povecavamo current size
		ch.lruList.PushFront(key)
		ch.cacheMap[key] = *value
		ch.currentSize += 1
		return nil
	}
	//nalazimo el i pomeramo ga napred ako je vec tu
	elem, err := ch.findElement(key)
	if err != nil {
		return err
	}
	ch.lruList.MoveToFront(elem)
	ch.cacheMap[key] = *value
	return nil
}

func (ch *Cache) Get(key string) (*[]byte, error, bool) {
	//trazimo kljuc
	value, ok := ch.cacheMap[key]
	if !ok {
		return nil, nil, false
	}
	//pomeramo element napred
	elem, err := ch.findElement(key)
	if err != nil {
		return nil, err, false
	}
	ch.lruList.MoveToFront(elem)
	return &value, nil, true
}

func (ch *Cache) findElement(key string) (*list.Element, error) {
	if ch.currentSize == 0 {
		return nil, fmt.Errorf("cannot find an element in an empty cache")
	}
	for elem := ch.lruList.Front(); elem != nil; elem = elem.Next() {
		if elem.Value == key {
			return elem, nil
		}
	}
	return nil, fmt.Errorf("element not found")
}

func (ch *Cache) evictOldest() error {
	if ch.currentSize == 0 {
		return fmt.Errorf("cannot remove an element from an empty cache")
	}
	oldest := ch.lruList.Back()
	ch.lruList.Remove(oldest)
	delete(ch.cacheMap, oldest.Value.(string))
	ch.currentSize -= 1
	return nil
}
