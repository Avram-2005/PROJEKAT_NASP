package wal

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	BlockManager "github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
	Record "github.com/Avram-2005/PROJEKAT_NASP/Record"
)

type WAL struct {
	filePath              string                     //Putanja do foldera gde se čuvaju WAL fajlovi
	segmentList           []string                   //Lista putanja do svih WAL segmenata(fajlova)
	writeFile             *os.File                   //Trenutni fajl u koji upisujemo
	currentWritePosition  int                        //Pozicija (bajt) na kojoj se trenutno nalazimo u fajlu
	segmentSize           int                        //Maksimalna dozvoljena veličina jednog fajla
	blockManager          *BlockManager.BlockManager //BlockManager za pisanje i citanje u blokovima
	lowWatermarks         []string                   //Segmenti koji se prate za eventualno brisanje
	memtableRotationCount int                        //Brojač rotacija memtable-a koji se koristi za praćenje kada treba obrisati stare WAL segmente
}

const (
	SEGMENT_NAME   = "wal_"
	FILE_PATH      = "./WAL/walDATA"
	FILE_EXTENSION = ".wal"

	//Tipovi chunkova koji se koriste kada zapis ne staje u jedan blok i mora da se iseče na delove
	ChunkTypeZero   byte = 0
	ChunkTypeFull   byte = 1
	ChunkTypeFirst  byte = 2
	ChunkTypeMiddle byte = 3
	ChunkTypeLast   byte = 4
)

// CreatNewWAL pronalazi postojeće fajlove ili pravi nove, i vraća spremnu WAL strukturu
func CreatNewWAL(sizeSegment int, blocksize int, filePath string, memtableRotationCount int) (*WAL, error) {
	sizeSegment = sizeSegment * blocksize
	if sizeSegment < 64 {
		return nil, fmt.Errorf("size of segment must be at least 64KB")
	}

	if memtableRotationCount <= 0 {
		return nil, fmt.Errorf("number of memtable rotations before deleting old WAL segments must be a positive number")
	}

	//Pravimo folder za WAL ukoliko već ne postoji
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return nil, fmt.Errorf("critical error: %v", err)
	}

	//Skeniramo folder i skupljamo sve postojeće .wal fajlove u segmentList
	entries, _ := os.ReadDir(filePath)
	segments := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), SEGMENT_NAME) {
			segments = append(segments, filepath.Join(filePath, entry.Name()))
		}
	}

	//Ako je folder bio prazan, kreiramo prvi početni fajl (wal_0000.wal)
	if len(segments) == 0 {
		first := filepath.Join(filePath, fmt.Sprintf("%s0000%s", SEGMENT_NAME, FILE_EXTENSION))
		f, err := os.OpenFile(first, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		f.Close()
		segments = append(segments, first)
	}

	writePath := segments[len(segments)-1]
	writeFile, err := os.OpenFile(writePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	pos, err := writeFile.Seek(0, io.SeekEnd)
	if err != nil {
		writeFile.Close()
		return nil, err
	}

	walInstance := &WAL{
		filePath:              filePath,
		segmentList:           segments,
		writeFile:             writeFile,
		currentWritePosition:  int(pos),
		segmentSize:           sizeSegment * 1024,
		blockManager:          nil,
		memtableRotationCount: memtableRotationCount,
	}

	return walInstance, nil
}

func (wal *WAL) SetBlockManager(blockManager *BlockManager.BlockManager) error {
	if blockManager == nil {
		return fmt.Errorf("BlockManager must not be nil")
	}
	wal.blockManager = blockManager

	//proveravamo da li je segmentSize promenjen od poslednjeg pokretanja, i ako jeste, refaktorisemo fajlove da se uklope u novu veličinu
	err := wal.refactor()
	if err != nil {
		return err
	}
	return nil
}

// AddRecord pakuje ključ i vrednost u novi Record i šalje ga na upis
func (wal *WAL) AddRecord(key string, value []byte) (*Record.Record, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("key is empty")
	}

	rec, err := Record.NewRecord(key, value, false, time.Now())
	if err != nil {
		return nil, err
	}
	return rec, wal.appendRecord(rec)
}

// DeleteRecord kreira Record sa Tombstone markerom koji označava da je ključ obrisan, i šalje ga na upis
func (wal *WAL) DeleteRecord(key string) (*Record.Record, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("key is empty")
	}

	rec, err := Record.NewRecord(key, nil, true, time.Now())
	if err != nil {
		return nil, err
	}
	return rec, wal.appendRecord(rec)
}

// appendRecord pakuje Record u jedan ili više chunkova i upisuje ih u fajl
func (wal *WAL) appendRecord(newRecord *Record.Record) error {
	data := newRecord.Serialize()
	left := len(data)
	offset := 0
	isFirst := true
	blockSize := wal.blockManager.GetBlockSize()

	for left > 0 {
		if wal.currentWritePosition >= wal.segmentSize {
			if err := wal.rotateSegment(); err != nil {
				return err
			}
		}

		//Računamo koliko je slobodnog mesta ostalo u bloku
		blockOffset := wal.currentWritePosition % blockSize
		remainingInBlock := blockSize - blockOffset

		//Svaki chunk mora imati 1 bajt za tip i barem 20 bajtova za podatke.
		//Ako nema toliko mesta, ostatak bloka punimo nulama i idemo na sledeći blok
		if remainingInBlock <= 22 {
			padding := make([]byte, remainingInBlock)
			wal.blockManager.PutSpecific(wal.writeFile, wal.currentWritePosition/blockSize, blockOffset, remainingInBlock, &padding)
			wal.currentWritePosition += remainingInBlock
			continue
		}

		payloadSize := remainingInBlock - 1
		if left < payloadSize {
			payloadSize = left
		}

		var chunkType byte
		if isFirst {
			if left == payloadSize {
				chunkType = ChunkTypeFull //Ceo zapis je stao iz prve
			} else {
				chunkType = ChunkTypeFirst //Zapis kreće, ali će morati da se nastavi
			}
		} else {
			if left == payloadSize {
				chunkType = ChunkTypeLast //Ovo je poslednji deo zapisa
			} else {
				chunkType = ChunkTypeMiddle //Zapis se i dalje nastavlja
			}
		}

		chunk := make([]byte, 1+payloadSize)
		chunk[0] = chunkType
		copy(chunk[1:], data[offset:offset+payloadSize])

		err := wal.blockManager.PutSpecific(wal.writeFile, wal.currentWritePosition/blockSize, blockOffset, len(chunk), &chunk)
		if err != nil {
			return err
		}

		wal.currentWritePosition += len(chunk)
		offset += payloadSize
		left -= payloadSize
		isFirst = false
	}

	//Teranje operativnog sistema da fizički sačuva promene na disku
	return wal.writeFile.Sync()
}

// rotateSegment zatvara trenutni fajl i pravi novi
func (wal *WAL) rotateSegment() error {
	wal.writeFile.Close()
	newIndex := len(wal.segmentList)
	newPath := filepath.Join(wal.filePath, fmt.Sprintf("%s%04d%s", SEGMENT_NAME, newIndex, FILE_EXTENSION))

	file, err := os.OpenFile(newPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	wal.segmentList = append(wal.segmentList, newPath)
	wal.writeFile = file
	wal.currentWritePosition = 0
	return nil
}

// memtableRotation se poziva spolja kada se Memtable napuni
// Belezimo segmente koji su aktivni tokom memtable rotacije, jer su oni potencijalno potrebni za recovery
func (wal *WAL) memtableRotation() {
	wal.lowWatermarks = append(wal.lowWatermarks, wal.segmentList[len(wal.segmentList)-1])
	if len(wal.lowWatermarks) >= wal.memtableRotationCount {
		wal.FlushWAL()
	}
}

// FlushWAL fizički briše stare WAL segmente sa diska koji više nisu potrebni
func (wal *WAL) FlushWAL() error {
	if len(wal.lowWatermarks) == 0 {
		return nil
	}

	keepIndex := -1
	targetPath := wal.lowWatermarks[0]
	for i, path := range wal.segmentList {
		if path == targetPath {
			keepIndex = i
			break
		}
	}

	if keepIndex < 0 {
		return nil
	}

	//Brišemo sve fajlove iz liste pre indeksa koji želimo da zadržimo
	for i := 0; i < keepIndex; i++ {
		pathToDelete := wal.segmentList[i]

		if wal.writeFile != nil && wal.writeFile.Name() == pathToDelete {
			continue
		}

		err := os.Remove(pathToDelete)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete WAL segment %s: %v", pathToDelete, err)
		}
	}

	wal.segmentList = wal.segmentList[keepIndex:]
	wal.lowWatermarks = wal.lowWatermarks[1:]

	return nil
}

// Recovery prolazi kroz sve fajlove po redu i oživljava Memtable
func (wal *WAL) Recovery(mm *memtable.MemtableManager, lastSSTableTimestamp time.Time) error {
	var recordBuffer []byte
	for _, path := range wal.segmentList {
		if err := wal.recoverSingleSegment(path, mm, &recordBuffer, lastSSTableTimestamp); err != nil {
			return err
		}
	}
	return nil
}

// recoverSingleSegment čita jedan specifičan WAL fajl i obnavlja zapise
func (wal *WAL) recoverSingleSegment(path string, mm *memtable.MemtableManager, buffer *[]byte, lastSSTableTimestamp time.Time) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	//Čitamo blok po blok dok ne dođemo do kraja fajla
	for blockNum := 0; ; blockNum++ {
		block, err := wal.blockManager.Get(file, blockNum)
		if err != nil || block == nil || len(*block) == 0 {
			break
		}

		data := *block
		for offset := 0; offset < len(data); {
			if offset >= len(data) {
				break
			}

			chunkType := data[offset]
			//Ako je tip nula, ostatak bloka je prazan (padding), idemo na sledeći blok
			if chunkType == ChunkTypeZero {
				break
			}

			remainingInBlock := len(data) - offset

			switch chunkType {
			case ChunkTypeFull:
				//Zapis staje u jednom delu. Čitamo ga, prevodimo iz bajtova nazad u strukturu i dodajemo u Memtable
				if remainingInBlock < 1+Record.HEADER_SIZE {
					offset = len(data)
					continue
				}
				rec, totalSize, err := Record.DeserializeRecord(data[offset+1:])
				if err == nil {
					if rec.Timestamp.After(lastSSTableTimestamp) {
						mm.PutRecord(rec)
					}
					offset += 1 + totalSize
				} else {
					offset = len(data)
				}
				*buffer = (*buffer)[:0]

			case ChunkTypeFirst:
				//Zapis je isečen na više delova. Ovde je početak, pa ga čuvamo u bafer dok ne nađemo ostatak
				payload := data[offset+1:]
				*buffer = append((*buffer)[:0], payload...)
				offset += 1 + len(payload)

			case ChunkTypeMiddle:
				//Središnji deo isečenog zapisa se samo nastavlja na već postojeći bafer
				payload := data[offset+1:]
				if len(*buffer) > 0 {
					*buffer = append(*buffer, payload...)
				}
				offset += 1 + len(payload)

			case ChunkTypeLast:
				//Poslednji deo zapisa. Dodajemo ga u bafer i onda pokušavamo da rekonstrušemo ceo zapis
				if len(*buffer) < Record.HEADER_SIZE {
					*buffer = (*buffer)[:0]
					offset = len(data)
					continue
				}

				kSize := int(binary.BigEndian.Uint32((*buffer)[13:17]))
				vSize := int(binary.BigEndian.Uint32((*buffer)[17:21]))
				expectedTotal := Record.HEADER_SIZE + kSize + vSize
				missingBytes := expectedTotal - len(*buffer)

				if remainingInBlock < 1+missingBytes {
					*buffer = (*buffer)[:0]
					offset = len(data)
					continue
				}

				//Dodajemo poslednji deo, prevodimo ceo bafer u Record i ubacujemo u Memtable
				*buffer = append(*buffer, data[offset+1:offset+1+missingBytes]...)
				rec, _, err := Record.DeserializeRecord(*buffer)
				if err == nil {
					if rec.Timestamp.After(lastSSTableTimestamp) {
						mm.PutRecord(rec)
					}
				}

				*buffer = (*buffer)[:0]
				offset += 1 + missingBytes

			default:
				offset = len(data)
			}
		}
	}
	return nil
}

// refactor reorganizuje .wal fajlove ako se promeni velicina segmenta
func (wal *WAL) refactor() error {
	if len(wal.segmentList) <= 1 {
		return nil
	}

	firstFileInfo, err := os.Stat(wal.segmentList[0])
	if err != nil {
		return fmt.Errorf("failed to read first segment: %v", err)
	}

	if firstFileInfo.Size() == int64(wal.segmentSize) {
		return nil
	}

	if wal.writeFile != nil {
		wal.writeFile.Close()
	}

	var backupFiles []string
	for _, stariFajl := range wal.segmentList {
		binPath := stariFajl + ".bin"
		if err := os.Rename(stariFajl, binPath); err != nil {
			return fmt.Errorf("failed to create backup: %v", err)
		}
		backupFiles = append(backupFiles, binPath)
	}

	wal.segmentList = make([]string, 0)
	wal.currentWritePosition = 0

	firstPath := filepath.Join(wal.filePath, fmt.Sprintf("%s0000%s", SEGMENT_NAME, FILE_EXTENSION))
	newFile, err := os.OpenFile(firstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create first refactored file: %v", err)
	}
	wal.segmentList = append(wal.segmentList, firstPath)
	wal.writeFile = newFile
	blockSize := wal.blockManager.GetBlockSize()

	for _, binPath := range backupFiles {
		binFile, err := os.Open(binPath)
		if err != nil {
			return fmt.Errorf("failed to open backup file %s: %v", binPath, err)
		}

		for blockNum := 0; ; blockNum++ {
			block, err := wal.blockManager.Get(binFile, blockNum)
			if err != nil || block == nil || len(*block) == 0 {
				break
			}

			if wal.currentWritePosition >= wal.segmentSize {
				if err := wal.rotateSegment(); err != nil {
					binFile.Close()
					return fmt.Errorf("failed to rotate segment during refactoring: %v", err)
				}
			}

			err = wal.blockManager.PutSpecific(wal.writeFile, wal.currentWritePosition/blockSize, 0, len(*block), block)
			if err != nil {
				binFile.Close()
				return fmt.Errorf("failed to rewrite block: %v", err)
			}

			wal.currentWritePosition += len(*block)
		}
		binFile.Close()
	}

	wal.writeFile.Sync()

	for _, binPath := range backupFiles {
		os.Remove(binPath)
	}

	return nil
}

func (wal *WAL) Close() {
	if wal.writeFile != nil {
		wal.writeFile.Close()
	}
}
