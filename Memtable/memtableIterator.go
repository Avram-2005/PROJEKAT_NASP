package Memtable

type BaseImplIterator struct { //osnovna struktura/implementacija iteratora
	entries []KeyValue
	current int
	err     error
}

func NewBaseImplIterator(entries []KeyValue) *BaseImplIterator {
	return &BaseImplIterator{
		entries: entries,
		current: -1,
	}
}

func (iter *BaseImplIterator) Next() bool {
	if iter.err != nil {
		return false
	}
	iter.current++
	return iter.current < len(iter.entries)
}

func (iter *BaseImplIterator) Key() string {
	if iter.current < 0 || iter.current >= len(iter.entries) {
		return ""
	}
	return iter.entries[iter.current].Key
}

func (iter *BaseImplIterator) Value() []byte {
	if iter.current < 0 || iter.current >= len(iter.entries) {
		return nil
	}
	return iter.entries[iter.current].Value
}

func (iter *BaseImplIterator) Tombstone() bool {
	if iter.current < 0 || iter.current >= len(iter.entries) {
		return false
	}
	return iter.entries[iter.current].Tombstone
}

func (iter *BaseImplIterator) Reset() {
	iter.current = -1
	iter.err = nil
}

func (iter *BaseImplIterator) Error() error {
	return iter.err
}
