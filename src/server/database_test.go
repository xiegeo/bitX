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
	buf := make([]byte,1024)
	n := 0
	for i := int64(0); i < testFileLength; i +=int64(n) {
		n, _ = d.GetAt(buf, id, i)
		for j := int64(0); j < int64(n); j ++{
			if buf[j] != testFileG(i+j) {
				t.Fatalf("at:%d, got:%x, expected:%x, for file:%s", i+j, buf[j], testFileG(i+j), id.CompactId())
			}
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
