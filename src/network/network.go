package network

import (
	"../hashtree"
	"bytes"
	"fmt"
)

// additional functions to complement network.pb.go

// test if two ids are the same (in hash and length), panic if ether or both are nil
func (id StaticId) Equal(other StaticId) bool {
	if id.GetLength() != other.GetLength() {
		return false
	}
	return bytes.Equal(id.GetHash(), other.GetHash())
}

// return an url and file name safe string that uniquely represent this file
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

func NewFileData(from, length hashtree.Bytes, bytes []byte) FileData {
	f := int64(from)
	l := int32(length)
	return FileData{
		From:     &f,
		Length:   &l,
		InBlocks: nil,
		Data:     bytes,
	}
}

func (m *FileData) GetFromB() hashtree.Bytes {
	return hashtree.Bytes(m.GetFrom())
}

func (m *FileData) GetLengthB() hashtree.Bytes {
	if m.GetLength() == 0 {
		return hashtree.Bytes(len(m.GetData()))
	}
	return hashtree.Bytes(m.GetLength())
}

func shortString(data []byte) string {
	l := len(data)
	if l == 0 {
		return "(empty)"
	} else if l <= 32 {
		return fmt.Sprintf("(%v)%X", l, data)
	} else {
		return fmt.Sprintf("(%v)%X ... %X", l, data[:8], data[l-8:])
	}
}

func (m *FileData) ShortString() string {
	return fmt.Sprintf("from:%v length:%v block size:%v data:%v", m.GetFrom(), m.GetLength(), m.GetInBlocks(), shortString(m.GetData()))
}

func NewInnerHashes(height hashtree.Level, from hashtree.Nodes, length hashtree.Nodes, bytes []byte) InnerHashes {
	h := int32(height)
	f := int32(from)
	l := int32(length)
	return InnerHashes{
		Height: &h,
		From:   &f,
		Length: &l,
		Hashes: bytes,
	}
}

func (in *InnerHashes) GetHeightL() hashtree.Level {
	return hashtree.Level(in.GetHeight())
}

func (in *InnerHashes) GetLengthN() hashtree.Nodes {
	return hashtree.Nodes(in.GetLength())
}

func (in *InnerHashes) GetLengthB() hashtree.Bytes {
	if in.GetLengthN() == 0 {
		return hashtree.Bytes(len(in.GetHashes()))
	}
	return hashtree.Bytes(in.GetLengthN() * hashtree.HASH_BYTES)
}

func (in *InnerHashes) GetFromN() hashtree.Nodes {
	return hashtree.Nodes(in.GetFrom())
}

//check remote messages for problems that code protected by this don't have to worry about
func (in *InnerHashes) CheckWellFormedForId(id *StaticId) error {
	if in.GetFrom() < 0 {
		return fmt.Errorf("from %v is less than 0", in.GetFrom())
	}
	hashSize := len(in.GetHashes())
	if int(in.GetLength()*hashtree.HASH_BYTES) != hashSize {
		if hashSize > 0 && hashSize%hashtree.HASH_BYTES == 0 {
			//if length is not send or set, we can add it from the length of the hash
			length := int32(hashSize / hashtree.HASH_BYTES)
			in.Length = &length
		} else if hashSize == 0 {
			//no hash here yet, just a request length, this is ok.
		} else {
			return fmt.Errorf("reported length %v blocks != length of data %v bytes", in.GetLength(), len(in.GetHashes()))
		}
	}
	width := id.WidthForLevelOf(in)
	if in.GetFromN()+in.GetLengthN() > width {
		return fmt.Errorf("hash form %v for %v is longer than width of %v", in.GetFrom(), in.GetLength(), width)
	}

	//todo: more checks
	return nil
}

//Only use under SplitLocalSummable.
//Get the bytes of the local root sum.
func (in *InnerHashes) LocalSum() []byte {
	c := hashtree.NewNoPadTree()
	c.Write(in.GetHashes())
	return c.Sum(nil)
}

//Only use under SplitLocalSummable
//Get the position of the local root sum.
func (in *InnerHashes) LocalRoot(leafs hashtree.Nodes) (level hashtree.Level, node hashtree.Nodes) {
	h := in.GetHeightL()
	f := in.GetFromN()
	l := in.GetLengthN()
	n := hashtree.Levels(l) - 1
	level = h + n
	node = f >> uint32(n)
	return
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

func (in *InnerHashes) Part(from hashtree.Nodes, to hashtree.Nodes) InnerHashes {
	length := to - from + 1
	hashStart := from - in.GetFromN()
	return NewInnerHashes(in.GetHeightL(), from, length,
		in.Hashes[hashStart*hashtree.HASH_BYTES:(hashStart+length)*hashtree.HASH_BYTES])
}

func (in *InnerHashes) Parts(l [][2]hashtree.Nodes) []InnerHashes {
	r := make([]InnerHashes, 0, len(l))
	for _, v := range l {
		r = append(r, in.Part(v[0], v[1]))
	}
	return r
}

func mergeR(a [][2]hashtree.Nodes, b [][2]hashtree.Nodes) [][2]hashtree.Nodes {
	result := make([][2]hashtree.Nodes, len(a)+len(b))
	copy(result, a)
	copy(result[len(a):], b)
	return result
}

func shiftsls(sls [][2]hashtree.Nodes, delta hashtree.Nodes) [][2]hashtree.Nodes {
	for i := 0; i < len(sls); i++ {
		sls[i][0] += delta
		sls[i][1] += delta
	}
	return sls
}

// the input maybe gernerated by an adversary, return nil instead of panic that check coding errors
func slsUntrusted(from hashtree.Nodes, to hashtree.Nodes, width hashtree.Nodes) [][2]hashtree.Nodes {
	if from > to || to >= width {
		return nil
	} else {
		return sls(from, to, width)
	}
}

func sls(from hashtree.Nodes, to hashtree.Nodes, width hashtree.Nodes) [][2]hashtree.Nodes {
	from = (from + 1) / 2 * 2
	if from > to || to >= width {
		panic(fmt.Sprintf("from:%v, to:%v, width:%v", from, to, width))
	}

	if from == to {
		//there souldn't be any singles, unless it is the last one and even
		if from == width-1 && from%2 == 0 {
			return [][2]hashtree.Nodes{{from, to}}
		}
		return nil
	}
	if from == 0 {
		dev := expb(logb(to + 1))
		//log.Println(from,to,width,dev);
		if to == width-1 || to == dev-1 {
			return [][2]hashtree.Nodes{{from, to}}
		}
		return mergeR(sls(from, dev-1, dev), shiftsls(sls(0, to-dev, width-dev), dev))
	} else {
		dev := expb(logb(width - 1))
		//log.Println(from,to,width,dev);
		if from < dev {
			if to < dev {
				return sls(from, to, dev)
			} else {
				return mergeR(sls(from, dev-1, dev), shiftsls(sls(0, to-dev, width-dev), dev))
			}
		} else {
			return shiftsls(sls(from-dev, to-dev, width-dev), dev)
		}
	}
}

//Split inner hashes based on highest derivable ancestors
func (in *InnerHashes) SplitLocalSummable(id *StaticId) []InnerHashes {
	if err := in.CheckWellFormedForId(id); err != nil {
		panic(err)
	}
	ranges := slsUntrusted(hashtree.Nodes(in.GetFrom()), hashtree.Nodes(in.GetFrom()+in.GetLength())-1, id.WidthForLevelOf(in))
	return in.Parts(ranges)
}
