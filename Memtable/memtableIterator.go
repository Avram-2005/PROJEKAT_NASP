package memtable

import record "github.com/Avram-2005/PROJEKAT_NASP/Record"

type BaseIterator struct { //osnovna struktura/implementacija iteratora
	entries []*record.Record
	current int
	err     error
}

func NewBaseIterator(entries []*record.Record) *BaseIterator {
	return &BaseIterator{
		entries: entries,
		current: -1,
	}
}

func (iter *BaseIterator) Next() bool {
	if iter.err != nil {
		return false
	}
	iter.current++
	return iter.current < len(iter.entries)
}

func (iter *BaseIterator) Key() string {
	if iter.current < 0 || iter.current >= len(iter.entries) {
		return ""
	}
	return iter.entries[iter.current].Key
}

func (iter *BaseIterator) Value() []byte {
	if iter.current < 0 || iter.current >= len(iter.entries) {
		return nil
	}
	return iter.entries[iter.current].Value
}

func (iter *BaseIterator) Tombstone() bool {
	if iter.current < 0 || iter.current >= len(iter.entries) {
		return false
	}
	return iter.entries[iter.current].Tombstone
}

func (iter *BaseIterator) Reset() {
	iter.current = -1
	iter.err = nil
}

func (iter *BaseIterator) Error() error {
	return iter.err
}

func (iter *BaseIterator) Stop() {
	iter.entries = nil
	iter.current = -1
	iter.err = nil
}
