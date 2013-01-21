package hashtree

import (
	"crypto/sha256"
	"hash"
)

// fileDigest represents the partial evaluation of a file hash.
type fileDigest struct {
	len           uint64     // processed length
	leaf          hash.Hash  // a hash, used for hashing leaf nodes
	leafBlockSize int        // size of base block in bytes
	tree          treeDigest // the digest used for inner and root nodes
}

func NewFile() hash.Hash {
	return NewFile2(1024, sha256.New(), NewTree2(NoPad32bytes, ht_sha256block).(*treeDigest))
}

func NewFile2(leafBlockSize int, leaf hash.Hash, tree *treeDigest) hash.Hash {
	d := new(fileDigest)
	d.len = 0
	d.leafBlockSize = leafBlockSize
	d.leaf = leaf
	d.tree = *tree
	return d
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
	d.len += uint64(startLength)
	return startLength, nil
}

func (d0 *fileDigest) Sum(in []byte) []byte {
	if int(d0.len)%d0.leafBlockSize != 0 || d0.len == 0 {
		// Make a copy of d0.tree so that caller can keep writing and summing.
		tree := d0.tree
		tree.Write(d0.leaf.Sum(nil))
		return tree.Sum(in)
	}
	return d0.tree.Sum(in)
}
