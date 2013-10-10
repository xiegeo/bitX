package hashtree

import (
	"math"
)

type Level int
type Bytes int64
type Nodes int

func NodesFromBytes(len Bytes, blockSize Bytes) Nodes {
	return Nodes((len-1)/blockSize) + 1
}

func Levels(n Nodes) Level {
	return Level(math.Ilogb(float64(n*2-1)) + 1)
}

func LevelWidth(n Nodes, l Level) Nodes {
	if l < 0 {
		return 0
	}
	for l > 0 {
		n = (n + 1) / 2
		l--
	}
	return n
}

func HashNumber(leafs Nodes, l Level, n Nodes) int64 {
	sum := Nodes(0)
	for i := Level(0); i < l; i++ {
		sum += LevelWidth(leafs, i)
	}
	return int64(sum + n)
}

func HashTreeSize(leafs Nodes) int64 {
	return HashNumber(leafs, Levels(leafs), 0)
}

func HashPosition(leafs Nodes, l Level, n Nodes) Bytes {
	return Bytes(HashNumber(leafs, l, n)) * Bytes(treeNodeSize)
}
