package engine

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strconv"
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

	data1 := make([]byte, 100)
	binary.BigEndian.PutUint64(data1, 78)
	data2 := make([]byte, 120)
	binary.BigEndian.PutUint32(data2, 56)
	data3 := make([]byte, 80)
	binary.BigEndian.PutUint16(data3, 67)

	engine.Put("key1", data1)
	engine.Put("key2", data2)
	engine.Put("key3", data3)

	compare1, err := engine.Get("key1")
	if err != nil {
		fmt.Print(err)
		fmt.Print("Error getting key1")
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, compare1) {
		fmt.Print("data1 not equal before and after write")
		fmt.Print(data1, compare1)
		t.FailNow()
	}
	compare2, err := engine.Get("key2")
	if err != nil {
		fmt.Print(err)
		fmt.Print("Error getting key1")
		t.FailNow()
	}
	if !reflect.DeepEqual(data2, compare2) {
		fmt.Print("data2 not equal before and after write")
		fmt.Print(data2, compare2)
		t.FailNow()
	}
	compare3, err := engine.Get("key3")
	if err != nil {
		fmt.Print(err)
		fmt.Print("Error getting key1")
		t.FailNow()
	}
	if !reflect.DeepEqual(data3, compare3) {
		fmt.Print("data1 not equal before and after write")
		fmt.Print(data3, compare3)
		t.FailNow()
	}
	engine.ShutDown()

	//simulating engine shut down and boot up

	engine2, err := NewEngine(configPath, walPath, sstPath)
	if err != nil {
		fmt.Print("Error starting database up again")
		fmt.Print(err)
		t.FailNow()
	}

	compare1, err = engine2.Get("key1")
	if err != nil {
		fmt.Print(err)
		fmt.Print("Error getting key1")
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, compare1) {
		fmt.Print("data1 not equal before and after write")
		fmt.Print(data1, compare1)
		t.FailNow()
	}
	compare2, err = engine2.Get("key2")
	if err != nil {
		fmt.Print(err)
		fmt.Print("Error getting key1")
		t.FailNow()
	}
	if !reflect.DeepEqual(data2, compare2) {
		fmt.Print("data2 not equal before and after write")
		fmt.Print(data2, compare2)
		t.FailNow()
	}
	compare3, err = engine2.Get("key3")
	if err != nil {
		fmt.Print(err)
		fmt.Print("Error getting key1")
		t.FailNow()
	}
	if !reflect.DeepEqual(data3, compare3) {
		fmt.Print("data1 not equal before and after write")
		fmt.Print(data3, compare3)
		t.FailNow()
	}

	engine2.Put("key2", data1)
	compare4, err := engine2.Get("key2")
	if err != nil {
		fmt.Print(err)
		fmt.Print("Error getting key2")
		t.FailNow()
	}
	if !reflect.DeepEqual(data1, compare4) {
		fmt.Print("data1 not equal before and after write")
		fmt.Print(data1, compare4)
		t.FailNow()
	}

	engine2.ShutDown()

	err = os.RemoveAll(walPath)
	if err != nil {
		fmt.Print("Deleting WAL directory failed!")
		fmt.Print(err)
		t.FailNow()
	}
	err = os.RemoveAll(sstPath)
	if err != nil {
		fmt.Print("Deleting SSTable directory failed!")
		fmt.Print(err)
		t.FailNow()
	}
}

func TestStressEngineFunction(t *testing.T) {
	configPath := "engineConfig_test.yaml"
	walPath := "TestDataBase/walDATA"
	sstPath := "TestDataBase/sstable"

	engine, err := NewEngine(configPath, walPath, sstPath)
	if err != nil {
		fmt.Print("Engine initialization failed!\n")
		fmt.Print(err)
		t.FailNow()
	}

	n := 5000

	dataArray := make([][]byte, int(n))

	for i := 0; i < n; i++ {
		temp := make([]byte, 100)
		random := uint32(rand.Intn(100))
		binary.BigEndian.PutUint32(temp, random)
		dataArray[i] = temp
		key := "key" + strconv.Itoa(i)
		engine.Put(key, temp)
	}

	for i := n - 1; i >= 0; i-- {
		key := "key" + strconv.Itoa(i)
		temp, err := engine.Get(key)
		if err != nil {
			fmt.Print("error getting key: " + key + "\n")
			fmt.Print(err)
			t.FailNow()
		}
		if !reflect.DeepEqual(dataArray[i], temp) {
			fmt.Print("key: " + key + " not the same after put and get" + "\n")
			fmt.Print(dataArray[i], temp)
			t.FailNow()
		}
	}

	engine.ShutDown()

	engine2, err := NewEngine(configPath, walPath, sstPath)
	if err != nil {
		fmt.Print("Error after booting up engine again")
		fmt.Print(err)
		t.FailNow()
	}

	for i := n - 1; i >= 0; i-- {
		key := "key" + strconv.Itoa(i)
		temp, err := engine2.Get(key)
		if err != nil {
			fmt.Print("error getting key after second boot up: " + key + "\n")
			fmt.Print(err)
			t.FailNow()
		}
		if !reflect.DeepEqual(dataArray[i], temp) {
			fmt.Print("key: " + key + " not the same after put and get after second boot up" + "\n")
			fmt.Print(dataArray[i], temp)
			t.FailNow()
		}
	}

	engine2.ShutDown()

	err = os.RemoveAll(walPath)
	if err != nil {
		fmt.Print("Deleting WAL directory failed!")
		fmt.Print(err)
		t.FailNow()
	}
	err = os.RemoveAll(sstPath)
	if err != nil {
		fmt.Print("Deleting SSTable directory failed!")
		fmt.Print(err)
		t.FailNow()
	}
}
