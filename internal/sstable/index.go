package sstable

import "errors"

type summaryEntry struct {
	minKey string
	maxKey string
	pos    uint32
}

type summary struct {
}

func (s *summary) Find(key string) (uint32, error) {
	return 0, errors.New("not implemented")
}

type indexEntry struct {
	key string
	pos uint32
}

type index struct {
}

func (i *index) Find(key string) (uint32, error) {
	return 0, errors.New("not implemented")
}
