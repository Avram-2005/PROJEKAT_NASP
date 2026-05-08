package scan

import record "github.com/Avram-2005/PROJEKAT_NASP/Record"

type SystemIterator struct {
	records []*record.Record
	index   int
	stopped bool
}

// kreira iterator za dati prefiks
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
