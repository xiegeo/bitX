package network

import (
	"sync"
)

type fileChanSet map[chan *File]bool
type idMapType map[string]fileChanSet

type PacketListener struct {
	receive chan BitXPacket
	idMap   idMapType
	lock    sync.RWMutex // protect listener data structures
}

func newPacketListener() PacketListener {
	receive := make(chan BitXPacket)
	l := PacketListener{receive, make(idMapType), sync.RWMutex{}}
	go l.consume()
	return l
}

func (l *PacketListener) consume() {
	for p := range l.receive {
		l.process(p)
	}
}

func (l *PacketListener) process(bp BitXPacket) {
	rece := bp.Packet
	if rece.Files != nil {
		l.lock.RLock()
		defer l.lock.RUnlock()
		for _, f := range rece.Files {
			id := f.GetId()
			sid := id.CompactId()
			ls, ok := l.idMap[sid]
			if ok {
				for c := range ls {
					c <- f
				}
			}
		}
	}
}

func (l *PacketListener) Add(id *StaticId, c chan *File) {
	sid := id.CompactId()
	l.lock.Lock()
	defer l.lock.Unlock()
	ls, ok := l.idMap[sid]
	if !ok {
		ls = make(fileChanSet)
		l.idMap[sid] = ls
	}
	ls[c] = true
}

func (l *PacketListener) Remove(id *StaticId, c chan *File) {
	sid := id.CompactId()
	l.lock.Lock()
	defer l.lock.Unlock()
	ls, ok := l.idMap[sid]
	if ok {
		delete(ls, c)
		if len(ls) == 0 {
			delete(l.idMap, sid)
		}
	}
}
