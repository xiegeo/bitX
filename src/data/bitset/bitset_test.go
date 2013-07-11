package bitset

import (
	"testing"
)

func checkEmpty(t *testing.T, s BitSet, max int) {
	for i := 0; i < max; i++ {
		if s.Get(i) {
			t.Errorf("empty set contains %v", i)
		}
	}
}

func checkSet(t *testing.T, s BitSet, max int) {
	for i := 0; i < max; i++ {
		s.Set(i)
		for j := 0; j <= i; j++ {
			if !s.Get(j) {
				t.Errorf("set of %v didn't work", i)
			}
		}
		for j := i + 1; j < max; j++ {
			if s.Get(j) {
				t.Errorf("set of %v caused %v to be set", i, j)
			}
		}
	}
}

func checkUnset(t *testing.T, s BitSet, max int) {
	for i := 0; i < max; i++ {
		s.Unset(i)
		for j := 0; j <= i; j++ {
			if s.Get(j) {
				t.Errorf("unset of %v didn't work", i)
			}
		}
		for j := i + 1; j < max; j++ {
			if !s.Get(j) {
				t.Errorf("unset of %v caused %v to be cleared", i, j)
			}
		}
	}
}

func checkAll(t *testing.T, s BitSet, max int) {
	checkEmpty(t, s, max)
	checkSet(t, s, max)
	checkUnset(t, s, max)
}

func testSimple(cap int, t *testing.T) {
	s := NewSimple(cap)
	checkAll(t, s, cap)
}

func TestSet31(t *testing.T) { testSimple(31, t) }
func TestSet32(t *testing.T) { testSimple(32, t) }
func TestSet33(t *testing.T) { testSimple(33, t) }
func TestSet63(t *testing.T) { testSimple(63, t) }
func TestSet64(t *testing.T) { testSimple(64, t) }
func TestSet65(t *testing.T) { testSimple(65, t) }
