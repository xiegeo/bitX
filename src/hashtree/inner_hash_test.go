package hashtree

import (
	"testing"
)

type treeLevels struct {
	n      Nodes
	levels Level
}

var expectedTreeLevels = []treeLevels{
	{1, 1},
	{2, 2},
	{3, 3}, {4, 3},
	{5, 4}, {8, 4},
	{9, 5}, {16, 5},
	{17, 6},
}

func TestTreeLevels(t *testing.T) {
	h := NewTree()
	for i := 0; i < len(expectedTreeLevels); i++ {
		e := expectedTreeLevels[i]
		if h.Levels(e.n) != e.levels {
			t.Fatalf("Levels(%d) = %d want %d", e.n, h.Levels(e.n), e.levels)
		}
	}
}

type levelWidth struct {
	n     Nodes
	level Level
	width Nodes
}

var expectedLevelWidth = []levelWidth{
	{1, 0, 1},

	{2, 0, 2},
	{2, 1, 1},

	{3, 0, 3}, {4, 0, 4},
	{3, 1, 2}, {4, 1, 2},
	{3, 2, 1}, {4, 2, 1},

	{5, 0, 5}, {6, 0, 6}, {7, 0, 7}, {8, 0, 8},
	{5, 1, 3}, {6, 1, 3}, {7, 1, 4}, {8, 1, 4},
	{5, 2, 2}, {6, 2, 2}, {7, 2, 2}, {8, 2, 2},
	{5, 3, 1}, {6, 3, 1}, {7, 3, 1}, {8, 3, 1},
}

func TestLevelWidth(t *testing.T) {
	h := NewTree()
	for i := 0; i < len(expectedLevelWidth); i++ {
		e := expectedLevelWidth[i]
		if h.LevelWidth(e.n, e.level) != e.width {
			t.Fatalf("LevelWidth(%d,%d) = %d want %d", e.n, e.level, h.LevelWidth(e.n, e.level), e.width)
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
		}
		inner[l][i] += 2000 //mark heard
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
