package network

import (
	"../hashtree"
	"code.google.com/p/goprotobuf/proto"
)

func (p *Packet) IsEmpty() bool {
	return p.Hello == nil && p.HelloRequest == nil && p.Retains == nil && p.Files == nil && p.Bebug == nil
}

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

//add file hash request information to the packet
func (p *Packet) FillHashRequest(id StaticId, height hashtree.Level, from, length hashtree.Nodes) {
	hashReq := NewInnerHashes(height, from, length, nil)
	file := &File{Id: &id, HashAsk: []*InnerHashes{&hashReq}}
	p.MergeFile(file)
}

//add file hash send information to the packet
func (p *Packet) FillHashSend(id StaticId, hashes InnerHashes) {
	file := &File{Id: &id, HashSend: []*InnerHashes{&hashes}}
	p.MergeFile(file)
}

//add file data request information to the packet
func (p *Packet) FillDataRequest(id StaticId, from, length hashtree.Bytes) {
	fileReq := NewFileData(from, length, nil)
	file := &File{Id: &id, DataAsk: []*FileData{&fileReq}}
	p.MergeFile(file)
}

//add file data send information to the packet
func (p *Packet) FillDataSend(id StaticId, from, length hashtree.Bytes, data []byte) {
	fileData := NewFileData(from, length, data)
	file := &File{Id: &id, DataSend: []*FileData{&fileData}}
	p.MergeFile(file)
}
