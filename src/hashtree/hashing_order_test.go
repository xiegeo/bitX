package hashtree

import (
	"testing"
)

// Test the order and structure of the tree by using minus as the hash operation
// 0 -> 0
// 0 - 1 -> -1
// 0 - 1 - 2 -> -3
// (0-1) - (2-3) -> 0
// Whether the i'th node is + or - depends on the whether i has even or odd number of 1's in base 2
func TestOrder(t *testing.T) {
	c := NewTree2(NoPad32bytes, minus).(*digest)
	expect := int32(0)
	for i := int32(0); i < 100; i++ {
		n := i // n is the value of the i'th input, any function of i should pass test
		data := h256{uint32(n)}
		c.Write(data.toBytes())
		ans := int32(fromBytes(c.Sum(nil))[0])
		if evenBits(uint32(i)) {
			expect += n
		} else {
			expect -= n
		}
		if ans != expect {
			t.Fatalf("%v,%v> expect:%v != got:%v", i, n, expect, ans)
		}
	}
}

func evenBits(n uint32) bool {
	count := uint32(0)
	for n != 0 {
		count += n & 1
		n >>= 1
	}
	return count%2 == 0
}

func minus(left *h256, right *h256) *h256 {
	l := left[0]
	r := right[0]
	h := l - r
	return &h256{uint32(h)}
}
