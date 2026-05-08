package scan

import (
	"os"
	"testing"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	sstable "github.com/Avram-2005/PROJEKAT_NASP/LSM"
	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

// helper funkcije
func setupTestScanner(t *testing.T) (*SystemScanner, func()) {
	memConfig := memtable.MemtableConfig{
		Type:              "skip_list",
		MaxSizeEntries:    100,
		SkipListMaxHeight: 8,
	}
	mm, err := memtable.NewMemtableManager(3, memConfig, nil)
	if err != nil {
		t.Fatalf("Failed to create memtable manager: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	bm, err := BlockManager.NewBlockManager(10, 4)
	if err != nil {
		t.Fatalf("Failed to create block manager: %v", err)
	}

	sstConfig := sstable.SSTableConfig{
		SummaryInterval: 2,
		MultipleFiles:   false,
	}

	lsmConfig := sstable.LSMConfig{
		NumLevels:        3,
		CompactionFactor: 5,
	}

	lsm, err := sstable.NewLSM(lsmConfig, tmpDir, sstConfig, bm)
	if err != nil {
		t.Fatalf("Failed to create LSM: %v", err)
	}

	scanner := NewSystemScanner(mm, lsm)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return scanner, cleanup
}

func makeRecord(key, value string) *record.Record {
	return &record.Record{
		Key:       key,
		Value:     []byte(value),
		Tombstone: false,
		Timestamp: time.Now(),
	}
}

func TestRangeScanEmptySystem(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()

	result, err := scanner.RangeScan("a", "z", 1, 10)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}

	if len(result.Records) != 0 {
		t.Fatalf("Expected 0 records, got %d", len(result.Records))
	}
	if result.TotalCount != 0 {
		t.Fatalf("Expected TotalCount=0, got %d", result.TotalCount)
	}
	if result.HasMore {
		t.Fatal("HasMore should be false for empty result")
	}
}

func TestRangeScanOnlyMemtable(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()
	scanner.memtable.Put("apple", []byte("fruit1"))
	scanner.memtable.Put("banana", []byte("fruit2"))
	scanner.memtable.Put("cherry", []byte("fruit3"))
	scanner.memtable.Put("date", []byte("fruit4"))
	scanner.memtable.Put("elderberry", []byte("fruit5"))

	result, err := scanner.RangeScan("b", "e", 1, 10)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}

	expectedKeys := []string{"banana", "cherry", "date"}
	if len(result.Records) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result.Records))
	}

	for i, rec := range result.Records {
		if rec.Key != expectedKeys[i] {
			t.Fatalf("Expected %s, got %s", expectedKeys[i], rec.Key)
		}
	}
	if result.TotalCount != 3 {
		t.Fatalf("Expected TotalCount=3, got %d", result.TotalCount)
	}
}

func TestRangeScanPagination(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()

	for i := 0; i < 10; i++ {
		scanner.memtable.Put(string(rune('a'+i)), []byte("value"))
	}

	result, err := scanner.RangeScan("a", "z", 1, 3)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}
	if len(result.Records) != 3 {
		t.Fatalf("Page 1: expected 3 records, got %d", len(result.Records))
	}
	if result.TotalCount != 10 {
		t.Fatalf("Expected TotalCount=10, got %d", result.TotalCount)
	}
	if !result.HasMore {
		t.Fatal("HasMore should be true for page 1")
	}

	result, err = scanner.RangeScan("a", "z", 2, 3)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}
	if len(result.Records) != 3 {
		t.Fatalf("Page 2: expected 3 records, got %d", len(result.Records))
	}
	if !result.HasMore {
		t.Fatal("HasMore should be true for page 2")
	}
	result, err = scanner.RangeScan("a", "z", 4, 3)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}
	if len(result.Records) != 1 {
		t.Fatalf("Page 4: expected 1 record, got %d", len(result.Records))
	}
	if result.HasMore {
		t.Fatal("HasMore should be false for last page")
	}

	result, err = scanner.RangeScan("a", "z", 10, 3)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}
	if len(result.Records) != 0 {
		t.Fatalf("Page out of range: expected 0 records, got %d", len(result.Records))
	}
}

func TestRangeScanWithDeletion(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()

	scanner.memtable.Put("apple", []byte("fruit1"))
	scanner.memtable.Put("banana", []byte("fruit2"))
	scanner.memtable.Put("cherry", []byte("fruit3"))
	scanner.memtable.Delete("banana", time.Now())

	result, err := scanner.RangeScan("a", "z", 1, 10)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}

	expectedKeys := []string{"apple", "cherry"}
	if len(result.Records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(result.Records))
	}

	for i, rec := range result.Records {
		if rec.Key != expectedKeys[i] {
			t.Fatalf("Expected %s, got %s", expectedKeys[i], rec.Key)
		}
	}
}

func TestRangeScanWithUpdate(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()

	scanner.memtable.Put("apple", []byte("old_value"))
	time.Sleep(1 * time.Millisecond)
	scanner.memtable.Put("apple", []byte("new_value"))

	result, err := scanner.RangeScan("a", "z", 1, 10)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}

	if len(result.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result.Records))
	}
	if string(result.Records[0].Value) != "new_value" {
		t.Fatalf("Expected 'new_value', got %s", result.Records[0].Value)
	}
}

func TestRangeScanDefaultPageValues(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		scanner.memtable.Put(string(rune('a'+i)), []byte("value"))
	}

	result, err := scanner.RangeScan("a", "z", 0, 0)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}

	if len(result.Records) != 5 {
		t.Fatalf("Expected 5 records with default values, got %d", len(result.Records))
	}
	if result.PageNumber != 1 {
		t.Fatalf("Expected PageNumber=1, got %d", result.PageNumber)
	}
	if result.PageSize != 10 {
		t.Fatalf("Expected PageSize=10, got %d", result.PageSize)
	}
}

func TestRangeScanNoResults(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()

	scanner.memtable.Put("apple", []byte("fruit1"))
	scanner.memtable.Put("banana", []byte("fruit2"))

	result, err := scanner.RangeScan("x", "z", 1, 10)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}

	if len(result.Records) != 0 {
		t.Fatalf("Expected 0 records, got %d", len(result.Records))
	}
	if result.TotalCount != 0 {
		t.Fatalf("Expected TotalCount=0, got %d", result.TotalCount)
	}
}

func TestRangeScanBoundary(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()

	scanner.memtable.Put("a", []byte("1"))
	scanner.memtable.Put("b", []byte("2"))
	scanner.memtable.Put("c", []byte("3"))

	result, err := scanner.RangeScan("a", "a", 1, 10)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}
	if len(result.Records) != 1 || result.Records[0].Key != "a" {
		t.Fatalf("Expected only 'a', got %v", result.Records)
	}

	result, err = scanner.RangeScan("c", "c", 1, 10)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}
	if len(result.Records) != 1 || result.Records[0].Key != "c" {
		t.Fatalf("Expected only 'c', got %v", result.Records)
	}
}

func TestRangeScanMixedSources(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()
	scanner.memtable.Put("apple", []byte("mem_value"))
	scanner.memtable.Put("date", []byte("mem_date"))
	sstRecords := []*record.Record{
		makeRecord("apple", "sst_old_value"),
		makeRecord("banana", "sst_banana"),
		makeRecord("cherry", "sst_cherry"),
	}
	if err := scanner.lsm.Flush(sstRecords); err != nil {
		t.Fatalf("Failed to flush to SSTable: %v", err)
	}
	result, err := scanner.RangeScan("a", "z", 1, 10)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}
	if len(result.Records) != 4 {
		t.Fatalf("Expected 4 records, got %d", len(result.Records))
	}
	for _, rec := range result.Records {
		if rec.Key == "apple" && string(rec.Value) != "mem_value" {
			t.Fatalf("Apple should be from Memtable with 'mem_value', got '%s'", rec.Value)
		}
	}
}

func TestRangeScanWithMultipleSSTables(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()
	records1 := []*record.Record{
		makeRecord("apple", "v1"),
		makeRecord("banana", "v1"),
	}
	if err := scanner.lsm.Flush(records1); err != nil {
		t.Fatalf("Failed first flush: %v", err)
	}
	time.Sleep(1 * time.Millisecond)
	records2 := []*record.Record{
		makeRecord("apple", "v2"),
		makeRecord("cherry", "v2"),
	}
	if err := scanner.lsm.Flush(records2); err != nil {
		t.Fatalf("Failed second flush: %v", err)
	}

	result, err := scanner.RangeScan("a", "z", 1, 10)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}

	for _, rec := range result.Records {
		if rec.Key == "apple" {
			if string(rec.Value) != "v2" {
				t.Fatalf("Apple should have value 'v2', got '%s'", rec.Value)
			}
		}
	}
}

func TestRangeScanWithTombstoneInSSTable(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()
	records1 := []*record.Record{
		makeRecord("apple", "fruit"),
	}
	if err := scanner.lsm.Flush(records1); err != nil {
		t.Fatalf("Failed first flush: %v", err)
	}
	records2 := []*record.Record{
		&record.Record{Key: "apple", Value: nil, Tombstone: true, Timestamp: time.Now().Add(1 * time.Millisecond)},
	}
	if err := scanner.lsm.Flush(records2); err != nil {
		t.Fatalf("Failed tombstone flush: %v", err)
	}

	result, err := scanner.RangeScan("a", "z", 1, 10)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}

	for _, rec := range result.Records {
		if rec.Key == "apple" {
			t.Fatal("Apple should be deleted (tombstone)")
		}
	}
}

func TestRangeScanMemtableHasNewerDataThanSSTable(t *testing.T) {
	scanner, cleanup := setupTestScanner(t)
	defer cleanup()
	sstRecords := []*record.Record{
		makeRecord("apple", "sst_value"),
	}
	if err := scanner.lsm.Flush(sstRecords); err != nil { // Flush u sstable sa starom vrednoscu
		t.Fatalf("Failed flush to SSTable: %v", err)
	}
	time.Sleep(1 * time.Millisecond)
	scanner.memtable.Put("apple", []byte("mem_value")) // Memtable dobija noviju vrednost

	result, err := scanner.RangeScan("a", "z", 1, 10)
	if err != nil {
		t.Fatalf("RangeScan failed: %v", err)
	}

	if len(result.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result.Records))
	}
	if string(result.Records[0].Value) != "mem_value" {
		t.Fatalf("Apple should be from Memtable with 'mem_value', got '%s'", result.Records[0].Value)
	}
}
