package server

import (
	"../hashtree"
	"../network"
	"io"
	"sync"
	"time"
)

// Implements WaitFor for Database, a database expect WaitFor
// to happen on different go routines than all other functions
type ListeningDatabase struct {
	Database
	listeners map[string]map[chan FileState]bool
	lock      sync.RWMutex //locks listeners R & W
}

func NewListeningDatabase(d Database) *ListeningDatabase {
	return &ListeningDatabase{d, make(map[string]map[chan FileState]bool), sync.RWMutex{}}
}

func (d *ListeningDatabase) AddListener(id network.StaticId, listener chan FileState) {
	if cap(listener) < 1 {
		panic("listener must have a buffer")
	}
	sid := id.CompactId()
	d.lock.Lock()
	defer d.lock.Unlock()
	ls, ok := d.listeners[sid]
	if !ok {
		ls = make(map[chan FileState]bool)
		d.listeners[sid] = ls
	}
	ls[listener] = true
}

func (d *ListeningDatabase) RemoveListener(id network.StaticId, listener chan FileState) {
	sid := id.CompactId()
	d.lock.Lock()
	defer d.lock.Unlock()
	ls, ok := d.listeners[sid]
	if ok {
		delete(ls, listener)
		if len(ls) == 0 {
			delete(d.listeners, sid)
		}
	}
}

func (d *ListeningDatabase) writeHappend(id network.StaticId) {
	sid := id.CompactId()
	_, ok := d.listeners[sid]
	if ok {
		state := d.GetState(id)
		// only get state (which may use disk) if there are listeners,
		// but before lock so it does not block, then we have to get
		// listeners again after lock (which is just one more map lookup)
		d.lock.RLock()
		defer d.lock.RUnlock()
		ls, ok := d.listeners[sid]
		if ok {
			for listener := range ls {
				select {
				//remove old stuff in listener, if any
				case <-listener:
					continue
				default:
					listener <- state
				}
			}
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
	listener := make(chan FileState, 1)
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
