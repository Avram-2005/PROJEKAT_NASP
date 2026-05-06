package engine

import (
	"fmt"
	"os"
	"testing"
)

func initEngine(t *testing.T) Engine {
	root := t.TempDir()
	configPath := root + "/config.json"
	walPath := root + "/wal"
	sstablePath := root + "/sstables"

	engine, err := NewEngine(configPath, walPath, sstablePath)
	if err != nil {
		t.Fatalf("Failed to initialize engine: %v", err)
	}
	return *engine
}

func TestEngineOnePutGet(t *testing.T) {

}

func TestEngineBasicFunctions(t *testing.T) {
	configPath := "engineConfig_test.yaml"
	walPath := "TestDataBase/walDATA"
	sstPath := "TestDataBase/sstable"

	engine, err := NewEngine(configPath, walPath, sstPath)
	if err != nil {
		fmt.Print("Engine initialization failed!")
		t.FailNow()
	}

	engine.ShutDown()
	err = os.RemoveAll(walPath)
	if err != nil {
		fmt.Print("Deleting WAL directory failed!")
		t.FailNow()
	}
	err = os.RemoveAll(sstPath)
	if err != nil {
		fmt.Print("Deleting SSTable directory failed!")
		t.FailNow()
	}
}
