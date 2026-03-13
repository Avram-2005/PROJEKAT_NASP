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

func (ch *Cache) Put(key string, value *[]byte) error {
	_, ok := ch.cacheMap[key]
	if !ok {
		if ch.currentSize >= ch.maxSize {
			err := ch.evictOldest()
			if err != nil {
				return err
			}
		}
		ch.lruList.PushFront(key)
		ch.cacheMap[key] = *value
		ch.currentSize += 1
		return nil
	}
	elem, err := ch.findElement(key)
	if err != nil {
		return err
	}
	ch.lruList.MoveToFront(elem)
	ch.cacheMap[key] = *value
	return nil
}

func (ch *Cache) Get(key string) (*[]byte, error) {
	value, ok := ch.cacheMap[key]
	if !ok {
		return nil, fmt.Errorf("tried getting a nonexistent elemnt")
	}
	elem, err := ch.findElement(key)
	if err != nil {
		return nil, err
	}
	ch.lruList.MoveToFront(elem)
	return &value, nil
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
	return nil
}
