package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	"go.yaml.in/yaml/v4"
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

	config.BufferPoolConfig.BlockSize = 4
	config.BufferPoolConfig.MaxSize = 4
	data, err := yaml.Marshal(config)
	if err != nil {
		fmt.Print(err)
		t.FailNow()
	}
	fmt.Print(string(data))
}
