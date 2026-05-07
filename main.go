package main

import (
	"bufio"
	"fmt"
	"os"
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

	printHelp()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		command := strings.ToLower(parts[0])

		switch command {
		case "exit", "quit", "shutdown":
			engine.ShutDown()
			fmt.Println("Sistem je ugasen.")
			return

		case "help":
			printHelp()

		case "put":
			if len(parts) < 3 {
				fmt.Println("Greska: put <key> <value>")
				continue
			}
			key := parts[1]
			value := strings.Join(parts[2:], " ")

			err := engine.Put(key, []byte(value))
			if err != nil {
				fmt.Println("Greska pri upisu:", err)
				continue
			}
			fmt.Printf("OK - Upisana vrednost za kljuc '%s'\n", key)

		case "get":
			if len(parts) < 2 {
				fmt.Println("Greska: get <key>")
				continue
			}
			key := parts[1]

			value, err := engine.Get(key)
			if err != nil {
				fmt.Println("Greska pri citanju: " + err.Error())
				continue
			}
			if len(value) == 0 {
				fmt.Println("Nije pronadjena vrednost.")
				continue
			}

			fmt.Printf("'%s' => '%s'\n", key, string(value))

		case "delete", "del":
			if len(parts) < 2 {
				fmt.Println("Greska: delete <key>")
				continue
			}
			key := parts[1]

			err := engine.Delete(key)
			if err != nil {
				fmt.Println("Greska pri brisanju:", err)
				continue
			}
			fmt.Printf("OK - Obrisan kljuc '%s'\n", key)

		case "prefix_scan":
			if len(parts) < 2 {
				fmt.Println("Greska: prefix_scan <prefix>")
				continue
			}
			prefix := parts[1]

			records := engine.PrefixScan(prefix)
			printRecords(records)

		case "range_scan":
			if len(parts) < 3 {
				fmt.Println("Greska: range_scan <start> <end>")
				continue
			}
			startKey := parts[1]
			endKey := parts[2]

			if startKey > endKey {
				fmt.Println("Greska: pocetni kljuc ne sme biti veci od krajnjeg kljuca.")
				continue
			}

			records := engine.RangeScan(startKey, endKey)
			printRecords(records)

		case "prefix_iterate":
			fmt.Println("prefix_iterate nije implementiran!")

		case "range_iterate":
			fmt.Println("range_iterate nije implementiran!")

		case "snapshot":
			fmt.Println("snapshot nije implementiran!")

		case "checkpoint":
			fmt.Println("checkpoint nije implementiran!")

		case "validate":
			fmt.Println("validacija_merkle_stabla nije implementirana!")

		default:
			fmt.Printf("Nepoznata komanda: '%s'. Ukucajte 'help' za popis komandi.\n", command)
		}
	}
}

func printHelp() {
	fmt.Println()
	fmt.Println("Dostupne komande:")
	fmt.Println("  put <key> <value>          - Upiši vrednost sa ključem")
	fmt.Println("  get <key>                  - Pročitaj vrednost po ključu")
	fmt.Println("  delete <key>               - Obriši ključ")
	fmt.Println("  prefix_scan <prefix>       - Skenira sve ključeve sa datim prefiksom")
	fmt.Println("  range_scan <start> <end>   - Skenira sve ključeve u rasponu")
	fmt.Println("  help                       - Prikaži ovu poruku")
	fmt.Println("  exit                       - Gasi sistem i izlazi")
	fmt.Println()
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
