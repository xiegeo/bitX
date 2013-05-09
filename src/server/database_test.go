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

func testFileSize(len hashtree.Bytes, t *testing.T) {
	t.Log("testing size:", len)
	d := OpenSimpleDatabase(testDatabase, testLevelLow)

	id := d.ImportFromReader(&testFile{length: len})
	if hashtree.Bytes(id.GetLength()) != len {
		t.Fatalf("Length is %x, should be %x", id.GetLength(), len)
	}

	buf := make([]byte, 1024)
	n := 0
	for i := 0; i < int(len); i += n {
		n, _ = d.GetAt(buf, id, int64(i))
		for j := 0; j < n; j++ {
			if buf[j] != testFileG(i+j) {
				t.Fatalf("at:%d, got:%x, expected:%x, for file:%s", i+j, buf[j], testFileG(i+j), id.CompactId())
			}
		}
	}

	hash := hashtree.NewTree()
	leafs := hashtree.Nodes((len + hashtree.FILE_BLOCK_SIZE - 1) / hashtree.FILE_BLOCK_SIZE)
	Levels := hash.Levels(leafs)
	for i := hashtree.Level(testLevelLow); i < Levels; i++ {
		inner, _ := d.GetInnerHashes(id, network.InnerHashes{
			Height: int32p(testLevelLow),
			From:   int32p(0),
			Length: int32p(int(hash.LevelWidth(leafs, i))),
		})
		list := inner.GetHashes()
		hash.Write(list)
		listSum := hash.Sum(nil)
		if !bytes.Equal(listSum, id.Hash) {
			t.Fatalf("At level:%d , inner hashes:%x, sums to:%x, expected:%x", i, list, listSum, id.Hash)
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
