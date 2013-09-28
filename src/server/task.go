package server

import (
	"../hashtree"
	"../network"
	"net"
)

const (
	UDP_START_REQUEST_SIZE = 1024 * 8
	//TODO:  congestion control
)

type source interface {
	RequestSize() hashtree.Bytes
	AddRequest(file *network.File)
	Reset()
}

type bitxUDPsource struct {
	addr        *net.UDPAddr
	conn        *network.BitXConn
	requestSize hashtree.Bytes
}

type downloadTask struct {
	id      network.StaticId
	sources []source
}

func (b *bitxUDPsource) RequestSize() hashtree.Bytes {
	return b.requestSize
}
