package checkpoint

import (
	"container/list"
	"errors"
	"os"
	"strings"
)

type Checkpoint struct {
	checkpointName string
}

func NewCheckpoint(name string) *Checkpoint {
	return &Checkpoint{
		checkpointName: name,
	}
}

// Funkcija stvara hard link na specifican fajl unutar poddirektorijuma
// checkpoints direktorijuma nase aplikacije, koji skladisti sve nase checkpoint-e
// filename-ime fajla ciji hard link kreiramo
// checkpointdirectory-ime poddirektorijuma u kojem stvaramo hard link
func CreateHardLink(fileName string, checkpointDirectory string) error {
	//odvajamo poslednji deo naziva fajla koji cuvamo
	fileNameParts := strings.Split(fileName, "/")
	currentParts := "checkpoints/" + checkpointDirectory + "/"
	for i := 1; i < len(fileNameParts)-1; i++ {
		currentParts += fileNameParts[i]
		_, err := os.Stat(currentParts)

		if errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(currentParts, 0755)
			if err != nil {
				return err
			}
		}
		currentParts += "/"
	}
	lastPart := fileNameParts[len(fileNameParts)-1]
	currentParts += lastPart
	//linkujemo ga unutar odabranog direktorijuma unutar checkpoint
	err := os.Link(fileName, currentParts)
	if err != nil {
		return err
	}
	return nil
}
func generateFileList(targetDirectory string) (*list.List, error) {
	fileList := list.New()
	files, err := os.ReadDir(targetDirectory)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			directoryFiles, err := generateFileList(targetDirectory + "/" + file.Name())
			if err != nil {
				return nil, err
			}
			for elem := directoryFiles.Front(); elem != nil; elem = elem.Next() {
				fileList.PushBack(elem.Value)
			}
		} else {
			fileList.PushBack(targetDirectory + "/" + file.Name())
		}
	}
	return fileList, nil
}

// Stvaramo checkpoint odredjenog celokupnog direktorijuma(ocekivano je da direktorijum sadrzi sstable-ove)
// targetdirectory-direktorijum za koji stvaramo checkpoint
func CreateCheckpoint(targetDirectory string, checkPointDirectory string) (*Checkpoint, error) {
	//dobavljamo sve fajlove direktorijuma
	/*files, err := os.ReadDir(targetDirectory)
	if err != nil {
		return nil, err
	}*/
	//stvaramo direktorijum unutar checkpoints-a koji ce skladistiti hard linkove
	err := os.Mkdir("checkpoints/"+checkPointDirectory, 0755)
	if err != nil {
		return nil, err
	}
	fileList, err := generateFileList(targetDirectory)
	if err != nil {
		return nil, err
	}
	for elem := fileList.Front(); elem != nil; elem = elem.Next() {
		err = CreateHardLink(elem.Value.(string), checkPointDirectory)
		if err != nil {
			return nil, err
		}
	}
	//iteriramo kroz fajlove
	/*for _, file := range files {
		if file.IsDir() {

		}
		CreateHardLink(targetDirectory+"/"+file.Name(), checkPointDirectory)
	}*/

	return NewCheckpoint(checkPointDirectory), nil
}

func DeleteCheckpoint(targetCheckpoint string) error {
	err := os.RemoveAll("checkpoints/" + targetCheckpoint)
	return err
}

func (ch *Checkpoint) Delete() error {
	err := DeleteCheckpoint(ch.checkpointName)
	return err
}

func (ch *Checkpoint) OpenFileRead(fileName string) (*os.File, error) {
	file, err := os.OpenFile("checkpoints/"+ch.checkpointName+"/"+fileName, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (ch *Checkpoint) OpenFileReadWrite(fileName string) (*os.File, error) {
	file, err := os.OpenFile("checkpoints/"+ch.checkpointName+"/"+fileName, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}
