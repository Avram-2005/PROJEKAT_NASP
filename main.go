package main

import (
	"fmt"
	"strconv"
	"strings"

	checkpoint "github.com/Avram-2005/PROJEKAT_NASP/Checkpoint"
	eng "github.com/Avram-2005/PROJEKAT_NASP/Engine"
	engine "github.com/Avram-2005/PROJEKAT_NASP/Engine"
	scan "github.com/Avram-2005/PROJEKAT_NASP/Scan"
	tokenbucket "github.com/Avram-2005/PROJEKAT_NASP/TokenBucket"
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
		fmt.Println("Error during system initialization:", err)
		return
	}
	engine.CheckTokenBucketInsert()

	for {
		fmt.Println()
		fmt.Print("Enter command: ")
		fmt.Println("0  - SHUT DOWN SYSTEM")
		fmt.Println("1  - PUT")
		fmt.Println("2  - DELETE")
		fmt.Println("3  - GET")
		fmt.Println("4  - PREFIX_SCAN")
		fmt.Println("5  - RANGE_SCAN")
		fmt.Println("6  - PREFIX_ITERATE")
		fmt.Println("7  - RANGE_ITERATE")
		fmt.Println("8  - SNAPSHOT")
		fmt.Println("9  - CHECKPOINT")
		fmt.Println("10 - MERKLE TREE VALIDATION")
		fmt.Println("----------------------------------------------")

		var command int
		_, err := fmt.Scanln(&command)
		if err != nil {
			fmt.Println("Invalid command input")
			continue
		}

		if command == 0 {
			engine.ShutDown()
			fmt.Println("System is shut down.")
			break
		}

		ok, err := checkTokens(engine)
		if err != nil {
			break
		}
		if !ok {
			continue
		}

		switch command {
		case 1:
			key := readLine("Enter key: ")
			value := readLine("Enter value: ")

			if key == "" || value == "" {
				fmt.Println("Key and value must not be empty.")
				continue
			}

			//User cannot put to tokenbucket key
			if key == tokenbucket.INTERNAL_KEY {
				fmt.Println("Key reserved for token bucket.")
				continue
			}

			err := engine.Put(key, []byte(value))
			if err != nil {
				fmt.Println("Error during write:", err)
				continue
			}

		case 2:
			key := readLine("Enter key for deletion: ")
			if key == "" {
				fmt.Println("Key must not be empty.")
				continue
			}

			//User cannot delete tokenbucket key
			if key == tokenbucket.INTERNAL_KEY {
				fmt.Println("Key reserved for token bucket.")
				continue
			}

			err := engine.Delete(key)
			if err != nil {
				fmt.Println("Error during deletion:", err)
				continue
			}

		case 3:
			key := readLine("Unesite kljuc za pretragu: ")
			if key == "" {
				fmt.Println("Key must not be empty.")
				continue
			}

			//User cannot get tokenbucket key
			if key == tokenbucket.INTERNAL_KEY {
				fmt.Println("Key reserved for token bucket.")
				continue
			}

			value, err := engine.Get(key)
			if err != nil {
				fmt.Println("Error during reading: " + err.Error())
				continue
			}
			if len(value) == 0 {
				fmt.Println("Value not found.")
				continue
			}

			fmt.Println("Value:", string(value))

		case 4:
			prefix := readLine("Enter prefix: ")
			if prefix == "" {
				fmt.Println("Prefix must not be empty.")
				continue
			}
			if strings.HasPrefix(tokenbucket.INTERNAL_KEY, prefix) {
				fmt.Println("Cannot scan token bucket internal key.")
				continue
			}
			pageNumber, pageSize := readPage()
			result, err := engine.PrefixScan(prefix, pageNumber, pageSize)
			if err != nil {
				fmt.Printf("Prefix scan error: %v\n", err)
				continue
			}
			printScanResult(result)

		case 5:
			startKey := readLine("Enter start key: ")
			endKey := readLine("Enter end key: ")

			if startKey == "" || endKey == "" {
				fmt.Println("Start and end keys must not be empty.")
				continue
			}
			if startKey > endKey {
				fmt.Println("Start key must not be greater than end key.")
				continue
			}

			pageNumber, pageSize := readPage()
			result, err := engine.RangeScan(startKey, endKey, pageNumber, pageSize)
			if err != nil {
				fmt.Printf("Range scan error: %v\n", err)
				continue
			}
			printScanResult(result)

		case 6:
			prefix := readLine("Enter prefix: ")
			if prefix == "" {
				fmt.Println("Prefix must not be empty")
				continue
			}
			if strings.HasPrefix(tokenbucket.INTERNAL_KEY, prefix) {
				fmt.Println("Cannot iterate over token bucket internal key.")
				continue
			}
			iter, err := engine.NewPrefixIterator(prefix)
			if err != nil {
				fmt.Printf("Failed to create iterator: %v\n", err)
				continue
			}
			defer iter.Stop()
			fmt.Println("Iterating through records: ")
			fmt.Println("----------------------------------------------")
			count := 0
			for {
				fmt.Print("\nPress 'n' for next record, 's' to stop: ")
				var command string
				fmt.Scanln(&command)
				if command == "s" || command == "stop" {
					iter.Stop()
					fmt.Println("Iteration stopped.")
					break
				}
				if command == "n" || command == "next" {
					if !iter.Next() {
						fmt.Println("No more records.End of iteration.")
						break
					}
					count++
					fmt.Printf("%d. Key: %s, Value: %s\n", count, iter.Key(), string(iter.Value()))
				} else {
					fmt.Println("Unknown command. Use 'n' for next, 's' to stop.")
				}
			}
			if count == 0 {
				fmt.Println("No results found.")
			}
			fmt.Println("----------------------------------------------")

		case 7:
			fmt.Println("RANGE_ITERATE is not implemented!")

		case 8:
			fmt.Println("SNAPSHOT is not implemented!")

		case 9:
			checkPointMenu(engine)

		case 10:
			all := engine.GetAllSSTables()
			if len(all) == 0 {
				fmt.Println("There are no SSTables to validate.")
				continue
			}

			for i, info := range all {
				fmt.Printf("%d. Level: %d, Path: %s\n", i+1, info.Level, info.Path)
			}
			fmt.Println("Choose SSTable: ")
			var choice int
			_, err := fmt.Scanln(&choice)
			if err != nil || choice > len(all) || choice <= 0 {
				fmt.Print("Choice doesnt exist!")
				continue
			}
			selected := all[choice-1].Table

			valid, corrupted, err := engine.ValidateSSTable(selected)
			if err != nil {
				fmt.Printf("Validation error %v", err)
				continue
			}
			if valid {
				fmt.Print("SSTable is valid")
			} else {
				fmt.Printf("Found corrupted records: %d", len(corrupted))
				for _, rec := range corrupted {
					fmt.Printf("\nKey: %s, Value: %s, Timestamp: %s, Tombstone: %t", rec.Key, rec.Value, rec.Timestamp.String(), rec.Tombstone)
				}
			}

		default:
			fmt.Println("ERROR UNKNOWN COMMAND")
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
		fmt.Print("Enter command: ")
		fmt.Println("0  - GO BACK")
		fmt.Println("1  - CREATE SYSTEM CHECKPOINT")
		fmt.Println("2  - OPEN CHECKPOINT")

		var command int
		_, err := fmt.Scanln(&command)
		if err != nil {
			fmt.Println("Invalid command input")
			continue
		}

		if command == 0 {
			break
		}

		ok, err := checkTokens(engine)
		if err != nil {
			break
		}
		if !ok {
			continue
		}

		if command == 1 {
			fmt.Println()
			fmt.Print("Enter checkpoint name: ")
			var command string
			_, err := fmt.Scanln(&command)
			if err != nil {
				fmt.Println("Invalid command input")
				continue
			}
			checkPointManager.AddCheckpoint(engine.GetRoot(), command)
			fmt.Print("Checkpoint successfully added: ")
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
	fmt.Print("Enter the name of the checkpoint you want to open: ")
	var command string
	_, err := fmt.Scanln(&command)
	if err != nil {
		fmt.Println("Invalid command input")
		checkPointOpen(checkPointManager, engine)
	}
	isFound := false
	for elem := checkpointList.Front(); elem != nil; elem = elem.Next() {
		fmt.Print(elem.Value.(string))
		if elem.Value == command {
			checkpoint, err := checkPointManager.GetCheckpoint(elem.Value.(string))
			if err != nil {
				fmt.Println("Something went wrong")
				return
			}
			isFound = true
			openedCheckpoint(checkpoint, engine)
		}
	}
	if !isFound {
		fmt.Println("Non-existent checkpoint")
	}

}

func openedCheckpoint(ch *checkpoint.Checkpoint, originalEngine *eng.Engine) {
	configPath := "config/config.yaml"

	walPath := "checkpoints/" + ch.GetName() + "/walDATA"
	sstablePath := "checkpoints/" + ch.GetName() + "/sstable"

	engine, err := eng.NewEngine(configPath, walPath, sstablePath)
	if err != nil {
		fmt.Println("Error during system initialization:", err)
		return
	}
	originalEngine.CheckTokenBucketInsert()

	for {
		fmt.Println()
		fmt.Print("Enter command: ")
		fmt.Println("0  - SHUT DOWN SYSTEM")
		fmt.Println("1  - GET")
		fmt.Println("2  - PREFIX_SCAN")
		fmt.Println("3  - RANGE_SCAN")
		fmt.Println("4  - PREFIX_ITERATE")
		fmt.Println("5  - RANGE_ITERATE")
		fmt.Println("6 - MERKLE TREE VALIDATION")
		fmt.Println("7 - DELETE CHECKPOINT")
		fmt.Println("----------------------------------------------")

		var command int
		_, err := fmt.Scanln(&command)
		if err != nil {
			fmt.Println("Invalid command input")
			continue
		}

		if command == 0 {
			engine.ShutDown()
			fmt.Println("System is shut down.")
			break
		}

		ok, err := checkTokens(originalEngine)
		if err != nil {
			break
		}
		if !ok {
			continue
		}

		switch command {
		case 1:
			key := readLine("Enter key for search: ")
			if key == "" {
				fmt.Println("Key must not be empty.")
				continue
			}

			//User cannot get tokenbucket key
			if key == tokenbucket.INTERNAL_KEY {
				fmt.Println("Key reserved for token bucket.")
				continue
			}

			value, err := engine.Get(key)
			if err != nil {
				fmt.Println("Error during reading: " + err.Error())
				continue
			}
			if len(value) == 0 {
				fmt.Println("Value not found.")
				continue
			}

			fmt.Println("Value:", string(value))

		case 2:
			prefix := readLine("Enter prefix: ")
			if prefix == "" {
				fmt.Println("Prefix must not be empty.")
				continue
			}

			//records := engine.PrefixScan(prefix)
			//printRecords(records)

		case 3:
			startKey := readLine("Enter start key: ")
			endKey := readLine("Enter end key: ")

			if startKey == "" || endKey == "" {
				fmt.Println("Start and end keys must not be empty.")
				continue
			}
			if startKey > endKey {
				fmt.Println("Start key must not be greater than end key.")
				continue
			}

			//records := engine.RangeScan(startKey, endKey)
			//printRecords(records)

		case 4:
			//TODO: implement token bucket once functionality is implemented
			//tokenbucket.INTERNAL_KEY may not be accessed by prefix iterate
			fmt.Println("PREFIX_ITERATE is not implemented!")

		case 5:
			//TODO: implement token bucket once functionality is implemented
			//tokenbucket.INTERNAL_KEY may not be accessed by range iterate
			fmt.Println("RANGE_ITERATE is not implemented!")

		case 6:
			all := engine.GetAllSSTables()
			if len(all) == 0 {
				fmt.Println("There are no SSTables to validate.")
				continue
			}

			for i, info := range all {
				fmt.Printf("%d. Level: %d, Path: %s\n", i+1, info.Level, info.Path)
			}
			fmt.Println("Choose SSTable: ")
			var choice int
			_, err := fmt.Scanln(&choice)
			if err != nil || choice > len(all) || choice <= 0 {
				fmt.Print("Choice doesnt exist!")
				continue
			}
			selected := all[choice-1].Table

			valid, corrupted, err := engine.ValidateSSTable(selected)
			if err != nil {
				fmt.Printf("Validation error %v", err)
				continue
			}
			if valid {
				fmt.Print("SSTable is valid")
			} else {
				fmt.Printf("Found corrupted records: %d", len(corrupted))
				for _, rec := range corrupted {
					fmt.Printf("Key: %s, Value: %s, Timestamp: %s, Tombstone: %t", rec.Key, rec.Value, rec.Timestamp.String(), rec.Tombstone)
				}
			}

		case 7:
			engine.ShutDown()
			err := ch.Delete()
			if err != nil {
				fmt.Print(err)
			}
			fmt.Print("Checkpoint successfully deleted")
			return

		default:
			fmt.Println("ERROR UNKNOWN COMMAND")
		}
	}
}

func checkTokens(engine *eng.Engine) (bool, error) {
	tokenBucketBytes, err := engine.Get(tokenbucket.INTERNAL_KEY)
	if err != nil {
		fmt.Print(err)
		return false, err
	}
	tokenBucket, err := tokenbucket.Deserialize(tokenBucketBytes)
	if err != nil {
		fmt.Print(err)
		return false, err
	}
	allowed := tokenBucket.Allow()
	if !allowed {
		fmt.Println("You currently have no tokens")
		timeToRefill := tokenBucket.GetNextRefill()
		output := strconv.Itoa(int(timeToRefill)) + " seconds to next refill"
		fmt.Println(output)
		return false, nil
	}
	putBytes := tokenBucket.Serialize()
	err = engine.Put(tokenbucket.INTERNAL_KEY, putBytes)
	if err != nil {
		fmt.Print(err)
	}
	return true, nil
}

func readLine(prompt string) string {
	fmt.Print(prompt)
	var text string
	fmt.Scanln(&text)
	return strings.TrimSpace(text)
}

/*func printRecords(records *[]record.Record) {
	if records == nil || len(*records) == 0 {
		fmt.Println("No results.")
		return
	}

	fmt.Println("Results:")
	fmt.Println("----------------------------------------------")
	for i, r := range *records {
		fmt.Printf("%d. Key: %s, Value: %s\n", i+1, r.Key, string(r.Value))
	}
	fmt.Println("----------------------------------------------")
}*/

// helper for pagination
func readPage() (int, int) {
	fmt.Print("Enter page number: ")
	var pageNumber int
	fmt.Scanln(&pageNumber)
	if pageNumber <= 0 {
		pageNumber = 1
	}
	fmt.Print("Enter page size: ")
	var pageSize int
	fmt.Scanln(&pageSize)
	if pageSize <= 0 {
		pageSize = 10
	}
	return pageNumber, pageSize
}

// helper for scan results
func printScanResult(result *scan.ScanResult) {
	if result == nil || len(result.Records) == 0 {
		fmt.Println("No results found")
		return
	}
	fmt.Printf("Total records: %d\n", result.TotalCount)
	fmt.Printf("Page %d out of %d (page size: %d)\n", result.PageNumber, (result.TotalCount+result.PageSize-1)/result.PageSize, result.PageSize)
	fmt.Println("----------------------------------------------")

	for i, rec := range result.Records {
		fmt.Printf("%d. Key: %s, Value: %s\n", (result.PageNumber-1)*result.PageSize+i+1, rec.Key, string(rec.Value))
	}
	fmt.Println("----------------------------------------------")
	if result.HasMore {
		fmt.Println("More pages available. Use a bigger pageNumber than the current to see more.")
	}
}
