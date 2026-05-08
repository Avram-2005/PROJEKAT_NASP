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
}

func TestEngineDelete(t *testing.T) {
	configPath := "engineConfig_test.yaml"
	walPath := "TestDataBase/walDATA"
	sstPath := "TestDataBase/sstable"

	engine, err := NewEngine(configPath, walPath, sstPath)
	if err != nil {
		fmt.Print("Engine initialization failed!\n")
		fmt.Print(err)
		t.FailNow()
	}

	engine.Put("delete-key", []byte("value1"))
	value, err := engine.Get("delete-key")
	if err != nil || len(value) == 0 {
		t.Fatal("Failed to get key after put")
	}

	err = engine.Delete("delete-key")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	value, err = engine.Get("delete-key")
	if err != nil || len(value) > 0 {
		fmt.Print(value)
		t.Fatal("Key should be deleted but was found")
	}

	err = engine.Delete("non-existent-key")
	if err != nil {
		t.Fatalf("Delete of non-existent key should not error: %v", err)
	}

	engine.Put("keep1", []byte("value1"))
	engine.Put("delete2", []byte("value2"))
	engine.Put("keep2", []byte("value3"))

	err = engine.Delete("delete2")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	val1, _ := engine.Get("keep1")
	if !reflect.DeepEqual(val1, []byte("value1")) {
		t.Fatal("keep1 should still exist")
	}

	val2, _ := engine.Get("keep2")
	if !reflect.DeepEqual(val2, []byte("value3")) {
		t.Fatal("keep2 should still exist")
	}

	deletedVal, _ := engine.Get("delete2")
	if len(deletedVal) > 0 {
		t.Fatal("delete2 should be deleted")
	}

	engine.ShutDown()

	engine2, err := NewEngine(configPath, walPath, sstPath)
	if err != nil {
		fmt.Print("Engine initialization failed!\n")
		fmt.Print(err)
		t.FailNow()
	}

	val1, _ = engine2.Get("keep1")
	if !reflect.DeepEqual(val1, []byte("value1")) {
		t.Fatal("keep1 should still exist after restart")
	}

	deletedVal, _ = engine2.Get("delete2")
	if len(deletedVal) > 0 {
		t.Fatal("delete2 should still be deleted after restart")
	}

	engine2.Put("delete2", []byte("new-value"))
	val, _ := engine2.Get("delete2")
	if !reflect.DeepEqual(val, []byte("new-value")) {
		t.Fatal("Re-put after delete failed")
	}

	engine2.ShutDown()
}
