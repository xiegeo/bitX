package bitset

// a bitset for list the index of 0s
type NextZeroBitSet struct {
	*SimpleBitSet
	pos int
}

func NewNextZeroBitSet(s *SimpleBitSet) *NextZeroBitSet {
	return &NextZeroBitSet{s, 0}
}

// the next 0 bit, indexed from 0 to cap-1, index is undefined if done
func (n *NextZeroBitSet) Next() (index int, done bool) {
	for ; n.pos < n.Capacity(); n.pos++ {
		if n.Get(n.pos) {
			n.pos++
			return n.pos - 1, false
		}
		// todo: skip words with all 1s
	}
	return -1, true
}

// the next 0 bits, length is the number of consecutive 0s including the first,
// done (see Next()) if length = 0
func (n *NextZeroBitSet) NextRange(maxRange int) (index int, length int) {
	index, done := n.Next()
	if done {
		return index, 0
	} else {
		r := 1
		i, d := n.Next()
		for r < maxRange && i == index+r && !d {
			r++
			i, d = n.Next()
		}
		if !d {
			n.pos--
		}

		return index, r
	}

}
