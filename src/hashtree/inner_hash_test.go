package hashtree

import (
	"testing"
)

type treeLevels struct {
	len    Bytes
	levels Level
}

var expectedTreeLevels = []treeLevels{
	{0, 1}, {1, 1}, {treeNodeSize, 1},
	{treeNodeSize + 1, 2}, {2 * treeNodeSize, 2},
	{2*treeNodeSize + 1, 3}, {4 * treeNodeSize, 3},
	{4*treeNodeSize + 1, 4}, {8 * treeNodeSize, 4},
	{8*treeNodeSize + 1, 5}, {16 * treeNodeSize, 5},
}

func TestTreeLevels(t *testing.T) {
	h := NewTree()
	for i := 0; i < len(expectedTreeLevels); i++ {
		e := expectedTreeLevels[i]
		if h.Levels(e.len) != e.levels {
			t.Fatalf("Levels(%d) = %d want %d", e.len, h.Levels(e.len), e.levels)
		}
	}
}

type levelWidth struct {
	len   Bytes
	level Level
	width Nodes
}

var expectedLevelWidth = []levelWidth{
	{0, 0, 1}, {1, 0, 1}, {treeNodeSize, 0, 1},

	{treeNodeSize + 1, 0, 2}, {2 * treeNodeSize, 0, 2},
	{treeNodeSize + 1, 1, 1}, {2 * treeNodeSize, 1, 1},

	{2*treeNodeSize + 1, 0, 3}, {4 * treeNodeSize, 0, 4},
	{2*treeNodeSize + 1, 1, 2}, {4 * treeNodeSize, 1, 2},
	{2*treeNodeSize + 1, 2, 1}, {4 * treeNodeSize, 2, 1},

	{5 * treeNodeSize, 0, 5}, {6 * treeNodeSize, 0, 6}, {7 * treeNodeSize, 0, 7}, {8 * treeNodeSize, 0, 8},
	{5 * treeNodeSize, 1, 3}, {6 * treeNodeSize, 1, 3}, {7 * treeNodeSize, 1, 4}, {8 * treeNodeSize, 1, 4},
	{5 * treeNodeSize, 2, 2}, {6 * treeNodeSize, 2, 2}, {7 * treeNodeSize, 2, 2}, {8 * treeNodeSize, 2, 2},
	{5 * treeNodeSize, 3, 1}, {6 * treeNodeSize, 3, 1}, {7 * treeNodeSize, 3, 1}, {8 * treeNodeSize, 3, 1},
}

func TestLevelWidth(t *testing.T) {
	h := NewTree()
	for i := 0; i < len(expectedLevelWidth); i++ {
		e := expectedLevelWidth[i]
		if h.LevelWidth(e.len, e.level) != e.width {
			t.Fatalf("LevelWidth(%d,%d) = %d want %d", e.len, e.level, h.LevelWidth(e.len, e.level), e.width)
		}
	}
}

func TestInnerHashListener(t *testing.T) {
	inner := [][]int32{
		{1, 1, 2, 3, 5, 8},
		{0, -1, -3},
		{1, -3},
		{4},
	}
	listener := func(l Level, i Nodes, hash *H256) {
		h := int32(hash[0])
		if inner[l][i] != h {
			if inner[l][i] == h+2000 {
				t.Fatalf("Level:%d, Node:%d was repeated", l, i)
			}
			t.Fatalf("Level:%d, Node:%d, hash:%d, should be %d", l, i, h, inner[l][i])
		} else {
			inner[l][i] += 2000 //mark heard
		}
	}
	c := NewTree2(NoPad32bytes, minus).(*treeDigest)
	c.SetInnerHashListener(listener)
	for _, n := range inner[0] {
		data := H256{uint32(n)}
		c.Write(data.toBytes())
	}
	c.Sum(nil)

	for l, array := range inner {
		for i, n := range array {
			if n < 1000 {
				t.Fatalf("Level:%d, Node:%d was not heard", l, i)
			}
		}
	}
}
