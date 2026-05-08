package engine

import (
	"os"
	"strings"

	blockmanager "github.com/Avram-2005/PROJEKAT_NASP/BlockManager"
	cache "github.com/Avram-2005/PROJEKAT_NASP/Cache"
	configuration "github.com/Avram-2005/PROJEKAT_NASP/Config"
	lsm "github.com/Avram-2005/PROJEKAT_NASP/LSM"
	memtable "github.com/Avram-2005/PROJEKAT_NASP/Memtable"
	record "github.com/Avram-2005/PROJEKAT_NASP/Record"
	wal "github.com/Avram-2005/PROJEKAT_NASP/WAL"
)

type Engine struct {
	configuration *configuration.Config
	cache         *cache.Cache
	memtable      *memtable.MemtableManager
	lsmTree       *lsm.LSM
	writeAheadLog *wal.WAL
	blockManager  *blockmanager.BlockManager
	//TODO: add token bucket after tokenbucket merge
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
	}, nil
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
		return *value, nil
	}
	//ako nije pronadjeno u cache-u trazimo u memtableu
	foundRecord, ok, err := engine.memtable.GetRecord(key)
	if err != nil {
		return nil, err
	}
	if ok && !foundRecord.Tombstone {
		retVal := foundRecord.Value
		//dodajemo pronadjenu vrednost u cache
		engine.cache.Put(key, &retVal)
		return retVal, nil
	}
	//trazimo u lsm stablu vrednost ako nije pronadjena
	sstValue, err := engine.lsmTree.Get(key)
	if err != nil {
		return nil, err
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

func (engine *Engine) PrefixScan(prefix string) *[]record.Record {
	return nil
}

func (engine *Engine) RangeScan(startKey, endKey string) *[]record.Record {
	return nil
}
