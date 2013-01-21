package hashtree

import (
	"fmt"
	"hash"
	"io"
)

const treeNodeSize = 32

type h256 [8]uint32 //the internal hash

//bytes must have a length of 32
func fromBytes(bytes []byte) *h256 {
	var h h256
	for i := 0; i < 8; i++ {
		j := i * 4
		h[i] = uint32(bytes[j])<<24 | uint32(bytes[j+1])<<16 | uint32(bytes[j+2])<<8 | uint32(bytes[j+3])
	}
	return &h
}

func (h *h256) toBytes() []byte {
	bytes := make([]byte, 32)
	for i, s := range h {
		bytes[i*4] = byte(s >> 24)
		bytes[i*4+1] = byte(s >> 16)
		bytes[i*4+2] = byte(s >> 8)
		bytes[i*4+3] = byte(s)
	}
	return bytes
}

// treeDigest represents the partial evaluation of a hashtree.
type treeDigest struct {
	x          [treeNodeSize]byte            // unprocessed bytes
	xn         int                           //length of x
	len        uint64                        // processed length
	stack      [64]*h256                     // partial hashtree of more height then ever needed
	sn         int                           // top of stack, depth of tree
	padder     func(d io.Writer, len uint64) //the padding function
	compressor func(l, r *h256) *h256        //512 to 256 hash function
}

func NewTree() hash.Hash {
	return NewTree2(ZeroPad32bytes, ht_sha256block)
}

func NewTree2(padder func(d io.Writer, len uint64), compressor func(l, r *h256) *h256) hash.Hash {
	d := new(treeDigest)
	d.Reset()
	d.padder = padder
	d.compressor = compressor
	return d
}
func (d *treeDigest) Size() int { return 32 }

func (d *treeDigest) BlockSize() int { return treeNodeSize }

func (d *treeDigest) Reset() {
	d.xn = 0
	d.len = 0
	d.stack = [64]*h256{nil}
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
func (d *treeDigest) writeStack(node *h256, level int) {
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

	var right *h256
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