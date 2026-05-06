package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
)

func TestInitialize(t *testing.T) {
	bm, err := BlockManager.NewBlockManager(4, 4)
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	file, err := os.OpenFile("config_test.yaml", 0644, os.FileMode(os.O_RDONLY))
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	config := NewConfig()
	err = config.Initialize(bm, file)
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}

	if config.BufferPoolConfig.BlockSize != 4 || config.BufferPoolConfig.MaxSize != 8 {
		fmt.Print("Wrong bufferpool configuration")
		fmt.Print(config.BufferPoolConfig)
		t.FailNow()
	}
	if config.CacheConfig.MaxSize != 20 {
		fmt.Print("Wrong cache configuration")
		fmt.Print(config.CacheConfig)
		t.FailNow()
	}
	if config.MemtableConfig.MaxCount != 2 || config.MemtableConfig.Type != "hashmap" ||
		config.MemtableConfig.MaxSizeBytes != 100 || config.MemtableConfig.MaxSizeEntries != 100 ||
		config.MemtableConfig.BPlusTreeDegree != 0 || config.MemtableConfig.SkipListMaxHeight != 0 {
		fmt.Print("Wrong memtable configuration")
		fmt.Print(config.MemtableConfig)
		t.FailNow()
	}
	if config.TokenBucketConfig.MaxNumTokens != 3 || config.TokenBucketConfig.RefillTime != 60 {
		fmt.Print("Wrong token bucket configuration")
		fmt.Print(config.TokenBucketConfig)
		t.FailNow()
	}
	if config.SSTableConfig.TablesRoot != "./TestDataBase/sstable" || config.SSTableConfig.SummaryInterval != 40 || config.SSTableConfig.MultipleFiles != false {
		fmt.Print("Wrong sstable configuration")
		fmt.Print(config.SSTableConfig)
		t.FailNow()
	}
	if config.WriteAheadLogConfig.SegmentSize != 40 || config.WriteAheadLogConfig.FilePath != "./TestDataBase/walDATA" {
		fmt.Print("Wrong WAL configuration")
		fmt.Print(config.WriteAheadLogConfig)
		t.FailNow()
	}
	if config.LSMConfig.NumLevels != 4 || config.LSMConfig.CompactionFactor != 10 {
		fmt.Print("Wrong LSM configuration")
		fmt.Print(config.LSMConfig)
		t.FailNow()
	}
	file.Close()
	err = os.RemoveAll("./TestDataBase/sstable")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	err = os.RemoveAll("./TestDataBase/walDATA")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
}

func TestInitializeDefualt(t *testing.T) {
	config := NewConfig()
	err := config.InitializeDefault()
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}

	if config.BufferPoolConfig.BlockSize != 4 || config.BufferPoolConfig.MaxSize != 8 {
		fmt.Print("Wrong bufferpool configuration")
		fmt.Print(config.BufferPoolConfig)
		t.FailNow()
	}
	if config.CacheConfig.MaxSize != 20 {
		fmt.Print("Wrong cache configuration")
		fmt.Print(config.CacheConfig)
		t.FailNow()
	}
	if config.MemtableConfig.MaxCount != 2 || config.MemtableConfig.Type != "hashmap" ||
		config.MemtableConfig.MaxSizeBytes != 100 || config.MemtableConfig.MaxSizeEntries != 100 ||
		config.MemtableConfig.BPlusTreeDegree != 0 || config.MemtableConfig.SkipListMaxHeight != 0 {
		fmt.Print("Wrong memtable configuration")
		fmt.Print(config.MemtableConfig)
		t.FailNow()
	}
	if config.TokenBucketConfig.MaxNumTokens != 3 || config.TokenBucketConfig.RefillTime != 60 {
		fmt.Print("Wrong token bucket configuration")
		fmt.Print(config.TokenBucketConfig)
		t.FailNow()
	}
	if config.SSTableConfig.TablesRoot != "../DataBase/sstable" || config.SSTableConfig.SummaryInterval != 40 || config.SSTableConfig.MultipleFiles != false {
		fmt.Print("Wrong sstable configuration")
		fmt.Print(config.SSTableConfig)
		t.FailNow()
	}
	if config.WriteAheadLogConfig.SegmentSize != 64 || config.WriteAheadLogConfig.FilePath != "../DataBase/walDATA" {
		fmt.Print("Wrong WAL configuration")
		fmt.Print(config.WriteAheadLogConfig)
		t.FailNow()
	}
	if config.LSMConfig.NumLevels != 4 || config.LSMConfig.CompactionFactor != 10 {
		fmt.Print("Wrong LSM configuration")
		fmt.Print(config.LSMConfig)
		t.FailNow()
	}
}

func TestIncorrectConfiguration(t *testing.T) {
	bm, err := BlockManager.NewBlockManager(4, 4)
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	file, err := os.OpenFile("configIncorrect_test.yaml", 0644, os.FileMode(os.O_RDONLY))
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	config := NewConfig()
	err = config.Initialize(bm, file)
	if err == nil {
		fmt.Print("expected error to happen, but it did not")
		t.FailNow()
	}

	if config.BufferPoolConfig.BlockSize != 4 || config.BufferPoolConfig.MaxSize != 8 {
		fmt.Print("Wrong bufferpool configuration")
		fmt.Print(config.BufferPoolConfig)
		t.FailNow()
	}
	if config.CacheConfig.MaxSize != 20 {
		fmt.Print("Wrong cache configuration")
		fmt.Print(config.CacheConfig)
		t.FailNow()
	}
	if config.MemtableConfig.MaxCount != 2 || config.MemtableConfig.Type != "hashmap" ||
		config.MemtableConfig.MaxSizeBytes != 100 || config.MemtableConfig.MaxSizeEntries != 100 ||
		config.MemtableConfig.BPlusTreeDegree != 0 || config.MemtableConfig.SkipListMaxHeight != 0 {
		fmt.Print("Wrong memtable configuration")
		fmt.Print(config.MemtableConfig)
		t.FailNow()
	}
	if config.TokenBucketConfig.MaxNumTokens != 3 || config.TokenBucketConfig.RefillTime != 60 {
		fmt.Print("Wrong token bucket configuration")
		fmt.Print(config.TokenBucketConfig)
		t.FailNow()
	}
	if config.SSTableConfig.TablesRoot != "./TestDataBase/sstable" || config.SSTableConfig.SummaryInterval != 40 || config.SSTableConfig.MultipleFiles != false {
		fmt.Print("Wrong sstable configuration")
		fmt.Print(config.SSTableConfig)
		t.FailNow()
	}
	if config.WriteAheadLogConfig.SegmentSize != 40 || config.WriteAheadLogConfig.FilePath != "./TestDataBase/walData" {
		fmt.Print("Wrong WAL configuration")
		fmt.Print(config.WriteAheadLogConfig)
		t.FailNow()
	}
	if config.LSMConfig.NumLevels != 4 || config.LSMConfig.CompactionFactor != 10 {
		fmt.Print("Wrong LSM configuration")
		fmt.Print(config.LSMConfig)
		t.FailNow()
	}

	file.Close()
	err = os.RemoveAll("./TestDataBase/sstable")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	err = os.RemoveAll("./TestDataBase/walDATA")
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
}
