package server

import (
	"../hashtree"
	"../network"
	//"fmt"
	"net"
)

const (
	UDP_START_REQUEST_SIZE = 1024 * 8
	//TODO:  congestion control
)

type bitxUDPSource struct {
	addr         *net.UDPAddr
	conn         network.BitXConnecter
	c            chan *network.File
	requesting   []*network.File
	_requestSize hashtree.Bytes
}

func newUDPSource(conn network.BitXConnecter, addr *net.UDPAddr) Source {
	u := &bitxUDPSource{addr: addr, conn: conn, c: make(chan *network.File)}
	u.Reset()
	go u.consume()
	return u
}

func (b *bitxUDPSource) RequestableSize() hashtree.Bytes {
	return b._requestSize
}

func (b *bitxUDPSource) AddRequest(p *network.Packet) {
	size := p.RequestedPayLoadSize()
	if b.RequestableSize() < size {
		//todo turn back on after requests are sized
		//panic(fmt.Errorf("too much requested, Allow:%v RequestedPayLoadSize:%v", b.RequestableSize(), size))
	}
	b.changeRequestSize(-size)
	for _, file := range p.GetFiles() {
		b.conn.GetListener().Add(file.GetId(), b.c)
	}
	b.conn.Send(p, b.addr)
}

func (b *bitxUDPSource) changeRequestSize(delta hashtree.Bytes) {
	b._requestSize += delta
	if b._requestSize < 0 {
		b._requestSize = 0
	} else if b._requestSize > UDP_START_REQUEST_SIZE {
		b._requestSize = UDP_START_REQUEST_SIZE
	}
}

func (b *bitxUDPSource) consume() {
	for f := range b.c {
		b.process(f)
	}
}

func (b *bitxUDPSource) process(f *network.File) {
	b.removeRequestFile(f)
}

func (b *bitxUDPSource) removeRequestFile(f *network.File) {
	//todo: should calculated how much data requested got
	b.changeRequestSize(1024)
}

func (b *bitxUDPSource) RemoveRequest(p *network.Packet) {
	for _, file := range p.GetFiles() {
		b.removeRequestFile(file)
	}
}

func (b *bitxUDPSource) Reset() {
	b.requesting = nil
	b._requestSize = UDP_START_REQUEST_SIZE
}
