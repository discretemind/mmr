package mmr

//IMerkleIndex index navigator
type IMerkleIndex interface {
	GetHeight() uint64
	GetLeftBranch() IMerkleIndex
	GetSibling() IMerkleIndex
	RightUp() IMerkleIndex

	IsRight() bool
	IsObject() bool
	GetTop() IMerkleIndex
	Index() uint64
	Children() []IMerkleIndex
	Hash(source ISource) ([]byte, error)
	//Value(source IMerkleStore) (*nodeData, bool)
}

//IObjectIndex object index navigator
type IObjectIndex interface {
	IMerkleIndex
	GetPeaks() []IMerkleIndex
}

var index1 = uint64(1)

type objectIndex uint64

//ObjectIndex object index navigator
func ObjectIndex(index uint64) (res IObjectIndex) {
	v := objectIndex(index)
	return &v
}

func (x *objectIndex) GetLeftBranch() IMerkleIndex {
	value := uint64(*x)
	if value&1 == 0 && value > 1 {
		return NodeIndex(uint64(*x) - 1)
	}
	return nil
}

func (x *objectIndex) IsObject() bool {
	return true
}

func (x *objectIndex) GetSibling() IMerkleIndex {
	value := uint64(*x)
	if value&1 == 1 {
		return ObjectIndex(value - 1)
	}
	return ObjectIndex(value + 1)
}

func (x *objectIndex) RightUp() IMerkleIndex {
	value := x.Index()
	if value&1 == 1 {
		return NodeIndex(value)
	}
	return nil
}

func (x *objectIndex) GetTop() IMerkleIndex {
	value := uint64(*x)
	if value&1 == 0 {
		return ObjectIndex(value)
	}
	return NodeIndex(value).GetTop()
}

func (x *objectIndex) Index() uint64 {
	return uint64(*x)
}

// Calculates Peaks
// Algorythm:
//  1. Get to from current. Take it.
//  2. Go to the left branch.
//     - if no any left brnaches - return
//     - go to 1
func (x *objectIndex) GetPeaks() (res []IMerkleIndex) {
	var peak IMerkleIndex = x
	for {
		peak = peak.GetTop()
		res = append(res, peak)
		if peak = peak.GetLeftBranch(); peak == nil {
			return
		}
	}
}

// Leaf is always on the Zero height
func (x *objectIndex) GetHeight() uint64 {
	return 0
}

func (x *objectIndex) IsRight() bool {
	return x.Index()&1 == 1
}

func (x *objectIndex) Hash(store ISource) (key []byte, err error) {
	obj, err := store.Object(x.Index())
	if err != nil {
		return nil, err
	}
	return obj.Hash(), nil
}

func (x *objectIndex) Children() []IMerkleIndex {
	return nil
}

//
//func (x *objectIndex) SetValue(mmr *mmr, data *BlockData) {
//	mmr.db.SetBlock(x.Index(), data)
//}
//
//func (x *objectIndex) Value(mmr *mmr) (*BlockData, bool) {
//	return mmr.db.GetBlock(x.Index())
//}
//
//func (x *objectIndex) AppendValue(mmr *mmr, data *BlockData) {
//	mmr.db.SetBlock(x.Index(), data)
//	var node IMerkleIndex = x
//	for node.IsRight() {
//		sibling := node.GetSibling()
//		if node = node.RightUp(); node != nil {
//			leftData, _ := sibling.Value(mmr)
//			node.SetValue(mmr, mmr.aggregate(leftData, data))
//			continue
//		}
//		return
//	}
//}
//
