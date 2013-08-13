package bitset

import (
	"os"
)

const (
	fileBlockSize = 4096 //byte rage of changes packed into one file write operation
	fileBlockBites = fileBlockSize * 8
)

/*
FileBackedBitSet assumes that OS intelligently cache file reads,
and flashes on demand. BitSets are nomally used as metadata, so
it should only flash after the main data flashes.
*/
type FileBackedBitSet struct {
	f *os.File
	c int
	changes map[int]map[int]bool //[block number][bit number in block] = set as
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
	size := int64(capacity+8-1)/8
	fileSize := fi.Size()
	if fileSize > size {
		panic("unexpected: file to big")//f.Truncate(0)
	}
	if fileSize < size {
		_, err := f.WriteAt(make([]byte,size),0)
		if err != nil {
			panic(err)
		}
	}
	return &FileBackedBitSet{f, capacity, nil}
}