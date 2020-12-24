package mmr

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

var (
	//ErrKeyNotFound key not found error
	ErrKeyNotFound = errors.New("Key not found")
)

type merkleData struct {
	sync.RWMutex
	nodes   map[uint64]INode
	objects map[uint64]IObjectNode
	root    *Root
}

func newMerkleData() ISource {
	res := &merkleData{
		nodes:   make(map[uint64]INode),
		objects: make(map[uint64]IObjectNode),
	}
	return res
}

func (s *merkleData) Node(id uint64) (INode, error) {
	val, ok := s.nodes[id]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return val, nil
}

func (s *merkleData) Object(id uint64) (IObjectNode, error) {
	val, ok := s.objects[id]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return val, nil
}

func (s *merkleData) SetNode(id uint64, val INode) error {
	s.nodes[id] = val
	return nil
}

func (s *merkleData) SetObject(id uint64, val IObjectNode) error {
	s.objects[id] = val
	return nil
}

func (s *merkleData) Root() *Root {
	return s.root
}

func (s *merkleData) SetRoot(root *Root) {
	s.root = root
}

func (s *merkleData) Dump() (res string) {
	res += fmt.Sprintln("Dump Memory DB")
	res += fmt.Sprintf("Objects: %d\n", len(s.objects))

	keys := make([]int, 0)
	for k := range s.objects {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	for _, k := range keys {
		val := s.objects[uint64(k)]
		res += fmt.Sprintf("\t%d %x\n", k, val.Hash())
	}

	res += fmt.Sprintf("Nodes: %d\n", len(s.nodes))
	keys = []int{}
	for k := range s.nodes {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		val := s.nodes[uint64(k)]
		res += fmt.Sprintf("\t%d %x\n", k, val.Hash())
	}
	return res
}
