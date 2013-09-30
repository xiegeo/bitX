package server

import (
	"../hashtree"
	"../network"
)

type source interface {
	RequestableSize() hashtree.Bytes
	AddRequest(file *network.File)
	RemoveRequest(file *network.File)
	Reset()
}

type downloadTask struct {
	id      network.StaticId
	sources []source
}
