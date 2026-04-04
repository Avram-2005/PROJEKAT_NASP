package checkpoint

import (
	"os"
	"strings"
)

// Funkcija stvara hard link na specifican fajl unutar poddirektorijuma
// checkpoints direktorijuma nase aplikacije, koji skladisti sve nase checkpoint-e
// filename-ime fajla ciji hard link kreiramo
// checkpointdirectory-ime poddirektorijuma u kojem stvaramo hard link
func CreateHardLink(fileName string, checkpointDirectory string) error {
	//odvajamo poslednji deo naziva fajla koji cuvamo
	fileNameParts := strings.Split(fileName, "/")
	lastPart := fileNameParts[len(fileNameParts)-1]
	//linkujemo ga unutar odabranog direktorijuma unutar checkpoint
	newFileName := "checkpoints/" + checkpointDirectory + "/" + lastPart
	err := os.Link(fileName, newFileName)
	if err != nil {
		return err
	}
	return nil
}

// Stvaramo checkpoint odredjenog celokupnog(ocekivano je da direktorijum sadrzi sstable-ove)
// targetdirectory-direktorijum za koji stvaramo checkpoint
func CreateCheckpoint(targetDirectory string) error {
	//dobavljamo sve fajlove direktorijuma
	files, err := os.ReadDir(targetDirectory)
	if err != nil {
		return err
	}
	err = os.Mkdir("checkpoints/"+targetDirectory, 0755)
	if err != nil {
		return err
	}
	for _, file := range files {
		CreateHardLink(targetDirectory+"/"+file.Name(), targetDirectory)
	}

	return nil
}
