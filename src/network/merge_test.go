package network

import (
	//"fmt"
	"testing"
)

func TestMergeFile(t *testing.T) {
	p := &Packet{}
	length := int64(1)
	f1 := &File{Id: &StaticId{Length: &length, Hash: []byte{1, 2, 3}}}
	f2 := &File{Id: &StaticId{Length: &length, Hash: []byte{1, 2, 3}}}
	f3 := &File{Id: &StaticId{Length: &length, Hash: []byte{1, 2, 4}}}

	if len(p.GetFiles()) != 0 {
		t.Errorf("packet %v should have 0 files", p)
	}
	p.MergeFile(f1)
	if len(p.GetFiles()) != 1 {
		t.Errorf("packet %v should have 1 files", p)
	}
	p.MergeFile(f2)
	if len(p.GetFiles()) != 1 {
		t.Errorf("packet %v should have 1 files", p)
	}
	p.MergeFile(f3)
	if len(p.GetFiles()) != 2 {
		t.Errorf("packet %v should have 2 files", p)
	}
}
