package engine

import "testing"

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
