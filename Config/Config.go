package config

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	cache "github.com/Avram-2005/PROJEKAT_NASP/Cache"
	sstable "github.com/Avram-2005/PROJEKAT_NASP/LSM"
	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
	wal "github.com/Avram-2005/PROJEKAT_NASP/WAL"

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
		MaxSizeEntries    int    `yaml:"MaxSizeEntries"`
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
		SegmentSize int    `yaml:"SegmentSize"`
		FilePath    string `yaml:"FilePath"`
	} `yaml:"WriteAheadLogConfig"`
	LSMConfig struct {
		NumLevels        int `yaml:"NumLevels"`
		CompactionFactor int `yaml:"CompactionFactor"`
	} `yaml:"LSMConfig"`
}

//TODO: delete when structs are merged to develop

/*type MemtableManager struct {
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

type LSM struct {
}

func NewLSM(lsmConfig LSMConfig, tablesRoot string, sstConfig SSTableConfig, bm *BlockManager.BlockManager) (*LSM, error) {
	return &LSM{}, nil
}

type LSMConfig struct {
	NumLevels        int
	CompactionFactor int
}*/

type TokenBucket struct {
}

func NewTokenBucket(maxNumTokens int64, refillInterval time.Duration) (*TokenBucket, error) {
	return &TokenBucket{}, nil
}

//Config related functions

func NewConfig() *Config {
	return &Config{}
}

func (config *Config) InitializeDefault() error {
	defaultValue := `BufferPoolConfig:
  #Amount of blocks stored 
  MaxSize: 8
  #Size of stored blocks 
  BlockSize: 4
CacheConfig:
  #Amount of key value pairs stored
  MaxSize: 20
MemtableConfig:
  #Amount of memtable instances active at one point
  MaxCount: 2
  #Whether memtable uses hashmap, skiplist or btree
  Type: hashmap
  #How large the memtable can be in bytes
  MaxSizeBytes: 100
  #How many entries memtable can contain
  MaxSizeEntries: 100
  #Bplus tree configuration
  BPlusTreeDegree: 0
  #SkipList configuration
  SkipListMaxHeight: 0
TokenBucketConfig:
  #How many tokens every user has
  MaxNumTokens: 3
  #Number of seconds between refills
  RefillTime: 60
SSTableConfig:
  TablesRoot: sstables
  SummaryInterval: 40
  MultipleFiles: false
WriteAheadLogConfig:
  SegmentSize: 40
LSMConfig:
  NumLevels: 4
  CompactionFactor: 4`
	bytesDefault := []byte(defaultValue)
	err := yaml.Unmarshal(bytesDefault, config)
	if err != nil {
		return err
	}
	return nil
}

func (config *Config) Initialize(bm *BlockManager.BlockManager, configFile *os.File) error {
	defaultConfig := NewConfig()
	err := defaultConfig.InitializeDefault()
	if err != nil {
		return err
	}
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
	bm, err = config.InitializeBlockManager()
	if err != nil {
		fmt.Print("Blockmanager configuration is incorrect, default configuration will be used.")
		config.BufferPoolConfig = defaultConfig.BufferPoolConfig
		isConfigValid = false
	}
	_, err = config.InitializeCache()
	if err != nil {
		fmt.Print("Cache configuration is incorrect, default configuration will be used.")
		config.CacheConfig = defaultConfig.CacheConfig
		isConfigValid = false
	}
	_, err = config.InitializeMemtable()
	if err != nil {
		fmt.Print("Memtable configuration is incorrect, default configuration will be used.")
		config.MemtableConfig = defaultConfig.MemtableConfig
		isConfigValid = false
	}
	_, err = config.InitializeTokenBucket()
	if err != nil {
		fmt.Print("Token bucket configuration is incorrect, default configuration will be used.")
		config.TokenBucketConfig = defaultConfig.TokenBucketConfig
		isConfigValid = false
	}
	_, err = config.InitializeSSTable(bm)
	if err != nil {
		fmt.Print("SSTable configuration is incorrect, default configuration will be used.")
		config.SSTableConfig = defaultConfig.SSTableConfig
		isConfigValid = false
	}
	_, err = config.InitializeWAL()
	if err != nil {
		fmt.Print("WAL configuration is incorrect, default configuration will be used.")
		config.WriteAheadLogConfig = defaultConfig.WriteAheadLogConfig
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

func (config *Config) InitializeMemtable() (*memtable.MemtableManager, error) {
	memtableConfig := &memtable.MemtableConfig{
		Type:              config.MemtableConfig.Type,
		MaxSizeBytes:      config.MemtableConfig.MaxSizeBytes,
		MaxSizeEntries:    config.MemtableConfig.MaxSizeEntries,
		BPlusTreeDegree:   config.MemtableConfig.BPlusTreeDegree,
		SkipListMaxHeight: config.MemtableConfig.SkipListMaxHeight,
	}
	return memtable.NewMemtableManager(config.MemtableConfig.MaxCount, *memtableConfig, nil)
}

func (config *Config) InitializeCache() (*cache.Cache, error) {
	return cache.NewCache(config.CacheConfig.MaxSize)
}

func (config *Config) InitializeTokenBucket() (*TokenBucket, error) {
	return NewTokenBucket(int64(config.TokenBucketConfig.MaxNumTokens), time.Millisecond*time.Duration(config.TokenBucketConfig.RefillTime))
}

func (config *Config) InitializeSSTable(bm *BlockManager.BlockManager) (*sstable.SSTableManager, error) {
	sstableConfig := &sstable.SSTableConfig{
		SummaryInterval: config.SSTableConfig.SummaryInterval,
		MultipleFiles:   config.SSTableConfig.MultipleFiles,
	}
	return sstable.SetupSSTableManager(config.SSTableConfig.TablesRoot, *sstableConfig, bm)
}

func (config *Config) InitializeWAL() (*wal.WAL, error) {
	return wal.CreatNewWAL(config.WriteAheadLogConfig.SegmentSize, config.BufferPoolConfig.BlockSize, config.WriteAheadLogConfig.FilePath, config.MemtableConfig.MaxCount)
}

func (config *Config) InitializeLSM(bm *BlockManager.BlockManager) (*sstable.LSM, error) {
	LSMConfig := &sstable.LSMConfig{
		NumLevels:        config.LSMConfig.NumLevels,
		CompactionFactor: config.LSMConfig.CompactionFactor,
	}
	SSTableConfig := &sstable.SSTableConfig{
		SummaryInterval: config.SSTableConfig.SummaryInterval,
		MultipleFiles:   config.SSTableConfig.MultipleFiles,
	}
	return sstable.NewLSM(*LSMConfig, config.SSTableConfig.TablesRoot, *SSTableConfig, bm)
}
