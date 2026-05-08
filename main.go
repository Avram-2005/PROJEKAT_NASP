package main

import (
	"fmt"
	"strings"

	checkpoint "github.com/Avram-2005/PROJEKAT_NASP/Checkpoint"
	eng "github.com/Avram-2005/PROJEKAT_NASP/Engine"
	engine "github.com/Avram-2005/PROJEKAT_NASP/Engine"
	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

func main() {
	mainMenu()
}

func mainMenu() {
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
			checkPointMenu(engine)

		case 10:
			fmt.Println("VALIDACIJA_MERKLE_STABLA nije implementirana!")

		default:
			fmt.Println("GRESKA NEPOZNATA KOMANDA")
		}
	}
}

func checkPointMenu(engine *engine.Engine) {
	checkPointManager, err := checkpoint.NewCheckpointManager()
	if err != nil {
		fmt.Print(err)
		return
	}
	for {
		fmt.Println()
		fmt.Print("Unesite komandu: ")
		fmt.Println("0  - IDI NAZAD")
		fmt.Println("1  - STVORI CHECKPOINT SISTEMA")
		fmt.Println("2  - OTVORI CHECKPOINT")

		var command int
		_, err := fmt.Scanln(&command)
		if err != nil {
			fmt.Println("Neispravan unos komande")
			continue
		}

		if command == 0 {
			break
		}

		if command == 1 {
			fmt.Println()
			fmt.Print("Unesite ime checkpoint-a: ")
			var command string
			_, err := fmt.Scanln(&command)
			if err != nil {
				fmt.Println("Neispravan unos komande")
				continue
			}
			checkPointManager.AddCheckpoint(engine.GetRoot(), command)
			fmt.Print("Checkpoint uspesno dodat: ")
		}
		if command == 2 {
			checkPointOpen(checkPointManager, engine)
		}
	}
}

func checkPointOpen(checkPointManager *checkpoint.CheckpointManager, engine *eng.Engine) {
	checkpointList := checkPointManager.GetCheckpointList()
	for elem := checkpointList.Front(); elem != nil; elem = elem.Next() {
		fmt.Println(elem.Value.(string))
	}
	fmt.Println()
	fmt.Print("Unesite ime checkpoint-a koji zelite da otvorite: ")
	var command string
	_, err := fmt.Scanln(&command)
	if err != nil {
		fmt.Println("Neispravan unos komande")
		checkPointOpen(checkPointManager, engine)
	}
	isFound := false
	for elem := checkpointList.Front(); elem != nil; elem = elem.Next() {
		fmt.Print(elem.Value.(string))
		if elem.Value == command {
			checkpoint, err := checkPointManager.GetCheckpoint(elem.Value.(string))
			if err != nil {
				fmt.Println("Nesto je poslo po zlu")
				return
			}
			isFound = true
			openedCheckpoint(checkpoint)
		}
	}
	if !isFound {
		fmt.Println("Nepostojeci checkpoint")
	}

}

func openedCheckpoint(ch *checkpoint.Checkpoint) {
	configPath := "config/config.yaml"

	walPath := "checkpoints/" + ch.GetName() + "/walDATA"
	sstablePath := "checkpoints/" + ch.GetName() + "/sstable"

	engine, err := eng.NewEngine(configPath, walPath, sstablePath)
	if err != nil {
		fmt.Println("Greska pri inicijalizaciji sistema:", err)
		return
	}

	for {
		fmt.Println()
		fmt.Print("Unesite komandu: ")
		fmt.Println("0  - UGASI SISTEM")
		fmt.Println("1  - GET")
		fmt.Println("2  - PREFIX_SCAN")
		fmt.Println("3  - RANGE_SCAN")
		fmt.Println("4  - PREFIX_ITERATE")
		fmt.Println("5  - RANGE_ITERATE")
		fmt.Println("6 - VALIDACIJA_MERKLE_STABLA")
		fmt.Println("7 - OBRISI CHECKPOINT")
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

		case 2:
			prefix := readLine("Unesite prefix: ")
			if prefix == "" {
				fmt.Println("Prefix ne sme biti prazan.")
				continue
			}

			records := engine.PrefixScan(prefix)
			printRecords(records)

		case 3:
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

		case 4:
			fmt.Println("PREFIX_ITERATE nije implementiran!")

		case 5:
			fmt.Println("RANGE_ITERATE nije implementiran!")

		case 6:
			fmt.Println("VALIDACIJA_MERKLE_STABLA nije implementirana!")

		case 7:
			engine.ShutDown()
			err := ch.Delete()
			if err != nil {
				fmt.Print(err)
			}
			fmt.Print("Checkpoint uspesno obrisan")
			return

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
