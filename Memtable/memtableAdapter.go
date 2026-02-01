package Memtable

/*import (
	"fmt"
	"github.com/Avram-2005/PROJEKAT_NASP/HashMap"
	"github.com/Avram-2005/PROJEKAT_NASP/SkipList"
	"github.com/Avram-2005/PROJEKAT_NASP/BPlusTree"
)

type MemtableAdapter struct{
	config MemtableConfig
	size int
	total int

	dataStructure interface{}
	structureType string

	getFunction func(key string) ([]byte,bool,error)
	putFunction func(key string, value []byte) error
	deleteFunction func(key string) (bool,error)
	clearFunction func()

}

func NewMemtableAdapter(config MemtableConfig) (*MemtableAdapter,error){
	adapter:=&MemtableAdapter{
		config: config,
		size:0,
		total:0,
	}
	//inicijalizacija na osnovu tipa
	switch config.Type{
	case "hashmap":
		hm:=HashMap.NewHashMap()
		adapter.dataStructure=hm
		adapter.structureType="hashmap"
		adapter.initHashMap(hm)
	case "skip_list":
		sl,err:=SkipList.NewSkipList(config.SkipListMaxHeight)
		if err!=nil{
			return nil,err
		}
		adapter.dataStructure=sl
		adapter.structureType="skip_list"
		adapter.initSkipList(sl)
	case "b_plus_tree":
		bpt,err:=BPlusTree.NewBPlusTree(config.BPlusTreeDegree)
		if err!=nil{
			return nil,err
		}
		adapter.dataStructre=bpt
		adapter.structureType="b_plus_tree"
		adapter.initBPlusTree(bpt)
	default:
		return nil,fmt.Errorf("Memtable type: %s, was not recognized",config.Type)
	}
	return adapter,nil
}*/

//TODO1: INIT IMPLEMENTACIJA ZA SVE TIPOVE + SCANS
//TODO2: Implementacija emmtable interfejsa, put,get,delete,size,total,isempty,clear,iterator,shoulFLush,isfull
