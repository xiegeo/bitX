package hashtree

import (
	"testing"
)

type treeLevels struct {
	len    uint64
	levels int
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
	len   uint64
	level int
	width int
}

var expectedLevelWidth = []levelWidth{
	{0, 1, 1}, {1, 1, 1}, {treeNodeSize, 1, 1},

	{treeNodeSize + 1, 1, 2}, {2 * treeNodeSize, 1, 2},
	{treeNodeSize + 1, 2, 1}, {2 * treeNodeSize, 2, 1},

	{2*treeNodeSize + 1, 1, 3}, {4 * treeNodeSize, 1, 4},
	{2*treeNodeSize + 1, 2, 2}, {4 * treeNodeSize, 2, 2},
	{2*treeNodeSize + 1, 3, 1}, {4 * treeNodeSize, 3, 1},

	{5 * treeNodeSize, 1, 5}, {6 * treeNodeSize, 1, 6}, {7 * treeNodeSize, 1, 7}, {8 * treeNodeSize, 1, 8},
	{5 * treeNodeSize, 2, 3}, {6 * treeNodeSize, 2, 3}, {7 * treeNodeSize, 2, 4}, {8 * treeNodeSize, 2, 4},
	{5 * treeNodeSize, 3, 2}, {6 * treeNodeSize, 3, 2}, {7 * treeNodeSize, 3, 2}, {8 * treeNodeSize, 3, 2},
	{5 * treeNodeSize, 4, 1}, {6 * treeNodeSize, 4, 1}, {7 * treeNodeSize, 4, 1}, {8 * treeNodeSize, 4, 1},
}

func TestLevelWidth(t *testing.T) {
	h := NewTree()
	for i := 0; i < len(expectedLevelWidth); i++ {
		e := expectedLevelWidth[i]
		if h.LevelWidth(e.len, e.level) != e.width {
			t.Fatalf("LevelWidth(%d,%d) = %d want %d", e.len, e.level, h.LevelWidth(e.len, e.level), e.level)
		}
	}
}
