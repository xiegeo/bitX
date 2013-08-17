package network

import (
	"../hashtree"
	//"fmt"
	"testing"
)

type slsTest struct {
	from   hashtree.Nodes
	to     hashtree.Nodes
	width  hashtree.Nodes
	expect [][2]hashtree.Nodes
}

var testdata = []slsTest{
	{0, 0, 1, [][2]hashtree.Nodes{{0, 0}}},
	{0, 0, 2, nil},
	{0, 1, 2, [][2]hashtree.Nodes{{0, 1}}},
	{0, 2, 3, [][2]hashtree.Nodes{{0, 2}}},
	{0, 2, 4, [][2]hashtree.Nodes{{0, 1}}},
	{1, 2, 3, [][2]hashtree.Nodes{{2, 2}}},
	{1, 2, 4, nil},
	{0, 3, 4, [][2]hashtree.Nodes{{0, 3}}},
	{1, 6, 8, [][2]hashtree.Nodes{{2, 3}, {4, 5}}},
	{1, 9, 10, [][2]hashtree.Nodes{{2, 3}, {4, 7}, {8, 9}}},
	{2, 5, 10, [][2]hashtree.Nodes{{2, 3}, {4, 5}}},
	{6, 9, 10, [][2]hashtree.Nodes{{6, 7}, {8, 9}}},
	{10, 18, 20, [][2]hashtree.Nodes{{10, 11}, {12, 15}, {16, 17}}},
	{1, 4, 5, [][2]hashtree.Nodes{{2, 3}, {4, 4}}},
	{1, 4, 6, [][2]hashtree.Nodes{{2, 3}}},
}

func TestSplitLocalSummable(t *testing.T) {
	for _, v := range testdata {
		sls := sls(v.from, v.to, v.width)
		if len(sls) != len(v.expect) {
			t.Errorf("%v got %v", v, sls)
		} else {
			for i := 0; i < len(sls); i++ {
				exp := v.expect[i]
				got := sls[i]
				if exp[0] != got[0] || exp[1] != got[1] {
					t.Errorf("part %v, got %v != exp %v, for test: %v", i, got, exp, v)
				}
			}
		}
	}
}

type lrTest struct {
	leafs   hashtree.Nodes
	height  hashtree.Level
	from    hashtree.Nodes
	length  hashtree.Nodes
	e_level hashtree.Level
	e_node  hashtree.Nodes
}

var lrData = []lrTest{
	{1, 0, 0, 1, 0, 0},
	{2, 1, 0, 1, 1, 0},
	{2, 0, 1, 1, 0, 1},
	{3, 0, 2, 1, 1, 1}, {4, 0, 2, 1, 0, 2},

	{99, 0, 0, 2, 1, 0},
	{99, 0, 0, 3, 2, 0},
	{99, 1, 0, 4, 3, 0},
	{99, 1, 0, 5, 4, 0},

	{999999, 10, 8, 2, 11, 4},
	{999999, 10, 8, 3, 12, 2},
	{999999, 10, 8, 4, 12, 2},
	{999999, 10, 8, 5, 13, 1},
}

func TestLocalRoot(t *testing.T) {
	for _, v := range lrData {
		inner := NewInnerHashes(v.height, v.from, v.length, nil)
		l, n := inner.LocalRoot(v.leafs)
		if l != v.e_level || n != v.e_node {
			t.Errorf("test:%v, got level:%v, and node:%v", v, l, n)
		}
	}
}
