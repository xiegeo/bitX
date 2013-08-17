package bitset

import (
	"os"
)

const (
	fileBlockSize = 4096 //byte rage of changes packed into one file write operation
	fileBlockBits = fileBlockSize * 8
)

/*
FileBackedBitSet assumes that OS intelligently cache file reads,
and flushes on demand. BitSets are nomally used as metadata, so
it should only flash after the main data flushes.

Currently, it appears that file read can be an expansive operation,
cached or not. A solution is pushed back until usage pattern is
more apparent.
*/
type FileBackedBitSet struct {
	f       *os.File
	c       int
	changes map[int]map[int]bool //[block number][bit number in block] = set as
}

func OpenFileBacked(fileName string, capacity int) *FileBackedBitSet {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	return NewFileBacked(file, capacity)
}

/*
If file size match capacity, use the file as bitmap.
Otherwise create the file sized to capacity and zero filled.
*/
func NewFileBacked(f *os.File, capacity int) *FileBackedBitSet {
	fi, err := f.Stat()
	if err != nil {
		panic(err)
	}
	size := int64(capacity+8-1) / 8
	fileSize := fi.Size()
	if fileSize > size {
		panic("unexpected: file to big") //f.Truncate(0)
	}
	if fileSize < size {
		_, err := f.WriteAt(make([]byte, size), 0)
		if err != nil {
			panic(err)
		}
	}
	return &FileBackedBitSet{f, capacity, make(map[int]map[int]bool)}
}

func (b *FileBackedBitSet) locateChange(key int) (bucket int, bit int) {
	checkIndex(key, b.c)
	bucket = key / fileBlockBits
	bit = key - (bucket * fileBlockBits)
	return
}

func (b *FileBackedBitSet) locateByteMask(key int) (bucket int, mask byte) {
	checkIndex(key, b.c)
	bucket = key / 8
	mask = 1 << byte(key%8)
	return
}

func (b *FileBackedBitSet) setAs(i int, v bool) {
	bucket, bit := b.locateChange(i)
	bmap, ok := b.changes[bucket]
	if !ok {
		bmap = make(map[int]bool)
		b.changes[bucket] = bmap
	}
	bmap[bit] = v
}

func (b *FileBackedBitSet) Set(i int) {
	b.setAs(i, true)
}

func (b *FileBackedBitSet) Unset(i int) {
	b.setAs(i, false)
}

func (b *FileBackedBitSet) Get(i int) bool {
	bucket, bit := b.locateChange(i)
	v, ok := b.changes[bucket][bit]
	if ok {
		return v
	}
	oneByte := make([]byte, 1)
	maskBucket, mask := b.locateByteMask(i)
	_, err := b.f.ReadAt(oneByte, int64(maskBucket))
	if err != nil {
		panic(err)
	}
	return (oneByte[0] & mask) != 0
}

func (b *FileBackedBitSet) Capacity() int { return b.c }

// Flush changes and Close the file, BitSet must not be used again
func (b *FileBackedBitSet) Close() {
	b.Flush()
	err := b.f.Close()
	if err != nil {
		panic(err)
	}
	//nil all internal pointers
	b.f = nil
	b.changes = nil
	b.c = -1
}

func (b *FileBackedBitSet) Flush() {
	if len(b.changes) == 0 {
		return
	}
	buffer := make([]byte, fileBlockSize)
	buckets := (b.c + fileBlockBits - 1) / fileBlockBits
	for i := 0; i < buckets; i++ {
		bmap := b.changes[i]
		delete(b.changes, i)
		if len(bmap) > 0 {
			starts := i * fileBlockSize
			if i == buckets-1 {
				bufferSize := (b.c+8-1)/8 - starts
				buffer = make([]byte, bufferSize)
			}
			_, err := b.f.ReadAt(buffer, int64(starts))
			if err != nil {
				panic(err)
			}
			for k, v := range bmap {
				maskBucket, mask := b.locateByteMask(k)
				if v {
					buffer[maskBucket] |= mask
				} else {
					buffer[maskBucket] &^= mask
				}
			}
			_, err = b.f.WriteAt(buffer, int64(starts))
			if err != nil {
				panic(err)
			}
		}
	}
}
