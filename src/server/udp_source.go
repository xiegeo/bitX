package server

import (
	"../hashtree"
	"../network"
	"fmt"
	"net"
)

const (
	UDP_START_REQUEST_SIZE = 1024 * 8
	//TODO:  congestion control
)

type bitxUDPSource struct {
	addr         *net.UDPAddr
	conn         *network.BitXConn
	c            chan *network.File
	requesting   []*network.File
	_requestSize hashtree.Bytes
}

func newUDPSource(conn *network.BitXConn, addr *net.UDPAddr) source {
	u := &bitxUDPSource{addr: addr, conn: conn, c: make(chan *network.File)}
	u.Reset()
	go u.consume()
	return u
}

func (b *bitxUDPSource) RequestableSize() hashtree.Bytes {
	return b.getRequestSize()
}

func (b *bitxUDPSource) AddRequest(file *network.File) {
	size := file.RequestedPayLoadSize()
	if b.getRequestSize() < size {
		panic(fmt.Errorf("to much requested %v < %v", b.getRequestSize(), size))
	}
	b.changeRequestSize(-size)
	b.conn.GetListener().Add(file.GetId(), b.c)
}

func (b *bitxUDPSource) getRequestSize() hashtree.Bytes {
	return b._requestSize
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
	b.RemoveRequest(f)
}

func (b *bitxUDPSource) RemoveRequest(f *network.File) {
	//todo: should calculated how much data requested got
	b.changeRequestSize(1024)
}

func (b *bitxUDPSource) Reset() {
	b.requesting = nil
	b._requestSize = UDP_START_REQUEST_SIZE
}
