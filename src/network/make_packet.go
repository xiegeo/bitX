package network

import (
	"../hashtree"
	"code.google.com/p/goprotobuf/proto"
)

//Put File information into Packet, reusing existing file hash if possible
func (p *Packet) MergeFile(f *File) {
	files := p.GetFiles()
	haveFile := false
	for _, v := range files {
		if f.GetId().Equal(v.GetId()) {
			proto.Merge(v, f) // todo: a better merge
			haveFile = true
			break
		}
	}
	if !haveFile {
		files = append(files, f)
		p.Files = files
	}
}

//add file request information to the packet
func (p *Packet) FillHashRequest(id *StaticId, height hashtree.Level, from, length hashtree.Nodes) {
	hashReq := NewInnerHashes(height, from, length, nil)
	file := &File{Id: id, HashAsk: []*InnerHashes{&hashReq}}
	p.MergeFile(file)
}
