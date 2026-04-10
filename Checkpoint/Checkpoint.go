package checkpoint

import (
	"container/list"
	"errors"
	"fmt"
	"os"
	"strings"
)

type CheckpointManager struct {
	checkpointMap map[string]*Checkpoint
}

type Checkpoint struct {
	checkpointName string
}

func NewCheckpointManager() (*CheckpointManager, error) {
	directoryMap := make(map[string]*Checkpoint)
	files, err := os.ReadDir("checkpoints")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		Checkpoint, err := NewCheckpoint(file.Name())
		if err != nil {
			return nil, err
		}
		directoryMap[file.Name()] = Checkpoint
	}
	return &CheckpointManager{
		checkpointMap: directoryMap,
	}, nil
}

// funkcija stvara nov checkpoint unutar checkpoints folder-a i automatski ga dodaje u manager
func (ch *CheckpointManager) AddCheckpoint(targetDirectory string, checkPointDirectory string) error {
	_, ok := ch.checkpointMap[checkPointDirectory]
	if ok {
		return fmt.Errorf("that checkpoint already exists")
	}
	checkpoint, err := CreateCheckpoint(targetDirectory, checkPointDirectory)
	if err != nil {
		return err
	}
	ch.checkpointMap[checkPointDirectory] = checkpoint
	return nil
}

// put koristi vec postojeci checkpoint za razliku od add, koji ga stvori za nas
// funkcija napravljena radi fleksibilnosti
func (ch *CheckpointManager) PutCheckpoint(checkpoint *Checkpoint) error {
	_, ok := ch.checkpointMap[checkpoint.checkpointName]
	if ok {
		return fmt.Errorf("that checkpoint already exists")
	}
	ch.checkpointMap[checkpoint.checkpointName] = checkpoint
	return nil
}

func (ch *CheckpointManager) GetCheckpoint(name string) (*Checkpoint, error) {
	checkpoint, ok := ch.checkpointMap[name]
	if !ok {
		return nil, fmt.Errorf("checkpoint not found in manager")
	}
	return checkpoint, nil
}

func (ch *CheckpointManager) DeleteCheckpoint(name string) error {
	checkpoint, ok := ch.checkpointMap[name]
	if !ok {
		return fmt.Errorf("checkpoint not found in manager")
	}
	checkpoint.Delete()
	delete(ch.checkpointMap, name)
	return nil
}

func NewCheckpoint(name string) (*Checkpoint, error) {
	info, err := os.Stat("checkpoints/" + name)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("the checkpoint must be of a directory")
	}
	return &Checkpoint{
		checkpointName: name,
	}, nil
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

// funkcija namenjena da olaksa internu logiku checkpoint-a
// stvara listu svih fajlova koji moraju da se rekreiraju u checkpoint-u
func generateFileList(targetDirectory string) (*list.List, error) {
	fileList := list.New()
	files, err := os.ReadDir(targetDirectory)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			// rekurzivni poziv u slucaju da je fajl na koji smo naisli direktorijum
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
	checkpoint, err := NewCheckpoint(checkPointDirectory)
	if err != nil {
		return nil, err
	}
	return checkpoint, nil
}

// funkcija namenjeno da brise specificno direktorijum unutar checkpoints folder-a
func DeleteCheckpointDirectory(targetCheckpoint string) error {
	err := os.RemoveAll("checkpoints/" + targetCheckpoint)
	return err
}

// funkcija se poziva nad checkpoint objektom, i brise sadrzaj tog checkpoint-a
func (ch *Checkpoint) Delete() error {
	err := DeleteCheckpointDirectory(ch.checkpointName)
	return err
}

// otvaramo fajl unutar checkpoint-a u modu za citanje
func (ch *Checkpoint) OpenFileRead(fileName string) (*os.File, error) {
	file, err := os.OpenFile("checkpoints/"+ch.checkpointName+"/"+fileName, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// otvaramo fajl unutar checkpoint-a u modu za citanje i pisanje
func (ch *Checkpoint) OpenFileReadWrite(fileName string) (*os.File, error) {
	file, err := os.OpenFile("checkpoints/"+ch.checkpointName+"/"+fileName, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}
