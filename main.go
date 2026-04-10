package main

import (
	"fmt"

	WAL "github.com/Avram-2005/PROJEKAT_NASP/WAL"
)

func main() {
	wal, err := WAL.CreatNewWAL(4048, 4)
	if err != nil {
		fmt.Print(err)
	}

	wal.AddRecord("Avram", []byte("string"))
	wal.DeleteRecord("Avram")
	wal.AddRecord("Avram1", []byte("string"))
	wal.DeleteRecord("Avram1")
	wal.AddRecord("Avram2", []byte("string"))
	wal.DeleteRecord("Avram2")

	wal.ReadAll()

}
