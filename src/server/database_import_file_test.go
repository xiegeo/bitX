package server

import (
	"../data/network"
	"../hashtree"
	"bytes"
	"io"
	"os"
	"testing"
)

const (
	testDatabase = ".testDatabase"
	testLevelLow = 0
)

func TestFileIO(t *testing.T) {
	err := os.RemoveAll(testDatabase)
	if err != nil {
		t.Fatal(err)
	}
	testFileSize(0, t)
	testFileSize(1, t)
	testFileSize(1024, t)
	testFileSize(1025, t)
	testFileSize(2345, t)
	testFileSize(12345, t)
}

func testFileSize(size hashtree.Bytes, t *testing.T) {
	t.Log("testing size:", size)
	d := OpenSimpleDatabase(testDatabase, testLevelLow)

	id := d.ImportFromReader(&testFile{length: size})
	if hashtree.Bytes(id.GetLength()) != size {
		t.Fatalf("Length is %x, should be %x", id.GetLength(), size)
	}

	buf := make([]byte, 1024)
	n := hashtree.Bytes(0)
	for i := hashtree.Bytes(0); i < size; i += n {
		bufLen, _ := d.GetAt(buf, id, i)
		n = hashtree.Bytes(bufLen)
		for j := hashtree.Bytes(0); j < n; j++ {
			if buf[j] != testFileG(i+j) {
				t.Fatalf("at:%d, got:%x, expected:%x, for file:%s", i+j, buf[j], testFileG(i+j), id.CompactId())
			}
		}
	}

	hash := hashtree.NewTree()
	leafs := refHash.Nodes(size)
	t.Log("leafs:", leafs)
	levels := hash.Levels(leafs)
	for i := hashtree.Level(testLevelLow); i < levels-1; i++ {
		req := network.InnerHashes{
			Height: int32p(int(i)),
			From:   int32p(0),
			Length: int32p(int(hash.LevelWidth(leafs, i))),
		}
		got, _ := d.GetInnerHashes(id, req)
		list := got.GetHashes()
		hash.Write(list)
		listSum := hash.Sum(nil)
		hash.Reset()
		if !bytes.Equal(listSum, id.Hash) {
			t.Fatalf("Req:%s , got hashes:%x, len:%d, sums to:%x, expected:%x", req.String(), list, len(list), listSum, id.Hash)
		}
	}
}

type testFile struct {
	index  hashtree.Bytes
	length hashtree.Bytes
}

func (f *testFile) Read(b []byte) (int, error) {
	i := 0
	for ; i < len(b); i++ {
		if f.index == f.length {
			return i, io.EOF
		}
		b[i] = testFileG(f.index)
		f.index++
	}

	return i, nil
}

func testFileG(index hashtree.Bytes) byte {
	return byte(index * index / 10007)
}

func int32p(n int) *int32 {
	m := int32(n)
	return &m
}
