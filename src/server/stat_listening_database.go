package server

import (
	"../hashtree"
	"../network"
	"io"
	"time"
)

type ListeningDatabase struct {
	Database
	listeners map[string][]chan FileState
}

func NewListeningDatabase(d Database) *ListeningDatabase {
	return &ListeningDatabase{d, make(map[string][]chan FileState)}
}

func (d *ListeningDatabase) AddListener(id network.StaticId, listener chan FileState) {
	sid := id.CompactId()
	d.listeners[sid] = append(d.listeners[sid], listener)
}

func (d *ListeningDatabase) RemoveListener(id network.StaticId, listener chan FileState) {
	sid := id.CompactId()
	ls, ok := d.listeners[sid]
	if ok {
		if len(ls) == 1 && ls[0] == listener {
			delete(d.listeners, sid)
		} else {
			for k, v := range ls {
				if v == listener {
					last := len(ls) - 1
					ls[k] = ls[last]
					d.listeners[sid] = ls[0:last]
					return
				}
			}
		}
	}
}

func (d *ListeningDatabase) writeHappend(id network.StaticId) {
	sid := id.CompactId()
	ls, ok := d.listeners[sid]
	if ok {
		state := d.GetState(id)
		for _, v := range ls {
			v <- state
		}
	}
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

func (d *ListeningDatabase) WaitFor(id network.StaticId, toState FileState, timeOut time.Duration) (ok bool, curState FileState) {
	listener := make(chan FileState)
	defer close(listener)
	d.AddListener(id, listener)
	defer d.RemoveListener(id, listener)

	startState := d.GetState(id)
	if startState == toState {
		return true, startState
	}
	for true {
		select {
		case state := <-listener:
			if state == toState {
				return true, state
			}
		case <-time.After(timeOut):
			state := d.GetState(id)
			return state == toState, state
		}
	}
	panic("code should not reach here")
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
