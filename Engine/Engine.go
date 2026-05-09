package engine

import (
	"fmt"
	"os"
	"strings"

	blockmanager "github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	cache "github.com/Avram-2005/PROJEKAT_NASP/Cache"
	configuration "github.com/Avram-2005/PROJEKAT_NASP/Config"
	lsm "github.com/Avram-2005/PROJEKAT_NASP/LSM"
	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
	scan "github.com/Avram-2005/PROJEKAT_NASP/Scan"
	tokenbucket "github.com/Avram-2005/PROJEKAT_NASP/TokenBucket"
	wal "github.com/Avram-2005/PROJEKAT_NASP/WAL"
)

type Engine struct {
	configuration *configuration.Config
	cache         *cache.Cache
	memtable      *memtable.MemtableManager
	lsmTree       *lsm.LSM
	writeAheadLog *wal.WAL
	blockManager  *blockmanager.BlockManager
	tokenBucket   *tokenbucket.TokenBucket
	scanner       *scan.SystemScanner
}

// TODO: incorporate token bucket after merge
func NewEngine(configPath string, walPath string, sstablePath string) (*Engine, error) {
	configuration := configuration.NewConfig()
	//kreiramo blockmanager koji samo cita sadrzaj config fajla
	temporaryBlockManager, err := blockmanager.NewBlockManager(2, 4)
	if err != nil {
		return nil, err
	}
	//otvaramo specificirani config fajl
	file, err := os.OpenFile(configPath, 0644, os.FileMode(os.O_RDONLY))
	if err != nil {
		return nil, err
	}
	//initialize cita config fajl, proverava validnost vrednosti i cuva sve sto je dobro
	//bilo koje nevalidne konfiguracije config zamenjuje za default
	//engine.configuration = *configuration
	configuration.Initialize(temporaryBlockManager, file)
	engineCache, err := configuration.InitializeCache()
	if err != nil {
		return nil, err
	}

	file.Close()

	engineTokenBucket, err := tokenbucket.NewTokenBucket(configuration.TokenBucketConfig.MaxNumTokens, int64(configuration.TokenBucketConfig.RefillTime))
	if err != nil {
		return nil, err
	}

	engineBlockManager, err := configuration.InitializeBlockManager()
	if err != nil {
		return nil, err
	}

	configuration.SetSSTableRoot(sstablePath)
	configuration.SetWALRoot(walPath)

	engineLSMTree, err := configuration.InitializeLSM(engineBlockManager)
	if err != nil {
		return nil, err
	}

	engineMemtable, err := configuration.InitializeMemtable(
		func(entries []*record.Record) error {
			return engineLSMTree.Flush(entries)
		})
	if err != nil {
		return nil, err
	}

	engineScanner := scan.NewSystemScanner(engineMemtable, engineLSMTree)

	engineWriteAheadLog, err := configuration.InitializeWAL()
	if err != nil {
		return nil, err
	}

	err = engineWriteAheadLog.Recovery(engineMemtable)
	if err != nil {
		return nil, err
	}

	return &Engine{
		configuration: configuration,
		cache:         engineCache,
		memtable:      engineMemtable,
		lsmTree:       engineLSMTree,
		writeAheadLog: engineWriteAheadLog,
		blockManager:  engineBlockManager,
		tokenBucket:   engineTokenBucket,
		scanner:       engineScanner,
	}, nil
}

func (engine *Engine) CheckTokenBucketInsert() (bool, error) {
	bucketBytes, err := engine.Get(tokenbucket.INTERNAL_KEY)
	if err != nil {
		fmt.Print("Error trying to get token bucket")
		newBucketBytes := engine.tokenBucket.Serialize()
		engine.Put(tokenbucket.INTERNAL_KEY, newBucketBytes)
		return false, err
	}
	//if bucket is empty report error and input a new one
	if len(bucketBytes) == 0 {
		//if no tokenbucket was found, insert a new tokenbucket
		newBucketBytes := engine.tokenBucket.Serialize()
		engine.Put(tokenbucket.INTERNAL_KEY, newBucketBytes)
		return false, nil
	}
	_, err = tokenbucket.Deserialize(bucketBytes)
	//if old bucket cannot be deserialized input a new one
	if err != nil {
		fmt.Print("Error deserializing current bucket")
		newBucketBytes := engine.tokenBucket.Serialize()
		engine.Put(tokenbucket.INTERNAL_KEY, newBucketBytes)
		return false, err
	}
	//if everything went right return true
	return true, nil
}

func (engine *Engine) GetRoot() string {
	sstableRoot := engine.configuration.SSTableConfig.TablesRoot
	split := strings.Split(sstableRoot, "/")
	return split[0]
}

func (engine *Engine) ShutDown() {
	engine.writeAheadLog.Close()
	engine.configuration = nil
	engine.blockManager = nil
	engine.memtable = nil
	engine.lsmTree = nil
	engine.writeAheadLog = nil
	engine.blockManager = nil
}

func (engine *Engine) WritePath(key string, value []byte) error {

	rec, err := engine.writeAheadLog.AddRecord(key, value)
	if err != nil {
		return err
	}
	err = engine.memtable.PutRecord(rec)
	if err != nil {
		return err
	}
	err = engine.cache.Put(key, &value)
	if err != nil {
		return err
	}
	return nil
}

func (engine *Engine) ReadPath(key string) ([]byte, error) {
	//proveravamo da li vrednost u cache-u
	value, err, ok := engine.cache.Get(key)
	if err != nil {
		return nil, err
	}
	if ok {
		//ako vidimo da je kljuc obrisan u cache-u, vracamo nil
		if value == nil {
			return nil, nil
		}
		return *value, nil
	}

	//ako nije pronadjeno u cache-u trazimo u memtableu
	foundRecord, ok, err := engine.memtable.GetRecord(key)
	if err != nil {
		return nil, err
	}
	if ok {
		//ako je record u memtable-u obrisan, oznacavamo da je obrisan i unutar cache-a
		//i vracamo nil
		if foundRecord.Tombstone {
			engine.cache.Delete(key)
			return nil, nil
		}
		//ako record nije tu, vracamo vrednost
		retVal := foundRecord.Value
		//dodajemo pronadjenu vrednost u cache
		engine.cache.Put(key, &retVal)
		return retVal, nil
	}
	//trazimo u lsm stablu vrednost ako nije pronadjena
	sstValue, err := engine.lsmTree.Get(key)
	if err != nil {
		return nil, nil
	}
	//vracamo vrednost iz record-a
	return sstValue, nil
}

func (engine *Engine) Put(key string, value []byte) error {
	return engine.WritePath(key, value)
}

func (engine *Engine) Get(key string) ([]byte, error) {
	return engine.ReadPath(key)
}

func (engine *Engine) Delete(key string) error {
	rec, err := engine.writeAheadLog.DeleteRecord(key)
	if err != nil {
		return err
	}
	err = engine.memtable.Delete(key, rec.Timestamp)
	if err != nil {
		return err
	}
	err = engine.cache.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (engine *Engine) PrefixScan(prefix string, pageNumber, pageSize int) (*scan.ScanResult, error) {
	return engine.scanner.PrefixScan(prefix, pageNumber, pageSize)
}

func (engine *Engine) RangeScan(startKey, endKey string, pageNumber, pageSize int) (*scan.ScanResult, error) {
	return engine.scanner.RangeScan(startKey, endKey, pageNumber, pageSize)
}

func (engine *Engine) NewPrefixIterator(prefix string) (*scan.SystemIterator, error) {
	return engine.scanner.NewSystemPrefixIterator(prefix)
}

func (engine *Engine) NewRangeIterator(startKey, endKey string) (*scan.SystemIterator, error) {
	return engine.scanner.NewSystemRangeIterator(startKey, endKey)
}

func (engine *Engine) GetAllSSTables() []lsm.SSTableInfo {
	return engine.lsmTree.GetAllSSTables()
}

func (engine *Engine) ValidateSSTable(sst *lsm.SSTable) (bool, []record.Record, error) {
	return engine.lsmTree.ValidateSSTable(sst)
}
