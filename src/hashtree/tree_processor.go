package hashtree

import (
	"fmt"
	"hash"
	"io"
	"math"
)

const treeNodeSize = 32

type H256 [8]uint32 //the internal hash

//bytes must have a length of 32
func fromBytes(bytes []byte) *H256 {
	var h H256
	for i := 0; i < 8; i++ {
		j := i * 4
		h[i] = uint32(bytes[j])<<24 | uint32(bytes[j+1])<<16 | uint32(bytes[j+2])<<8 | uint32(bytes[j+3])
	}
	return &h
}

func (h *H256) toBytes() []byte {
	bytes := make([]byte, 32)
	for i, s := range h {
		bytes[i*4] = byte(s >> 24)
		bytes[i*4+1] = byte(s >> 16)
		bytes[i*4+2] = byte(s >> 8)
		bytes[i*4+3] = byte(s)
	}
	return bytes
}

type HashTree interface {
	hash.Hash
	Copy() HashTree
	Levels(len int64) int
	LevelWidth(len int64, level int) int
	SetInnerHashLissener(func(level int, index int, hash *H256))
}

// treeDigest represents the partial evaluation of a hashtree.
type treeDigest struct {
	x          [treeNodeSize]byte            // unprocessed bytes
	xn         int                           // length of x
	len        uint64                        // processed length
	stack      [64]*H256                     // partial hashtree of more height then ever needed
	sn         int                           // top of stack, depth of tree
	padder     func(d io.Writer, len uint64) // the padding function
	compressor func(l, r *H256) *H256        // 512 to 256 hash function
}

func NewTree() HashTree {
	return NewTree2(ZeroPad32bytes, ht_sha256block)
}

// Create a binary tree hash using padder and compressor.
// Padder mush pad to intervals of 256 bits.
// Compressor mush hash 2 H256s to 1.
func NewTree2(padder func(d io.Writer, len uint64), compressor func(l, r *H256) *H256) HashTree {
	d := new(treeDigest)
	d.Reset()
	d.padder = padder
	d.compressor = compressor
	return d
}
func (d *treeDigest) Copy() HashTree {
	d0 := *d
	return &d0
}

type countingWriter int

func (c *countingWriter) Write(p []byte) (length int, nil error) {
	length = len(p)
	*c += countingWriter(length)
	return
}

func (d *treeDigest) Levels(len int64) int {
	cw := countingWriter(0)
	d.padder(&cw, uint64(len))
	padded := len + int64(cw)
	return math.Ilogb(float64(padded/treeNodeSize*2-1)) + 1
}

func (d *treeDigest) LevelWidth(len int64, level int) int {
	return 0
}

func (d *treeDigest) SetInnerHashLissener(func(level int, index int, hash *H256)) {

}

func (d *treeDigest) Size() int { return 32 }

func (d *treeDigest) BlockSize() int { return treeNodeSize }

func (d *treeDigest) Reset() {
	d.xn = 0
	d.len = 0
	d.stack = [64]*H256{nil}
}
func (d *treeDigest) Write(p []byte) (startLength int, nil error) {
	startLength = len(p)
	for len(p)+d.xn >= treeNodeSize {
		for i := 0; i < treeNodeSize-d.xn; i++ {
			d.x[d.xn+i] = p[i]
		}
		p = p[treeNodeSize-d.xn:]
		d.xn = 0
		d.writeStack(fromBytes(d.x[:]), 0)
	}
	if len(p) > 0 {
		for i := 0; i < len(p); i++ {
			d.x[d.xn+i] = p[i]
		}
		d.xn += len(p)
	}
	d.len += uint64(startLength)
	return
}
func (d *treeDigest) writeStack(node *H256, level int) {
	if d.sn == level {
		d.stack[level] = node
		d.sn++
	} else if d.stack[level] == nil {
		d.stack[level] = node
	} else {
		last := d.stack[level]
		d.stack[level] = nil
		d.writeStack(d.compressor(last, node), level+1)
	}
}

func (d0 *treeDigest) Sum(in []byte) []byte {
	// Make a copy of d0 so that caller can keep writing and summing.
	d := *d0
	d.padder(&d, d.len)

	if d.xn != 0 {
		panic(fmt.Sprintf("d.xn = %d", d.xn))
	}

	var right *H256
	i := 0
	for ; right == nil; i++ {
		right = d.stack[i]
	}
	for ; i < d.sn; i++ {
		left := d.stack[i]
		if left != nil {
			right = d.compressor(left, right)
		}
	}

	return append(in, right.toBytes()...)
}

// to pad with 0 or more of bytes 0x00
func ZeroPad32bytes(d io.Writer, len uint64) {
	padSize := (32 - (len % 32)) % 32
	if len == 0 {
		padSize = 32
	}
	d.Write(make([]byte, padSize))
}

// use this when there should not need any padding, input is already in blocks, or non.
func NoPad32bytes(d io.Writer, len uint64) {
	if len%32 != 0 || len == 0 {
		panic(fmt.Sprintf("need padding of %v bytes for length of %v", 32-len%32, len))
	}
}
