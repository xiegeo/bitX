package bitset

import (
	"bytes"
	"encoding/binary"
	_ "os"
)

const (
	COUNT_BITS  = 64
	COUNT_BYTES = COUNT_BITS / 8
)

type CountingFileBackedBitSet struct {
	FileBackedBitSet
	count int64
}

func (c *CountingFileBackedBitSet) readCount() int64 {
	count := int64(0)
	buf := make([]byte, COUNT_BYTES)
	_, err := c.f.ReadAt(buf, int64(c.Capacity()+7)/8)
	if err != nil {
		panic(err)
	}
	err = binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &count)
	if err != nil {
		panic(err)
	}
	return count
}

func (c *CountingFileBackedBitSet) writeCount(count int64) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, count)
	if err != nil {
		panic(err)
	}
	_, err = c.f.WriteAt(buf.Bytes(), int64(c.Capacity()+7)/8)
	if err != nil {
		panic(err)
	}
}

func (c *CountingFileBackedBitSet) recount() int64 {
	if len(c.changes) != 0 {
		panic("recount can not be used with changes")
	}
	count := int64(0)
	buffer := make([]byte, fileBlockSize)
	buckets := (c.Capacity() + fileBlockBits - 1) / fileBlockBits
	for i := 0; i < buckets; i++ {
		starts := i * fileBlockSize
		if i == buckets-1 {
			bufferSize := (c.Capacity()+8-1)/8 - starts
			buffer = make([]byte, bufferSize)
		}
		_, err := c.f.ReadAt(buffer, int64(starts))
		if err != nil {
			panic(err)
		}
		for _, b := range buffer {
			for i := uint(0); i < 8; i++ {
				if b&(1<<i) != 0 {
					count++
				}
			}
		}
	}
	return count
}

func OpenCountingFileBacked(fileName string, capacity int) *CountingFileBackedBitSet {
	fileBacked := OpenFileBacked(fileName, capacity+COUNT_BITS)
	counting := &CountingFileBackedBitSet{*fileBacked, 0}
	count := counting.readCount()
	if count == -1 {
		count = counting.recount()
	}
	counting.count = count
	return counting
}

func (c *CountingFileBackedBitSet) Capacity() int { return c.c - COUNT_BITS }

func (c *CountingFileBackedBitSet) Count() int {
	return int(c.count)
}

func (c *CountingFileBackedBitSet) Full() bool {
	return c.Count() == c.Capacity()
}

func (c *CountingFileBackedBitSet) Set(i int) {
	if c.Get(i) {
		return
	} else {
		c.count++
		c.FileBackedBitSet.Set(i)
		if CHECK_INTEX && int(c.count) > c.Capacity() {
			panic("count > c.Capacity()")
		}
	}
}

func (c *CountingFileBackedBitSet) Unset(i int) {
	if !c.Get(i) {
		return
	} else {
		c.count--
		c.FileBackedBitSet.Unset(i)
		if CHECK_INTEX && c.count < 0 {
			panic("count < 0")
		}
	}
}

func (c *CountingFileBackedBitSet) Flush() {
	if len(c.changes) == 0 {
		return
	}
	c.writeCount(-1) // mark count as dirty
	c.FileBackedBitSet.Flush()
	c.writeCount(c.count)
}
func (c *CountingFileBackedBitSet) Close() {
	c.Flush()
	c.FileBackedBitSet.Close()
}

func (c *CountingFileBackedBitSet) DataByteLength() int64 {
	return int64(c.Capacity()+8-1) / 8
}

func (c *CountingFileBackedBitSet) ExportBytes() []byte {
	return c.FileBackedBitSet.ExportBytes()[:c.DataByteLength()]
}

func (c *CountingFileBackedBitSet) ToSimple() *SimpleBitSet {
	return NewSimpleFromBytes(c.Capacity(), c.ExportBytes())
}
