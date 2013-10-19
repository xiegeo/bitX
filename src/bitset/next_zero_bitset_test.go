package bitset

import (
	"testing"
)

func TestRange(t *testing.T) {

	n := NewNextZeroBitSet(NewSimple(100))
	starts := []int{0, 2, 20, 98}
	lengths := []int{1, 5, 30, 2}
	maxRange := 10
	rstarts := []int{0, 2, 20, 30, 40, 98, -1}
	rlengths := []int{1, 5, 10, 10, 10, 2, 0}
	for i := 0; i < n.Capacity(); i++ {
		n.Set(i)
	}
	for k, s := range starts {
		l := lengths[k]
		for i := 0; i < l; i++ {
			n.Unset(s + i)
		}
	}

	for k, s := range rstarts {
		l := rlengths[k]
		start, length := n.NextRange(maxRange)
		if s != start || l != length {
			t.Errorf("expect:%v, %v got:%v, %v ", s, l, start, length)
		}
	}

}
