package main

import (
	"fmt"

	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"

	WAL "github.com/Avram-2005/PROJEKAT_NASP/WAL"
)

func main() {
	wal, err := WAL.CreatNewWAL(16, 4, "./WAL/walDATA", 10)
	if err != nil {
		fmt.Print(err)
	}

	wal.AddRecord("Avram", []byte("string is string and string is not string. it's just string and it is what it is and we can't do anything about it because it's just string and we can't change it and we have to accept it and move on with our lives and stop trying to change the nature of string because it's just string and it will always be string no matter what we do with it and we should just embrace the fact that it's string and not try to make it something else because it's just string and that's all it will ever be."))
	wal.DeleteRecord("Avram")
	wal.AddRecord("Avram3", []byte("string"))
	wal.DeleteRecord("Avram3")
	wal.AddRecord("Avram4", []byte("string1"))
	wal.DeleteRecord("Avram4")
	wal.AddRecord("Avram1", []byte("string2"))
	wal.DeleteRecord("Avram1")
	wal.AddRecord("Avram2", []byte("string3"))
	wal.DeleteRecord("Avram2")

	conf := memtable.MemtableConfig{
		Type:              "skip_list",
		MaxSizeEntries:    5,
		SkipListMaxHeight: 8,
		BPlusTreeDegree:   2,
	}
	m, err := memtable.NewMemtableManager(3, conf, nil)
	if err != nil {
		fmt.Print(err)
	}
	wal.Recovery(m)

}
