/*
a bitset with code first taken from math.big, and github.com/phf/go-intset/
*/
package bitset

import (
	"fmt"
)

const (
	/*
		Always checks the index.
		If false and index is outside of capacity, then behaviour is undefined.
	*/
	CHECK_INTEX = true
)

type GetBitSet interface {
	Get(k int) bool
	Capacity() int
}

type PutBitSet interface {
	Set(k int)
	Unset(k int)
}

type BitSet interface {
	GetBitSet
	PutBitSet
}

type FlushableBitSet interface {
	BitSet
	Flush()
}

type Word uintptr

type SimpleBitSet struct {
	d []Word
	c int
}

const (
	_m    = ^Word(0)
	_logS = _m>>8&1 + _m>>16&1 + _m>>32&1
	_S    = 1 << _logS // word size in bytes

	_W = _S << 3 // word size in bits
)

func checkIndex(key int, cap int) {
	if CHECK_INTEX && (key < 0 || key >= cap) {
		panic(fmt.Errorf("bitset: index %v outside of range 0 to %v", key, cap-1))
	}
}

func (s *SimpleBitSet) locate(key int) (bucket int, mask Word) {
	checkIndex(key, s.c)
	bucket = key / _W
	mask = 1 << Word(key%_W)
	return
}

func NewSimple(capacity int) *SimpleBitSet {
	return &SimpleBitSet{make([]Word, (capacity+_W-1)/_W), capacity}
}

func NewSimpleFromWords(capacity int, w []Word) *SimpleBitSet {
	if len(w) != (capacity+_W-1)/_W {
		panic("capacity and data length mismatch.")
	}
	return &SimpleBitSet{w, capacity}
}

func NewSimpleFromBytes(capacity int, d []byte) *SimpleBitSet {
	return NewSimpleFromWords(capacity, BytesToWords(d))
}

func BytesToWords(buf []byte) (w []Word) {
	//from math/big/nat.go -> setBytes
	w = make([]Word, (len(buf)+_S-1)/_S)

	k := 0
	s := uint(0)
	var d Word
	for i := len(buf); i > 0; i-- {
		d |= Word(buf[i-1]) << s
		if s += 8; s == _S*8 {
			w[k] = d
			k++
			s = 0
			d = 0
		}
	}
	if k < len(w) {
		w[k] = d
	}

	return w
}

func (s *SimpleBitSet) Set(i int) {
	bucket, mask := s.locate(i)
	s.d[bucket] |= mask
}

func (s *SimpleBitSet) Unset(i int) {
	bucket, mask := s.locate(i)
	s.d[bucket] &^= mask
}

func (s *SimpleBitSet) Get(i int) (b bool) {
	bucket, mask := s.locate(i)
	return (s.d[bucket] & mask) != 0
}

func (s *SimpleBitSet) Capacity() int { return s.c }
