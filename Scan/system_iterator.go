package scan

import record "github.com/Avram-2005/PROJEKAT_NASP/Record"

type SystemIterator struct {
	records []*record.Record
	index   int
	stopped bool
}

// kreira iterator za zadati prefiks
func (s *SystemScanner) NewSystemPrefixIterator(prefix string) (*SystemIterator, error) {
	result, err := s.PrefixScan(prefix, 1, 1000000)
	if err != nil {
		return nil, err
	}
	return &SystemIterator{
		records: result.Records,
		index:   -1,
		stopped: false,
	}, nil
}

//kreira iterator za zadati opseg
func (s *SystemScanner) NewSystemRangeIterator(startKey, endKey string) (*SystemIterator, error) {
	result, err := s.RangeScan(startKey, endKey, 1, 1000000)
	if err != nil {
		return nil, err
	}
	return &SystemIterator{
		records: result.Records,
		index:   -1,
		stopped: false,
	}, nil
}

//pomera iterator na sledeci zapis
func (it *SystemIterator) Next() bool {
	if it.stopped || it.index+1 >= len(it.records) {
		return false
	}
	it.index++
	return true
}

//vraca trenutni kljuc
func (it *SystemIterator) Key() string {
	if it.index < 0 || it.index >= len(it.records) {
		return ""
	}
	return it.records[it.index].Key
}

//vraca trenutnu vrednost
func (it *SystemIterator) Value() []byte {
	if it.index < 0 || it.index >= len(it.records) {
		return nil
	}
	return it.records[it.index].Value
}

func (it *SystemIterator) Tombstone() bool {
	return false //uvek vraca false jer su filtrirani
}

//zaustavlja iterator i oslobadja resurse
func (it *SystemIterator) Stop() {
	it.stopped = true
	it.records = nil
	it.index = -1
}

//vraca iterator na poocetak
func (it *SystemIterator) Reset() {
	if !it.stopped {
		it.index = -1
	}
}
