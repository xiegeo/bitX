package bitset

import (
	"fmt"
	"os"
	"testing"
)

func testFileBacked(cap int, t *testing.T) {
	f, err := os.Create(fmt.Sprintf(".testfile_%v", cap))
	if err != nil {
		t.Fatal(err)
	}
	s := NewFileBacked(f, cap)
	checkAll(t, s, cap)
	if s.Capacity() != cap {
		t.Fatalf("capacity should be %v but returns %v", cap, s.Capacity())
	}
	if CHECK_INTEX {
		tryOutSide(s, -1, t)
		tryOutSide(s, cap, t)
	}
}

func TestFileSet0(t *testing.T)   { testFileBacked(0, t) }
func TestFileSet1(t *testing.T)   { testFileBacked(1, t) }
func TestFileSet2(t *testing.T)   { testFileBacked(2, t) }
func TestFileSet7(t *testing.T)   { testFileBacked(7, t) }
func TestFileSet8(t *testing.T)   { testFileBacked(8, t) }
func TestFileSet9(t *testing.T)   { testFileBacked(9, t) }
func TestFileSet127(t *testing.T) { testFileBacked(127, t) }
func TestFileSet128(t *testing.T) { testFileBacked(128, t) }
func TestFileSet129(t *testing.T) { testFileBacked(129, t) }
