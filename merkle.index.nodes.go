package mmr

import (
	"math"
)

type nodeIndex uint64

//NodeIndex node index navigator
func NodeIndex(index uint64) (res IMerkleIndex) {
	v := nodeIndex(index)
	return &v
}

func (x *nodeIndex) GetHeight() uint64 {
	return getHeight(uint64(*x))
}

func (x *nodeIndex) GetLeftBranch() IMerkleIndex {
	pow := uint64(math.Pow(2, float64(x.GetHeight()+1)))
	value := x.Index()
	if value > pow {
		return NodeIndex(value - pow)
	}
	return nil
}

func (x *nodeIndex) GetSibling() IMerkleIndex {
	value := x.Index()
	shift := index1 << (x.GetHeight() + 1)
	return NodeIndex(value ^ shift)
}

func (x *nodeIndex) IsRight() bool {
	shift := index1 << (x.GetHeight() + 1)
	return shift&x.Index() == shift
}

func (x *nodeIndex) RightUp() IMerkleIndex {
	value := x.Index()

	shift := index1 << (x.GetHeight() + 1)
	if shift&value == shift {
		value ^= index1 << x.GetHeight()
		if value > 0 {
			return NodeIndex(value)
		}
	}

	return nil
}

func (x *nodeIndex) GetTop() IMerkleIndex {
	shift := index1 << x.GetHeight()
	value := uint64(*x)
	result := value
	for value != 0 && value&shift == shift {
		result = value
		value = value ^ shift
		shift <<= 1
	}
	return NodeIndex(result)
}

func (x *nodeIndex) IsObject() bool {
	return false
}

//func (x *nodeIndex) Value(source IMerkleStore) (*nodeData, bool) {
//	return nil, false
//}

func (x *nodeIndex) Children() []IMerkleIndex {
	h := x.GetHeight()
	pow := uint64(math.Pow(2, float64(h))) / 2
	index := x.Index()
	if h == 0 {
		return []IMerkleIndex{ObjectIndex(index - 1), ObjectIndex(index)}
	}
	return []IMerkleIndex{NodeIndex(index - pow), NodeIndex(index + pow)}

}

func (x *nodeIndex) Index() uint64 {
	return uint64(*x)
}

func (x *nodeIndex) Hash(store ISource) (key []byte, err error) {
	obj, err := store.Node(x.Index())
	if err != nil {
		return nil, err
	}
	return obj.Hash(), nil
}

//func (x *nodeIndex) SetValue(mmr *mmr, data *BlockData) {
//	mmr.db.SetNode(x.Index(), data)
//}
//
////func (x *nodeIndex) AppendValue(mmr *mmr, data *BlockData) {
////	mmr.db.SetNode(x.Index(), data)
////}
//
//func (x *nodeIndex) Value(mmr *mmr) (*BlockData, bool) {
//	return mmr.db.GetNode(x.Index())
//}

func getHeight(value uint64) (height uint64) {
	for {
		if value == 0 || value&1 == 1 {
			return
		}
		value = value >> 1
		height++
	}
}
