package server

import (
	"../hashtree"
	"../network"
	"bytes"
	"io"
	"testing"
)

const (
	testFileLength     = 1234
	testInnerHashLevel = 1
)

func TestImport(t *testing.T) {
	d := OpenSimpleDatabase(".testDatabase", testInnerHashLevel)

	id := d.ImportFromReader(&testFile{length: testFileLength})
	if id.GetLength() != testFileLength {
		t.Fatalf("Length is %x, should be %x", id.GetLength(), testFileLength)
	}

	buf := make([]byte, 1024)
	n := 0
	for i := int64(0); i < testFileLength; i += int64(n) {
		n, _ = d.GetAt(buf, id, i)
		for j := int64(0); j < int64(n); j++ {
			if buf[j] != testFileG(i+j) {
				t.Fatalf("at:%d, got:%x, expected:%x, for file:%s", i+j, buf[j], testFileG(i+j), id.CompactId())
			}
		}
	}

	inner, _ := d.GetInnerHashs(id, network.InnerHashs{
		Height: int32p(testInnerHashLevel),
		From:   int32p(0),
		Length: int32p(2),
	})
	list := inner.GetHashes()
	hash := hashtree.NewTree()
	hash.Write(list)
	listSum := hash.Sum(nil)
	if !bytes.Equal(listSum, id.Hash) {
		t.Fatalf("inner hashes:%x, sums to:%x, expected:%x", list, listSum, id.Hash)
	}
}

type testFile struct {
	index  int64
	length int64
}

func (f *testFile) Read(b []byte) (int, error) {
	if f.index == f.length {
		return 0, io.EOF
	}
	b[0] = testFileG(f.index)
	f.index++
	return 1, nil
}

func testFileG(index int64) byte {
	return byte(index)
}

func int32p(n int32) *int32 {
	return &n
}
