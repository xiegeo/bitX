package server

import (
	"../hashtree"
	"../network"
	"io"
	"time"
)

type ListeningDatabase struct {
	Database
	listeners map[*network.StaticId][]chan FileState
}

func (d *ListeningDatabase) writeHappend(id network.StaticId) {
	//send file state to listeners
}

func (d *ListeningDatabase) Close() {
	d.Database.Close()
	// terminate all listners
}

func (d *ListeningDatabase) ImportFromReader(r io.Reader) network.StaticId {
	id := d.Database.ImportFromReader(r)
	d.writeHappend(id)
	return id
}

func (d *ListeningDatabase) WaitFor(id network.StaticId, toState FileState, timeOut time.Duration) (ok bool, curState FileState, timeTook time.Duration) {
	//add listener
	return false, FILE_UNKNOW, 0
}

func (d *ListeningDatabase) StartPart(id network.StaticId) error {
	err := d.Database.StartPart(id)
	d.writeHappend(id)
	return err
}

func (d *ListeningDatabase) PutAt(b []byte, id network.StaticId, off hashtree.Bytes) (has hashtree.Nodes, complete bool, err error) {
	has, complete, err = d.Database.PutAt(b, id, off)
	d.writeHappend(id)
	return
}

func (d *ListeningDatabase) PutInnerHashes(id network.StaticId, set network.InnerHashes) (has hashtree.Nodes, complete bool, err error) {
	has, complete, err = d.Database.PutInnerHashes(id, set)
	d.writeHappend(id)
	return
}
