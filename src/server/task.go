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

type source interface {
	RequestableSize() hashtree.Bytes
	AddRequest(file *network.File)
	RemoveRequest(file *network.File)
	Reset()
}

type bitxUDPSource struct {
	addr        *net.UDPAddr
	conn        *network.BitXConn
	c           chan *network.File
	requesting  []*network.File
	requestSize hashtree.Bytes
}

type downloadTask struct {
	id      network.StaticId
	sources []source
}

func newUDPSource(conn *network.BitXConn, addr *net.UDPAddr) source {
	u := &bitxUDPSource{addr: addr, conn: conn, c: make(chan *network.File)}
	u.Reset()
	go u.consume()
	return u
}

func (b *bitxUDPSource) RequestableSize() hashtree.Bytes {
	return b.requestSize
}

func (b *bitxUDPSource) AddRequest(file *network.File) {
	size := file.RequestedPayLoadSize()
	if b.requestSize < size {
		panic(fmt.Errorf("to much requested %v < %v", b.requestSize, size))
	}
	b.requestSize -= size
	b.conn.GetListener().Add(file.GetId(), b.c)
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

}

func (b *bitxUDPSource) Reset() {
	b.requesting = nil
	b.requestSize = UDP_START_REQUEST_SIZE
}
