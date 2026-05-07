package main

import (
	"fmt"
	"strings"

	eng "github.com/Avram-2005/PROJEKAT_NASP/Engine"
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
			fmt.Println("OK")

		case 2:
			key := readLine("Unesite key za brisanje: ")

			err := engine.Delete(key)
			if err != nil {
				fmt.Println("Greska pri brisanju:", err)
				continue
			}
			fmt.Println("OK")

		case 3:
			key := readLine("Unesite key za pretragu: ")

			value, err := engine.Get(key)
			if err != nil {
				fmt.Println("Greska pri citanju:", err)
				continue
			}
			fmt.Println(string(value))

		case 4:
			Return(PrefixScan())

		case 5:
			Return(RangeScan())

		case 6:
			Return(PrefixIterate())

		case 7:
			Return(RangeIterate())

		case 8:
			Return(Snapshot())

		case 9:
			Return(Checkpoint())

		case 10:
			Return(ValidateMerkleTree())

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

func LoadConfig()                     {}
func InitializeSystem()               {}
func RecoverSystem()                  {}
func ShutdownSystem()                 {}
func IsTokenBucketEnabled() bool      { return false }
func IsTokenBucketAllowed(i int) bool { return true }

func Put() {
	if IsMemtableFull() {
		Flush()
	}
}

func Delete() {
	if IsMemtableFull() {
		Flush()
	}
}

func IsMemtableFull() bool { return false }
func Flush()               {}

func SearchCache() string    { return "" }
func SearchMemtable() string { return "" }
func SearchSSTable() string  { return "" }

func PrefixScan() string    { return "" }
func RangeScan() string     { return "" }
func PrefixIterate() string { return "" }
func RangeIterate() string  { return "" }

func Snapshot() string           { return "" }
func Checkpoint() string         { return "" }
func ValidateMerkleTree() string { return "" }

func Return(string) { /* Prikazuje rezultat korisniku */ }
