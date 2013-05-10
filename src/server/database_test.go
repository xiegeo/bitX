package server

import (
	"../hashtree"
	"../network"
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
	n := 0
	for i := 0; i < int(size); i += n {
		n, _ = d.GetAt(buf, id, hashtree.Bytes(i))
		for j := 0; j < n; j++ {
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
	if f.index == f.length {
		return 0, io.EOF
	}
	b[0] = testFileG(int(f.index))
	f.index++
	return 1, nil
}

func testFileG(index int) byte {
	return byte(index)
}

func int32p(n int) *int32 {
	m := int32(n)
	return &m
}
