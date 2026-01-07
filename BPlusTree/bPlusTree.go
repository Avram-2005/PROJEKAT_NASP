package BPlusTree

import "errors"

type TreeNode struct {
	isLeaf   bool //proverava da li je cvor list-true (za unutrasnji cvor daje false)
	keys     []string
	children []*TreeNode //pokazivaci na decu, samo za unutrasnje cvorove
	values   [][]byte
	next     *TreeNode //sledeci list
	parent   *TreeNode //roditelj

	//brojaci
	keyCount   int
	childCount int
	valueCount int
}

type BPlusTree struct {
	root *TreeNode //koren stabla
	b    int       //stepen stabla
	size int       //br elemenata stabla
}

// Pravljenje novog b+ stabla
func NewBPlusTree(b int) (*BPlusTree, error) {
	if b < 2 { //broj kljuceva mora biti veci ili jednak 2
		return nil, errors.New("Degree of a tree has to be larger or equal to 2")
	}
	root := &TreeNode{
		isLeaf:     true,
		keys:       make([]string, 2*b),
		values:     make([][]byte, 2*b),
		keyCount:   0,
		valueCount: 0,
		next:       nil,
		parent:     nil,
	}

	return &BPlusTree{
		root: root,
		b:    b,
		size: 0,
	}, nil

}

//Kreiranje novog cvora
func createNode(b int, isLeaf bool) *TreeNode {
	if isLeaf {
		return &TreeNode{
			isLeaf:     true,
			keys:       make([]string, 2*b),
			values:     make([][]byte, 2*b),
			keyCount:   0,
			valueCount: 0,
			next:       nil,
			parent:     nil,
		}
	} else {
		return &TreeNode{
			isLeaf:     false,
			keys:       make([]string, 2*b),
			children:   make([]*TreeNode, 2*b+1),
			keyCount:   0,
			childCount: 0,
			parent:     nil,
		}
	}
}

// Povratna vrednost-broj elemenata stabla
func (tree *BPlusTree) Size() int {
	return tree.size
}

// Povratana vrednost- bool
func (tree *BPlusTree) IsEmpty() bool {
	return tree.size == 0
}

// Trazi list u kom se nalazi kljuc
func (tree *BPlusTree) FindLeaf(key string) *TreeNode {
	current := tree.root //pretragu pocinjemo od korena
	for !current.isLeaf {
		i := 0
		//trazimo kuda idemo
		for i < current.keyCount && key >= current.keys[i] {
			i++
		}
		current = current.children[i] //idemo u odgovarajuce dete
	}
	return current
}

//Dodavanje elementa
func (tree *BPlusTree) Insert(key string, value []byte) error {
	if key == "" {
		return errors.New("Key cannot be empty")
	}
	if value == nil {
		return errors.New("Value cannot be nil")
	}
	leaf := tree.FindLeaf(key) //nalazimo list u koji treba da ubacimo element
	//provera da li kljuc vec postoji
	for i := 0; i < leaf.keyCount; i++ {
		if leaf.keys[i] == key {
			leaf.values[i] = value //kljuc postoji,samo update
			return nil
		}
	}

	index := 0 //brojac indeksa gde cemo upisati novi element
	for index < leaf.keyCount && key > leaf.keys[index] {
		index++
	}
	//svi kljucevi biraju pomereni desno kako bismo mogli da dodamo novi element (pod pomereni, mislim preko pokazivaca)
	for i := leaf.keyCount; i > index; i-- {
		leaf.keys[i] = leaf.keys[i-1]
		leaf.values[i] = leaf.values[i-1]
	}
	//dodavanje novog kljuca
	leaf.keys[index] = key
	leaf.values[index] = value

	leaf.keyCount++
	leaf.valueCount++
	tree.size++
	//provera da li je list prepunjen
	if leaf.keyCount > 2*tree.b-1 {
		tree.splitLeaf(leaf)
	}
	return nil
}

func (tree *BPlusTree) splitLeaf(leaf *TreeNode) {
	middle := leaf.keyCount / 2
	newLeaf := createNode(tree.b, true)
	//Polovinu kljuceva prebacujemo u novi list
	for i := middle; i < leaf.keyCount; i++ {
		newLeaf.keys[i-middle] = leaf.keys[i]
		newLeaf.values[i-middle] = leaf.values[i]

		leaf.keys[i] = ""
		leaf.values[i] = nil

	}
	newLeaf.keyCount = leaf.keyCount - middle
	newLeaf.valueCount = leaf.keyCount - middle
	leaf.keyCount = middle
	leaf.valueCount = middle

	//povezivanje pokazivaca
	newLeaf.next = leaf.next
	leaf.next = newLeaf
	newLeaf.parent = leaf.parent

	firstKeyNewLeaf := newLeaf.keys[0] //prvi kljuc novog lista
	tree.insertIntoParent(leaf, firstKeyNewLeaf, newLeaf)

}

func (tree *BPlusTree) insertIntoParent(left *TreeNode, key string, right *TreeNode) {
	parent := left.parent
	if parent == nil {
		newRoot := createNode(tree.b, false)
		newRoot.keys[0] = key
		newRoot.keyCount = 1
		newRoot.children[0] = left
		newRoot.children[1] = right
		newRoot.childCount = 2
		left.parent = newRoot
		right.parent = newRoot
		tree.root = newRoot
		return
	}
	index := 0 //pronalazimo mesto za kljuc u roditelju
	for index < parent.keyCount && key > parent.keys[index] {
		index++
	}
	//pomeranje kljuceva u desno
	for i := parent.keyCount; i > index; i-- {
		parent.keys[i] = parent.keys[i-1]
	}
	parent.keys[index] = key
	parent.keyCount++

	//pomeranje dece u desno
	childPosition := index + 1
	for i := parent.childCount; i > childPosition; i-- {
		parent.children[i] = parent.children[i-1]
	}
	//dodavanje novog deteta
	parent.children[childPosition] = right
	parent.childCount++
	right.parent = parent

	//provera da li je roditeljski cvor prepunjen
	if parent.keyCount > 2*tree.b-1 {
		tree.splitInternal(parent)
	}
}

func (tree *BPlusTree) splitInternal(node *TreeNode) {
	middle := node.keyCount / 2
	firstKey := node.keys[middle]
	newNode := createNode(tree.b, false)
	for i := middle + 1; i < node.keyCount; i++ {
		newNode.keys[i-middle-1] = node.keys[i]
		node.keys[i] = ""
	}
	for i := middle + 1; i < node.childCount; i++ {
		newNode.children[i-middle-1] = node.children[i]
		newNode.children[i-middle-1].parent = newNode
		node.children[i] = nil
	}
	newNode.keyCount = node.keyCount - middle - 1
	newNode.childCount = node.childCount - middle - 1
	node.keyCount = middle
	node.childCount = middle + 1

	tree.insertIntoParent(node, firstKey, newNode)
}

// Funkcija vrsi pretragu po kljucu
// Povratna vrednost je par vrednost, bool (true ako je uspesno pronadjen)
func (tree *BPlusTree) Search(key string) ([]byte, bool) {
	leaf := tree.FindLeaf(key) //trazimo list u kom se nalazi kljuc
	for i := 0; i < leaf.keyCount; i++ {
		if leaf.keys[i] == key {
			return leaf.values[i], true //element je uspesno pronadjen
		}
	}
	return nil, false //element nije pronadjen
}

// Brisanje elementa
// Povratna vrednost je boolean koji govori o uspesnosti brisanja
func (tree *BPlusTree) Delete(key string) bool {
	leaf := tree.FindLeaf(key)
	index := -1
	for i := 0; i < leaf.keyCount; i++ {
		if leaf.keys[i] == key {
			index = i
			break //element je uspesno pronadjen
		}
	}
	if index == -1 {
		return false //kljuc nije pronadjen
	}
	//pomeranje elemenata ulevo
	for i := index; i < leaf.keyCount-1; i++ {
		leaf.keys[i] = leaf.keys[i+1]
		leaf.values[i] = leaf.values[i+1]
	}
	leaf.keys[leaf.keyCount-1] = ""
	leaf.values[leaf.keyCount-1] = nil

	leaf.keyCount--
	leaf.valueCount--
	tree.size--

	if leaf.keyCount < tree.b-1 {
		tree.rebalanceLeaf(leaf)
	}
	return true //element je uspesno obrisan
}

func (tree *BPlusTree) rebalanceLeaf(leaf *TreeNode) {
	parent := leaf.parent
	if parent == nil {
		return //koren
	}
	childIndex := -1
	for i := 0; i < parent.childCount; i++ {
		if parent.children[i] == leaf {
			childIndex = i
			break
		}
	}
	if childIndex == -1 {
		return
	}

	//pozajmljivanje od desnog brata
	if childIndex < parent.childCount-1 {
		rightSibling := parent.children[childIndex+1]
		if rightSibling.keyCount > tree.b-1 {
			//pozajmljivanje prvog elementa od desnog brata
			for i := leaf.keyCount; i > 0; i-- { //pomeranje listova u desno
				leaf.keys[i] = leaf.keys[i-1]
				leaf.values[i] = leaf.values[i-1]
			}
			//prvi element desnog brata
			leaf.keys[0] = rightSibling.keys[0]
			leaf.values[0] = rightSibling.values[0]
			leaf.keyCount++
			leaf.valueCount++

			//pomeranje elemenata desnog brata u levo
			for i := 0; i < rightSibling.keyCount-1; i++ {
				rightSibling.keys[i] = rightSibling.keys[i+1]
				rightSibling.values[i] = rightSibling.values[i+1]
			}
			rightSibling.keys[rightSibling.keyCount-1] = ""
			rightSibling.values[rightSibling.keyCount-1] = nil
			rightSibling.keyCount--
			rightSibling.valueCount--

			parent.keys[childIndex] = rightSibling.keys[0]
			return //pozajmljivanje je uspesno
		}
	}
	//pozajmljivanje od levog brata
	if childIndex > 0 {
		leftSibling := parent.children[childIndex-1]
		if leftSibling.keyCount > tree.b-1 {
			//pozajmljivanje poslednjeg elementa levog brata
			leaf.keys[leaf.keyCount] = leftSibling.keys[leftSibling.keyCount-1]
			leaf.values[leaf.valueCount] = leftSibling.values[leftSibling.valueCount-1]
			leaf.keyCount++
			leaf.valueCount++

			leftSibling.keys[leftSibling.keyCount-1] = ""
			leftSibling.values[leftSibling.valueCount-1] = nil
			leftSibling.keyCount--
			leftSibling.valueCount--
			parent.keys[childIndex-1] = leaf.keys[0]
			return //uspesno pozajmljivanje
		}
	}
	//spajanje sa desnim bratom
	if childIndex < parent.childCount-1 {
		rightSibling := parent.children[childIndex+1]
		tree.mergeLeaves(leaf, rightSibling, childIndex)
	} else if childIndex > 0 { //spajanje sa levim bratom
		leftSibling := parent.children[childIndex-1]
		tree.mergeLeaves(leftSibling, leaf, childIndex-1)
	}
}

func (tree *BPlusTree) mergeLeaves(left *TreeNode, right *TreeNode, index int) {
	//prenos svih kljuceva iz desnog u levi
	for i := 0; i < right.keyCount; i++ {
		left.keys[left.keyCount+i] = right.keys[i]
		left.values[left.valueCount+i] = right.values[i]
	}
	left.keyCount += right.keyCount
	left.valueCount += right.valueCount

	left.next = right.next
	parent := left.parent
	//pomeranje kljuceva roditelja ulevo
	for i := index; i < parent.keyCount-1; i++ {
		parent.keys[i] = parent.keys[i+1]

	}
	parent.keys[parent.keyCount-1] = ""
	parent.keyCount--
	//pomeranje dece u roditelju ulevo
	for i := index + 1; i < parent.childCount-1; i++ {
		parent.children[i] = parent.children[i+1]
	}
	parent.children[parent.childCount-1] = nil
	parent.childCount--
	if parent.keyCount < tree.b-1 && parent.parent != nil {
		tree.rebalanceInternal(parent)
	}
}

func (tree *BPlusTree) rebalanceInternal(node *TreeNode) {
	parent := node.parent
	if parent == nil {
		return //u pitanju je koren
	}
	childIndex := -1
	for i := 0; i < parent.childCount; i++ {
		if parent.children[i] == node {
			childIndex = i
			break
		}
	}
	if childIndex == -1 {
		return
	}
	//pozajmljivanje od desnog brata
	if childIndex < parent.childCount-1 {
		rightSibling := parent.children[childIndex+1]
		if rightSibling.keyCount > tree.b-1 {
			parentKey := parent.keys[childIndex]
			for i := node.keyCount; i > 0; i-- {
				node.keys[i] = node.keys[i-1]
			}
			node.keys[0] = parentKey
			node.keyCount++

			for i := node.childCount; i > 0; i-- {
				node.children[i] = node.children[i-1]
			}

			node.children[0] = rightSibling.children[0]
			node.children[0].parent = node
			node.childCount++

			for i := 0; i < rightSibling.keyCount-1; i++ {
				rightSibling.keys[i] = rightSibling.keys[i+1]
			}
			rightSibling.keys[rightSibling.keyCount-1] = ""
			rightSibling.keyCount--

			for i := 0; i < rightSibling.childCount-1; i++ {
				rightSibling.children[i] = rightSibling.children[i+1]
			}
			rightSibling.children[rightSibling.childCount-1] = nil
			rightSibling.childCount--
			parent.keys[childIndex] = rightSibling.keys[0]
			return
		}
	}
	//pozajmljivanje od levog brata
	if childIndex > 0 {
		leftSibling := parent.children[childIndex-1]
		if leftSibling.keyCount > tree.b-1 {
			parentKey := parent.keys[childIndex-1]
			node.keys[node.keyCount] = parentKey
			node.keyCount++

			node.children[node.childCount] = leftSibling.children[leftSibling.childCount-1]
			node.children[node.childCount].parent = node
			node.childCount++

			leftSibling.children[leftSibling.childCount-1] = nil
			leftSibling.childCount--

			parent.keys[childIndex-1] = leftSibling.keys[leftSibling.keyCount-1]
			leftSibling.keys[leftSibling.keyCount-1] = ""
			leftSibling.keyCount--
			return
		}
	}
	//spajanje sa desnim bratom
	if childIndex < parent.childCount-1 {
		rightSibling := parent.children[childIndex+1]
		tree.mergeInternal(node, rightSibling, childIndex)
	} else if childIndex > 0 { //spaajanje sa levim bratom
		leftSibling := parent.children[childIndex-1]
		tree.mergeInternal(leftSibling, node, childIndex-1)
	}

}

func (tree *BPlusTree) mergeInternal(left *TreeNode, right *TreeNode, index int) {
	parent := left.parent
	left.keys[left.keyCount] = parent.keys[index]
	left.keyCount++
	for i := 0; i < right.keyCount; i++ {
		left.keys[left.keyCount+i] = right.keys[i]
	}
	left.keyCount += right.keyCount
	for i := 0; i < right.childCount; i++ {
		left.children[left.childCount+i] = right.children[i]
		left.children[left.childCount+i].parent = left
	}
	left.childCount += right.childCount
	for i := index; i < parent.keyCount-1; i++ {
		parent.keys[i] = parent.keys[i+1]
	}
	parent.keys[parent.keyCount-1] = ""
	parent.keyCount--

	for i := index + 1; i < parent.childCount-1; i++ {
		parent.children[i] = parent.children[i+1]
	}
	parent.children[parent.childCount-1] = nil
	parent.childCount--
	if parent.keyCount < tree.b-1 && parent.parent != nil {
		tree.rebalanceInternal(parent)
	}
}

func (tree *BPlusTree) RangeScan(startKey, endKey string) []struct {
	Key   string
	Value []byte
} {
	var result []struct {
		Key   string
		Value []byte
	}
	//pronalazenje pocetnog lista
	startLeaf := tree.FindLeaf(startKey)
	current := startLeaf
	for current != nil {
		for i := 0; i < current.keyCount; i++ {
			key := current.keys[i]
			if key >= startKey && key <= endKey {
				result = append(result, struct {
					Key   string
					Value []byte
				}{
					Key:   key,
					Value: current.values[i],
				})
			}
			if key > endKey {
				return result
			}

		}
		current = current.next //prelazak na sledeci list
	}
	return result
}

func (tree *BPlusTree) PrefixScan(prefix string) []struct {
	Key   string
	Value []byte
} {
	var result []struct {
		Key   string
		Value []byte
	}

	//pronalazak lista sa prviom kljucem koji pocinje sa prefiksom
	//ili prvi list koji bi mogao da sadrzi takve kljuceve
	startLeaf := tree.FindLeaf(prefix)
	current := startLeaf

	for current != nil {
		for i := 0; i < current.keyCount; i++ {
			key := current.keys[i]
			//provera da li kljuc pocinje sa prefiksom
			if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				result = append(result, struct {
					Key   string
					Value []byte
				}{
					Key:   key,
					Value: current.values[i],
				})
			} else if key > prefix && !startsWithPrefix(key, prefix) {
				//posto su kljucevi sortirani, ako je ovaj uslov ispunjen
				//znacemo da smo prosli sve kljuceve sa unetim prefiksom
				break
			}
		}
		current = current.next //ako smo stigli do kraja lista, proveriti sledeci u sliucaju da ima jos

	}
	return result
}

func startsWithPrefix(key, prefix string) bool {
	if len(key) < len(prefix) {
		return false
	}
	return key[:len(prefix)] == prefix
}
