package network

import (
	"../../hashtree"
	"fmt"
)

// additional functions to complement network.pb.go

func (id *StaticId) CompactId() string {
	return fmt.Sprintf("%x-%d", id.GetHash(), id.GetLength())
}

func (id *StaticId) Bytes() hashtree.Bytes {
	return hashtree.Bytes(id.GetLength())
}

func (id *StaticId) Blocks() hashtree.Nodes {
	return hashtree.FileNodes(id.Bytes(), hashtree.FILE_BLOCK_SIZE)
}

func (id *StaticId) WidthForLevelOf(in *InnerHashes) hashtree.Nodes {
	return hashtree.LevelWidth(id.Blocks(), hashtree.Level(in.GetHeight()))
}

func (in *InnerHashes) CheckWellFormedForId(id *StaticId) error {
	if in.GetFrom() < 0 {
		return fmt.Errorf("from %v is less than 0", in.GetFrom())
	}
	if in.GetLength()*hashtree.HASH_BYTES != int32(len(in.GetHashes())) {
		return fmt.Errorf("reported length %v blocks != length of data %v bytes", in.GetLength(), len(in.GetHashes()))
	}
	width := id.WidthForLevelOf(in)
	if in.GetFrom()+in.GetLength() > int32(width) {
		return fmt.Errorf("hash form %v for %v is longer than width of %v", in.GetFrom(), in.GetLength(), width)
	}

	//todo: more checks
	return nil
}

//
func (in *InnerHashes) LocalSum() []byte {
	c := hashtree.NewNoPadTree()
	c.Write(in.GetHashes())
	return c.Sum(nil)
}

func logb(n hashtree.Nodes) hashtree.Nodes {
	i := hashtree.Nodes(0)
	for ; n >= 16; i += 5 {
		n /= 32
	}
	for ; n > 0; i++ {
		n /= 2
	}
	return i
}

func expb(n hashtree.Nodes) hashtree.Nodes {
	if n == 0 {
		return 0
	}
	n--
	i := hashtree.Nodes(1)
	for ; n >= 5; i *= 32 {
		n -= 5
	}
	for ; n > 0; i *= 2 {
		n--
	}
	return i
}

func (in *InnerHashes) Part(from hashtree.Nodes, to hashtree.Nodes) *InnerHashes {
	h := in.GetHeight()
	f := int32(from)
	l := int32(to - from)
	return &InnerHashes{
		Height: &h,
		From:   &f,
		Length: &l,
		Hashes: in.Hashes[(int32(from)-in.GetFrom())*hashtree.HASH_BYTES : (to-from)*hashtree.HASH_BYTES],
	}
}

func (in *InnerHashes) Parts(l [][2]hashtree.Nodes) []*InnerHashes {
	r := make([]*InnerHashes, len(l))
	for _, v := range l {
		r = append(r, in.Part(v[0],v[1]))
	}
	return r
}

func mergeR(a [][2]hashtree.Nodes, b [][2]hashtree.Nodes) [][2]hashtree.Nodes {
	result := make([][2]hashtree.Nodes, len(a)+len(b))
	copy(result, a)
	copy(result[len(a):], b)
	return result
}

func sls(from hashtree.Nodes, to hashtree.Nodes, width hashtree.Nodes) [][2]hashtree.Nodes {
	if to-from <= 1 {
		return nil
	}
	if from == 0 {
		dev := expb(logb(to + 1))
		if to == width-1 || to == dev-1 {
			return [][2]hashtree.Nodes{{from, to}}
		}
		return mergeR(sls(from, dev-1, width), sls(dev, to, width))
	} else {
		dev := expb(logb(from) + 1)
		if to >= dev {
			return mergeR(sls(from, dev-1, width), sls(dev, to, width))
		} else if to == dev-1 {
			return [][2]hashtree.Nodes{{from, to}}
		}
		zoom := expb(logb(to - from))
		dev -= zoom
		return mergeR(sls(from, dev-1, width), sls(dev, to, width))
	}
}

func (in *InnerHashes) SplitLocalSummable(id *StaticId) []*InnerHashes {
	if err := in.CheckWellFormedForId(id); err != nil {
		panic(err)
	}
	nextEven := hashtree.Nodes(in.GetFrom()+1) / 2 * 2
	ranges := sls(nextEven, hashtree.Nodes(in.GetFrom()+in.GetLength()), id.WidthForLevelOf(in))
	return in.Parts(ranges)
}
