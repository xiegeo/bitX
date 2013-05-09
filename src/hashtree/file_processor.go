package hashtree

import (
	"crypto/sha256"
	"hash"
)

const (
	FILE_BLOCK_SIZE = 1024
)

// fileDigest represents the partial evaluation of a file hash.
type fileDigest struct {
	len           Bytes            // processed length
	leaf          hash.Hash        // a hash, used for hashing leaf nodes
	leafBlockSize int              // size of base block in bytes
	tree          CopyableHashTree // the digest used for inner and root nodes
}

// Create the standard file tree hash using leaf blocks of 1kB and sha256,
// and inner hash using sha256 without padding.  
func NewFile() HashTree {
	return NewFile2(FILE_BLOCK_SIZE, sha256.New(), NewTree2(NoPad32bytes, ht_sha256block))
}

// Create any tree hash using leaf blocks of size and leaf hash,
// and inner hash using tree hash, the tree stucture is internal to the tree hash.  
func NewFile2(leafBlockSize int, leaf hash.Hash, tree CopyableHashTree) HashTree {
	d := new(fileDigest)
	d.len = 0
	d.leafBlockSize = leafBlockSize
	d.leaf = leaf
	d.tree = tree
	return d
}

func (d *fileDigest) Nodes(len Bytes) Nodes {
	b := Bytes(d.BlockSize())
	return Nodes((len + b - 1) / b)
}

func (d *fileDigest) Levels(n Nodes) Level {
	return d.tree.Levels(n)
}

func (d *fileDigest) LevelWidth(n Nodes, level Level) Nodes {
	return d.tree.LevelWidth(n, level)
}

func (d *fileDigest) SetInnerHashListener(l func(level Level, index Nodes, hash *H256)) {
	d.tree.SetInnerHashListener(l)
}

func (d *fileDigest) Size() int { return d.tree.Size() }

func (d *fileDigest) BlockSize() int { return d.leafBlockSize }

func (d *fileDigest) Reset() {
	d.tree.Reset()
	d.leaf.Reset()
	d.len = 0
}
func (d *fileDigest) Write(p []byte) (int, error) {
	startLength := len(p)
	xn := int(d.len) % d.leafBlockSize
	for len(p)+xn >= d.leafBlockSize {
		writeLength := d.leafBlockSize - xn
		d.leaf.Write(p[0:writeLength])
		p = p[writeLength:]
		d.tree.Write(d.leaf.Sum(nil))
		d.leaf.Reset()
		xn = 0
	}
	if len(p) > 0 {
		d.leaf.Write(p)
	}
	d.len += Bytes(startLength)
	return startLength, nil
}

func (d *fileDigest) Sum(in []byte) []byte {
	if int(d.len)%d.leafBlockSize != 0 || d.len == 0 {
		// Make a copy of d.tree so that caller can keep writing and summing.
		tree := d.tree.Copy()
		tree.Write(d.leaf.Sum(nil))
		return tree.Sum(in)
	}
	return d.tree.Sum(in)
}
