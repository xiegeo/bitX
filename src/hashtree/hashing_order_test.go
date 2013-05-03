package hashtree

import (
	"bytes"
	"testing"
)

// Test the order and structure of the tree by using minus as the hash operation
// 0 -> 0
// 0 - 1 -> -1
// 0 - 1 - 2 -> -3
// (0-1) - (2-3) -> 0
// Whether the i'th node is + or - depends on the whether i has even or odd number of 1's in base 2
func TestTreeOrder(t *testing.T) {
	c := NewTree2(NoPad32bytes, minus).(*treeDigest)
	expect := int32(0)
	for i := int32(0); i < 100; i++ {
		n := i // n is the value of the i'th input, any function of i should pass test
		data := H256{uint32(n)}
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

func minus(left *H256, right *H256) *H256 {
	l := left[0]
	r := right[0]
	h := l - r
	return &H256{uint32(h)}
}

// Test the order and structure of the file processor by making it duplicate tree processor
func TestFileOrder(t *testing.T) {
	fileSize := treeNodeSize*5 + 1
	tree := NewTree().(*treeDigest)
	t1 := *tree
	t2 := *tree
	file := NewFile2(treeNodeSize, &t1, &t2)
	buf := make([]byte, fileSize)
	tree.Write(buf)
	tsum := tree.Sum(nil)
	file.Write(buf)
	fsum := file.Sum(nil)
	if !bytes.Equal(fsum, tsum) {
		t.Fatalf(" %x != %x", fsum, tsum)
	}
}
