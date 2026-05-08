package scan

import (
	"sort"

	sstable "github.com/Avram-2005/PROJEKAT_NASP/LSM"
	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

type ScanResult struct {
	Records    []*record.Record
	TotalCount int
	PageNumber int
	PageSize   int
	HasMore    bool
}

type SystemScanner struct {
	memtable *memtable.MemtableManager
	lsm      *sstable.LSM
}

func NewSystemScanner(mm *memtable.MemtableManager, lsm *sstable.LSM) *SystemScanner {
	return &SystemScanner{
		memtable: mm,
		lsm:      lsm,
	}
}

func (s *SystemScanner) RangeScan(startKey, endKey string, pageNumber, pageSize int) (*ScanResult, error) {
	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	//sakupljanje records iz svih instanci iz memtable-a
	allRecords := make([]*record.Record, 0)
	records := s.memtable.RangeScan(startKey, endKey)
	allRecords = append(allRecords, records...)

	//sakupljanje iz sstable-a
	sstabRecords, err := s.lsm.RangeScan(startKey, endKey)
	if err != nil {
		return nil, err
	}
	for i := range sstabRecords {
		allRecords = append(allRecords, sstabRecords[i])
	}
	//sortiranje po kljucu
	sort.Slice(allRecords, func(i, j int) bool {
		return allRecords[i].Key < allRecords[j].Key
	})
	//zadrzavanje novijeg timestampa i uklanjanje duplikata (tombstone)
	unique := make(map[string]*record.Record)
	for _, rec := range allRecords {
		if exists, ok := unique[rec.Key]; !ok || rec.Timestamp.After(exists.Timestamp) {
			unique[rec.Key] = rec
		}
	}
	//filtriranje tombstone-a
	activeRecords := make([]*record.Record, 0)
	for _, rec := range unique {
		if !rec.Tombstone {
			activeRecords = append(activeRecords, rec)
		}
	}
	//sortiranje po kljucu
	sort.Slice(activeRecords, func(i, j int) bool {
		return activeRecords[i].Key < activeRecords[j].Key
	})
	//paginacija
	startIndex := (pageNumber - 1) * pageSize
	if startIndex >= len(activeRecords) {
		return &ScanResult{
			Records:    []*record.Record{},
			TotalCount: len(activeRecords),
			PageNumber: pageNumber,
			PageSize:   pageSize,
			HasMore:    false,
		}, nil
	}

	endIndex := startIndex + pageSize
	if endIndex > len(activeRecords) {
		endIndex = len(activeRecords)
	}
	return &ScanResult{
		Records:    activeRecords[startIndex:endIndex],
		TotalCount: len(activeRecords),
		PageNumber: pageNumber,
		PageSize:   pageSize,
		HasMore:    endIndex < len(activeRecords),
	}, nil
}

func (s *SystemScanner) PrefixScan(prefix string, pageNumber, pageSize int) (*ScanResult, error) {
	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	allRecords := make([]*record.Record, 0)
	records := s.memtable.PrefixScan(prefix) //sakupljanje records iz svih instanci iz memtable-a
	allRecords = append(allRecords, records...)

	sstabRecords, err := s.lsm.PrefixScan(prefix) //sakupljanje svih records iz sstable-a
	if err != nil {
		return nil, err
	}
	for i := range sstabRecords {
		allRecords = append(allRecords, sstabRecords[i])
	}

	sort.Slice(allRecords, func(i, j int) bool { //sortiranje po kljucu
		return allRecords[i].Key < allRecords[j].Key
	})

	unique := make(map[string]*record.Record)
	for _, rec := range allRecords { //uklanjanje duplikata i zadrzavanje novijeg timestampa
		if exist, ok := unique[rec.Key]; !ok || rec.Timestamp.After(exist.Timestamp) {
			unique[rec.Key] = rec
		}
	}

	activeRecords := make([]*record.Record, 0)
	for _, rec := range unique { //filtriranje tombsone-a
		if !rec.Tombstone {
			activeRecords = append(activeRecords, rec)
		}
	}

	sort.Slice(allRecords, func(i, j int) bool { //sortiranje po kljucu
		return allRecords[i].Key < allRecords[j].Key
	})

	//paginacija
	startIndex := (pageNumber - 1) * pageSize
	if startIndex >= len(activeRecords) {
		return &ScanResult{
			Records:    []*record.Record{},
			TotalCount: len(activeRecords),
			PageNumber: pageNumber,
			PageSize:   pageSize,
			HasMore:    false,
		}, nil
	}

	endIndex := startIndex + pageSize
	if endIndex > len(activeRecords) {
		endIndex = len(activeRecords)
	}

	return &ScanResult{
		Records:    activeRecords[startIndex:endIndex],
		TotalCount: len(activeRecords),
		PageNumber: pageNumber,
		PageSize:   pageSize,
		HasMore:    endIndex < len(activeRecords),
	}, nil
}
