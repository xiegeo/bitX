package server

import (
	"../hashtree"
	"../network"
	"net"
	"net/url"
	"time"
)

const (
	UDP_START_REQUEST_SIZE  = 1024 * 8
	UDP_REQUEST_LOST_FACTER = 5 // times more than expected
	//TODO: more congestion control
)

type requestTracker struct {
	file    *network.File
	reqTime time.Time
}

type bitxUDPSource struct {
	addr         *net.UDPAddr
	conn         network.BitXConnecter
	c            chan *network.File
	requesting   []requestTracker
	_requestSize hashtree.Bytes
	lastTick     time.Time
}

func newUDPSource(conn network.BitXConnecter, addr *net.UDPAddr) Source {
	u := &bitxUDPSource{addr: addr, conn: conn, c: make(chan *network.File)}
	u.Reset()
	go u.consume()
	return u
}

func (b *bitxUDPSource) GetUrl() *url.URL {
	return &url.URL{Scheme: "bitx",
		Host: b.addr.String()}
}

func (b *bitxUDPSource) RequestableSize(now time.Time) hashtree.Bytes {
	if b.lastTick.Before(now) {
		b.lastTick = now
		notOutDated := make([]requestTracker, 0, len(b.requesting))
		keep := now.Add(-UDP_REQUEST_LOST_FACTER * time.Second / 10) //todo: calculate expected time
		for _, r := range b.requesting {
			if r.reqTime.After(keep) {
				notOutDated = append(notOutDated, r)
			} else {
				log.Printf("request timed out:%v", r)
				b.changeRequestSize(r.file.RequestedPayLoadSize())
			}
		}
		b.requesting = notOutDated
	}
	return b.requestableSize()
}

func (b *bitxUDPSource) requestableSize() hashtree.Bytes {
	return b._requestSize
}

func (b *bitxUDPSource) AddRequest(p *network.Packet, now time.Time) {
	size := p.RequestedPayLoadSize()
	if size == 0 {
		log.Printf("Error: empty request:%v", p)
	}
	if b.requestableSize() < size {
		log.Printf("Error: too much requested, Allow:%v RequestedPayLoadSize:%v", b.requestableSize(), size)
	}
	b.changeRequestSize(-size)
	for _, file := range p.GetFiles() {
		b.requesting = append(b.requesting, requestTracker{file, now})
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
