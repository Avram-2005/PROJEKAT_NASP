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
		fmt.Println("0  - UGASI SISTEM")
		fmt.Println("1  - PUT")
		fmt.Println("2  - DELETE")
		fmt.Println("3  - GET")
		fmt.Println("4  - PREFIX_SCAN")
		fmt.Println("5  - RANGE_SCAN")
		fmt.Println("6  - PREFIX_ITERATE")
		fmt.Println("7  - RANGE_ITERATE")
		fmt.Println("8  - SNAPSHOT")
		fmt.Println("9  - CHECKPOINT")
		fmt.Println("10 - VALIDACIJA_MERKLE_STABLA")
		fmt.Println("----------------------------------------------")

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
			key := readLine("Unesite kljuc: ")
			value := readLine("Unesite vrednost: ")

			if key == "" || value == "" {
				fmt.Println("Kljuc i vrednost ne smeju biti prazni.")
				continue
			}

			err := engine.Put(key, []byte(value))
			if err != nil {
				fmt.Println("Greska pri upisu:", err)
				continue
			}

		case 2:
			key := readLine("Unesite kljuc za brisanje: ")
			if key == "" {
				fmt.Println("Kljuc ne sme biti prazan.")
				continue
			}

			err := engine.Delete(key)
			if err != nil {
				fmt.Println("Greska pri brisanju:", err)
				continue
			}

		case 3:
			key := readLine("Unesite kljuc za pretragu: ")
			if key == "" {
				fmt.Println("Kljuc ne sme biti prazan.")
				continue
			}

			value, err := engine.Get(key)
			if err != nil {
				fmt.Println("Greska pri citanju: " + err.Error())
				continue
			}
			if len(value) == 0 {
				fmt.Println("Nije pronadjena vrednost.")
				continue
			}

			fmt.Println("Vrednost:", string(value))

		case 4:
			prefix := readLine("Unesite prefix: ")
			if prefix == "" {
				fmt.Println("Prefix ne sme biti prazan.")
				continue
			}

			records := engine.PrefixScan(prefix)
			printRecords(records)

		case 5:
			startKey := readLine("Unesite pocetni key: ")
			endKey := readLine("Unesite krajnji key: ")

			if startKey == "" || endKey == "" {
				fmt.Println("Pocetni i krajnji kljucevi ne smeju biti prazni.")
				continue
			}
			if startKey > endKey {
				fmt.Println("Pocetni kljuc ne sme biti veci od krajnjeg kljuca.")
				continue
			}

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
	if records == nil || len(*records) == 0 {
		fmt.Println("Nema rezultata.")
		return
	}

	fmt.Println("Rezultati:")
	fmt.Println("----------------------------------------------")
	for i, r := range *records {
		fmt.Printf("%d. Kljuc: %s, Vrednost: %s\n", i+1, r.Key, string(r.Value))
	}
	fmt.Println("----------------------------------------------")
}
