package config

import (
	"bytes"
	"os"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"go.yaml.in/yaml/v4"
)

type Config struct {
	BufferPoolConfig struct {
		MaxSize   int `yaml:"MaxSize"`
		BlockSize int `yaml:"BlockSize"`
	} `yaml:"BufferPoolConfig"`
}

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
	return nil
}
