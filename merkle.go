package mmr

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/discretemind/mmr/db"
	"hash"
	"sync"
)

//Hasher replaceable  hasher function
type Hasher func() hash.Hash

type rootModel struct {
	Index uint64
	Peeks [][]byte
}

//ISource MMR data model
type ISource interface {
	Node(uint64) (INode, error)
	SetNode(uint64, INode) error
	Object(uint64) (IObjectNode, error)
	SetObject(uint64, IObjectNode) error
	SetRoot(*Root)
	Root() *Root
	Dump() string
}

// INode node model
type INode interface {
	Hash() []byte
}

//IObjectNode object model
type IObjectNode interface {
	INode
	Payload() []byte
}

// Root mmr Root model
type Root struct {
	Hash  []byte
	Index IMerkleIndex
	Peeks []IMerkleIndex
}

type node struct {
	hash []byte
}

func (n *node) Hash() []byte {
	return n.hash
}

type objectNode struct {
	node
	payload []byte
}

func (n *objectNode) Payload() []byte {
	return n.payload
}

//RootNode Root Node
type RootNode struct {
	node
	Index uint64
}

//IMerkle Merkle Mountain Range interface
type IMerkle interface {
	Add(value interface{}) (root []byte, err error)
	Get(index uint64, value interface{}) (err error)
	Save(db.IDatabase) error
	Root() *RootNode
}

type merkleSource struct {
	sync.RWMutex
	hasher Hasher
	source ISource
	salt   []byte
}

//Merkle MMR instance
func Merkle(hasher Hasher) IMerkle {
	res := &merkleSource{
		hasher: hasher,
		source: newMerkleData(),
	}
	res.salt = make([]byte, 8)
	rand.Read(res.salt)
	return res
}

//MerkleFromSource load MMR from database
func MerkleFromSource(hasher Hasher, root []byte, db db.IDatabase) (IMerkle, error) {
	res := &merkleSource{
		hasher: hasher,
		source: newMerkleData(),
	}

	if err := res.load(db, root); err != nil {
		return nil, err
	}

	return res, nil
}

func (ms *merkleSource) Add(value interface{}) (data []byte, err error) {
	newObj := &objectNode{}
	newObj.payload, err = json.Marshal(value)
	if err != nil {
		return nil, err
	}
	newObj.hash = ms.getHash(newObj.payload)

	if err := ms.appendMerkle(newObj); err != nil {
		return nil, err

	}
	return ms.source.Root().Hash, nil
}

func (ms *merkleSource) Get(index uint64, value interface{}) (err error) {
	root := ms.source.Root()

	if index > root.Index.Index() {
		return errors.New("Out of range")
	}
	obj, err := ms.source.Object(index)
	if err != nil {
		return err
	}
	return json.Unmarshal(obj.Payload(), value)
}

func (ms *merkleSource) Root() *RootNode {
	root := ms.source.Root()
	h := make([]byte, len(root.Hash))
	copy(h, root.Hash)
	return &RootNode{
		Index: root.Index.Index(),
		node: node{
			hash: h,
		},
	}
}

func (ms *merkleSource) saveNode(d db.IDatabase, index IMerkleIndex, hash []byte, source ISource) error {
	if d.Has(hash) {
		return nil
	}

	if index.IsObject() {
		obj, err := source.Object(index.Index())
		if err != nil {
			return err
		}
		return d.Set(hash, obj.Payload())
	}

	children := index.Children()
	var childHashes []byte
	for _, child := range children {
		h, err := child.Hash(source)
		if err != nil {
			return err
		}
		if err := ms.saveNode(d, child, h, source); err != nil {
			return err
		}
		childHashes = append(childHashes, h...)
	}
	return d.Set(hash, childHashes)

}

func (ms *merkleSource) loadNode(database db.IDatabase, isObject bool, index uint64, hash []byte) error {
	hashSize := ms.hasher().Size()

	content, err := database.Get(hash)
	if err != nil {
		return err
	}
	if isObject {
		if err := ms.source.SetObject(index, &objectNode{
			node: node{
				hash: hash,
			},
			payload: content,
		}); err != nil {
			return err
		}
	} else {
		if err := ms.source.SetNode(index, &node{
			hash: hash,
		}); err != nil {
			return err
		}
		children := NodeIndex(index).Children()
		for i, child := range children {
			h := content[i*hashSize : (i+1)*hashSize]
			if err := ms.loadNode(database, child.IsObject(), child.Index(), h); err != nil {
				return err
			}
		}

	}

	return nil
}

func (ms *merkleSource) load(database db.IDatabase, root []byte) error {
	data, err := database.Get(root)
	if err != nil {
		return err
	}
	if len(data) < 8 {
		return errors.New("Invalid data size for root")
	}

	rm := &rootModel{}

	rm.Index, data, err = readUvarint(data)
	ms.salt, data = data[:8], data[8:]
	if err != nil {
		return err
	}

	var index uint64
	var h []byte
	var peeks []IMerkleIndex
	for len(data) > 1 {
		isObject := data[0] == 1
		index, data, err = readUvarint(data[1:])
		if err != nil {
			return err
		}
		h, data = data[:32], data[32:]
		if err := ms.loadNode(database, isObject, index, h); err != nil {
			return err
		}

		if isObject {
			peeks = append(peeks, ObjectIndex(index))
		} else {
			peeks = append(peeks, NodeIndex(index))
		}
	}
	ms.source.SetRoot(&Root{
		Hash:  root,
		Index: ObjectIndex(rm.Index),
		Peeks: peeks,
	})

	return nil
}

func (ms *merkleSource) Save(d db.IDatabase) error {
	root := ms.source.Root()
	rm := &rootModel{
		Index: root.Index.Index(),
	}

	data := writeUvarint(rm.Index, []byte{})
	data = append(data, ms.salt...)
	for _, peek := range root.Peeks {
		h, err := peek.Hash(ms.source)
		if err != nil {
			return err
		}

		if peek.IsObject() {
			data = append(data, 1)
		} else {
			data = append(data, 0)
		}

		data = writeUvarint(peek.Index(), data)
		data = append(data, h...)
		if err := ms.saveNode(d, peek, h, ms.source); err != nil {
			return err
		}
		rm.Peeks = append(rm.Peeks, h)
	}
	return d.Set(root.Hash, data)
}

func (ms *merkleSource) getHash(data []byte) (res []byte) {
	h := ms.hasher()
	h.Write(data[:])
	res = make([]byte, h.Size())
	copy(res[:], h.Sum([]byte{}))
	return
}

func (ms *merkleSource) aggregate(left, right []byte) (result node) {
	h := ms.hasher()
	h.Write(ms.salt[:])
	h.Write(right[:])
	h.Write(left[:])
	result = node{}
	result.hash = make([]byte, h.Size())
	copy(result.hash[:], h.Sum([]byte{}))
	return
}

func (ms *merkleSource) appendMerkle(obj *objectNode) (err error) {
	ms.Lock()
	defer ms.Unlock()
	root := ms.source.Root()

	var index uint64 = 0
	if root != nil {
		index = ms.source.Root().Index.Index() + 1
	}

	objIndex := ObjectIndex(index)
	if err := ms.source.SetObject(objIndex.Index(), obj); err != nil {
		return err
	}

	if err := ms.recalculateNodes(objIndex); err != nil {
		return err
	}
	// Calculating root
	peaks := objIndex.GetPeaks()
	rootNode := &RootNode{
		Index: index,
		node: node{
			hash: []byte("~ROOT~"),
		},
	}

	for _, peak := range peaks {
		h, err := peak.Hash(ms.source)
		if err != nil {
			return err
		}
		rootNode.node = ms.aggregate(h, rootNode.node.hash)
	}

	ms.source.SetRoot(&Root{
		Index: objIndex,
		Peeks: peaks,
		Hash:  rootNode.hash,
	})
	return nil
}

func (ms *merkleSource) recalculateNodes(objIndex IMerkleIndex) error {
	var node = objIndex
	for node.IsRight() {
		sibling := node.GetSibling()
		if parent := node.RightUp(); parent != nil {
			leftData, lok := sibling.Hash(ms.source)
			rightData, rok := node.Hash(ms.source)
			if lok == nil && rok == nil {
				nd := ms.aggregate(leftData, rightData)
				if err := ms.source.SetNode(parent.Index(), &nd); err != nil {
					return nil
				}
			}
			node = parent
			continue
		}
		break
	}
	return nil
}

func readUvarint(data []byte) (uint64, []byte, error) {
	res, bytes := binary.Uvarint(data)
	if bytes == 0 {
		return 0, nil, errors.New("Invalid data")
	}
	return res, data[bytes:], nil
}

func writeUvarint(num uint64, data []byte) []byte {
	variant := make([]byte, 8)
	n := binary.PutUvarint(variant, num)

	return append(data, variant[:n]...)
}
