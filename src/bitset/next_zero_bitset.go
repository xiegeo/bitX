package bitset

// a bitset for list the index of 0s
type NextZeroBitSet struct {
	SimpleBitSet
	pos int
}

// the next 0 bit, indexed from 0 to cap-1, index is undefined if done
func (n *NextZeroBitSet) Next() (index int, done bool) {
	for ; n.pos < n.Capacity(); n.pos++ {
		if n.Get(n.pos) {
			return n.pos, false
		}
		// todo: skip words with all 1s
	}
	return -1, true
}
