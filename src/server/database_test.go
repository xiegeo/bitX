package server

import (
	"io"
	"testing"
)

const (
	testFileLength = 1234
)

func TestImport(t *testing.T) {
	d := OpenSimpleDatabase(".testDatabase")
	id := d.ImportFromReader(&testFile{length: testFileLength})
	if id.GetLength() != testFileLength {
		t.Fatalf("Length is %x, should be %x", id.GetLength(), testFileLength)
	}
	for i := int64(0); i < testFileLength; i++ {
		got, _ := d.Get(id, i, 1)
		if got[0] != testFileG(i) {
			t.Fatalf("at i:%x, got:%x, expected:%x, for file:%s", i, got, []byte{testFileG(i)}, id.CompactId())
		}
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
