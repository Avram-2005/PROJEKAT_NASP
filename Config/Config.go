package config

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"go.yaml.in/yaml/v4"
)

type Config struct {
	BufferPoolConfig struct {
		MaxSize   int `yaml:"MaxSize"`
		BlockSize int `yaml:"BlockSize"`
	} `yaml:"BufferPoolConfig"`
	CacheConfig struct {
		MaxSize int `yaml:"MaxSize"`
	} `yaml:"CacheConfig"`
	MemtableConfig struct {
		MaxCount          int    `yaml:"MaxCount"`
		Type              string `yaml:"Type"`
		MaxSizeBytes      int    `yaml:"MaxSizeBytes"`
		MaxSizeEntries    int    `yaml:"MaxSizeBytes"`
		BPlusTreeDegree   int    `yaml:"BPlusTreeDegree"`
		SkipListMaxHeight int    `yaml:"SkipListMaxHeight"`
	} `yaml:"MemtableConfig"`
	TokenBucketConfig struct {
		MaxNumTokens int64 `yaml:"MaxNumTokens"`
		RefillTime   int   `yaml:"RefillTime"`
	} `yaml:"TokenBucketConfig"`
	SSTableConfig struct {
		TablesRoot      string `yaml:"TablesRoot"`
		SummaryInterval int    `yaml:"SummaryInterval"`
		MultipleFiles   bool   `yaml:"MultipleFiles"`
	} `yaml:"SSTableConfig"`
	WriteAheadLogConfig struct {
		Segmented   bool `yaml:"Segmented"`
		SegmentSize int  `yaml:"SegmentSize"`
	} `yaml:"WriteAheadLogConfig"`
}

//TODO: delete when structs are merged to develop

type MemtableManager struct {
}

type MemtableConfig struct {
	Type              string //neka od tri strukture: hashmapa, skiplista ili b+ stablo
	MaxSizeBytes      int    //max velicina u bajtovima
	MaxSizeEntries    int    //max broj elemenata koji moze da primi
	BPlusTreeDegree   int    //max stepen stabla
	SkipListMaxHeight int    //max visina skipliste

}

func FakeFlush(kv []KeyValue) error {
	return nil
}

type KeyValue struct {
	Key       string
	Value     []byte
	Tombstone bool //za brisanje, true ako je obrisan
}

func NewMemtableManager(maxCount int, config MemtableConfig, Flush func([]KeyValue) error) (*MemtableManager, error) {
	return &MemtableManager{}, nil
}

// Fake cache

type Cache struct {
}

func NewCache(size int) (*Cache, error) {
	return &Cache{}, nil
}

// Fake token bucket

type TokenBucket struct {
}

func NewTokenBucket(maxNumTokens int64, refillInterval time.Duration) (*TokenBucket, error) {
	return &TokenBucket{}, nil
}

// Fake SSTable

type SSTableConfig struct {
	SummaryInterval int
	MultipleFiles   bool
}

type SSTableManager struct {
}

func SetupSSTableManager(root string, summaryInt int, multFiles bool, bm *BlockManager.BlockManager) (*SSTableManager, error) {
	return &SSTableManager{}, nil
}

// Fake WAL
type WAL struct {
}

func CreatNewWAL(sizeSegment int, blocksize int) (*WAL, error) {
	return &WAL{}, nil
}

//Config related functions

func NewConfig() *Config {
	return &Config{}
}

func (config *Config) Initialize(bm *BlockManager.BlockManager, configFile *os.File) error {
	data, err := bm.Get(configFile, 0)
	if err != nil {
		return err
	}
	cleanedData := bytes.TrimRight(*data, "\x00")
	err = yaml.Unmarshal(cleanedData, config)
	if err != nil {
		return err
	}
	isConfigValid := true
	_, err = config.InitializeBlockManager()
	if err != nil {
		fmt.Print("Blockmanager configuration is incorrect, default configuration will be used.")
		isConfigValid = false
	}
	_, err = config.InitializeCache()
	if err != nil {
		fmt.Print("Cache configuration is incorrect, default configuration will be used.")
		isConfigValid = false
	}
	_, err = config.InitializeMemtable()
	if err != nil {
		fmt.Print("Memtable configuration is incorrect, default configuration will be used.")
		isConfigValid = false
	}
	_, err = config.InitializeTokenBucket()
	if err != nil {
		fmt.Print("Token bucket configuration is incorrect, default configuration will be used.")
		isConfigValid = false
	}
	_, err = config.InitializeSSTable()
	if err != nil {
		fmt.Print("SSTable configuration is incorrect, default configuration will be used.")
		isConfigValid = false
	}
	_, err = config.InitializeWAL()
	if err != nil {
		fmt.Print("WAL configuration is incorrect, default configuration will be used.")
		isConfigValid = false
	}
	if !isConfigValid {
		return fmt.Errorf("configuration file has improper values")
	}
	return nil
}

func (config *Config) InitializeBlockManager() (*BlockManager.BlockManager, error) {
	return BlockManager.NewBlockManager(config.BufferPoolConfig.MaxSize, config.BufferPoolConfig.BlockSize)
}

func (config *Config) InitializeMemtable() (*MemtableManager, error) {
	memtableConfig := &MemtableConfig{
		Type:              config.MemtableConfig.Type,
		MaxSizeBytes:      config.MemtableConfig.MaxSizeBytes,
		MaxSizeEntries:    config.MemtableConfig.MaxSizeEntries,
		BPlusTreeDegree:   config.MemtableConfig.BPlusTreeDegree,
		SkipListMaxHeight: config.MemtableConfig.SkipListMaxHeight,
	}
	return NewMemtableManager(config.MemtableConfig.MaxCount, *memtableConfig, FakeFlush)
}

func (config *Config) InitializeCache() (*Cache, error) {
	return NewCache(config.CacheConfig.MaxSize)
}

func (config *Config) InitializeTokenBucket() (*TokenBucket, error) {
	return NewTokenBucket(int64(config.TokenBucketConfig.MaxNumTokens), time.Millisecond*time.Duration(config.TokenBucketConfig.RefillTime))
}

func (config *Config) InitializeSSTable() (*SSTableManager, error) {
	bm, err := config.InitializeBlockManager()
	if err != nil {
		return nil, err
	}
	return SetupSSTableManager(config.SSTableConfig.TablesRoot, config.SSTableConfig.SummaryInterval, config.SSTableConfig.MultipleFiles, bm)
}

func (config *Config) InitializeWAL() (*WAL, error) {
	return CreatNewWAL(config.WriteAheadLogConfig.SegmentSize, config.BufferPoolConfig.BlockSize)
}
