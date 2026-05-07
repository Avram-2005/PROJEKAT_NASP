package main

import (
	"fmt"
	"strings"

	eng "github.com/Avram-2005/PROJEKAT_NASP/Engine"
	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func main() {
	configPath := "config/config.yaml"
	walPath := "DataBase/walDATA"
	sstablePath := "DataBase/sstable"

	engine, err := eng.NewEngine(configPath, walPath, sstablePath)
	if err != nil {
		fmt.Println("Greska pri inicijalizaciji sistema:", err)
		return
	}

	for {
		fmt.Println()
		fmt.Print("Unesite komandu: ")
		fmt.Println("0 - UGASI SISTEM, 1 - PUT, 2 - DELETE, 3 - GET, 4 - PREFIX_SCAN, 5 - RANGE_SCAN, 6 - PREFIX_ITERATE, 7 - RANGE_ITERATE, 8 - SNAPSHOT, 9 - CHECKPOINT, 10 - VALIDACIJA_MERKLE_STABLA")

		var command int
		_, err := fmt.Scanln(&command)
		if err != nil {
			fmt.Println("Neispravan unos komande")
			continue
		}

		if command == 0 {
			engine.ShutDown()
			fmt.Println("Sistem je ugasen.")
			break
		}

		switch command {
		case 1:
			key := readLine("Unesite key: ")
			value := readLine("Unesite value: ")

			err := engine.Put(key, []byte(value))
			if err != nil {
				fmt.Println("Greska pri upisu:", err)
				continue
			}

		case 2:
			key := readLine("Unesite key za brisanje: ")

			err := engine.Delete(key)
			if err != nil {
				fmt.Println("Greska pri brisanju:", err)
				continue
			}

		case 3:
			key := readLine("Unesite key za pretragu: ")

			value, err := engine.Get(key)
			if err != nil {
				fmt.Println("Greska pri citanju:", err)
				continue
			}
			fmt.Println(string(value))

		case 4:
			prefix := readLine("Unesite prefix: ")
			records := engine.PrefixScan(prefix)
			printRecords(records)

		case 5:
			startKey := readLine("Unesite pocetni key: ")
			endKey := readLine("Unesite krajnji key: ")
			records := engine.RangeScan(startKey, endKey)
			printRecords(records)

		case 6:
			fmt.Println("PREFIX_ITERATE nije implementiran!")

		case 7:
			fmt.Println("RANGE_ITERATE nije implementiran!")

		case 8:
			fmt.Println("SNAPSHOT nije implementiran!")

		case 9:
			fmt.Println("CHECKPOINT nije implementiran!")

		case 10:
			fmt.Println("VALIDACIJA_MERKLE_STABLA nije implementirana!")

		default:
			fmt.Println("GRESKA NEPOZNATA KOMANDA")
		}
	}
}

func readLine(prompt string) string {
	fmt.Print(prompt)
	var text string
	fmt.Scanln(&text)
	return strings.TrimSpace(text)
}

func printRecords(records *[]record.Record) {
	if len(*records) == 0 {
		fmt.Println("Nema rezultata.")
		return
	}
	for _, rec := range *records {
		fmt.Printf("Key: %s, Value: %s\n", rec.Key, string(rec.Value))
	}
}
