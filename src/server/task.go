package server

import (
	"../hashtree"
	"../network"
	"../bitset"
)

type Source interface {
	RequestableSize() hashtree.Bytes
	AddRequest(file *network.File)
	RemoveRequest(file *network.File)
	Reset()
}

type DownloadTask struct {
	id      network.StaticId
	sources []Source
	requesting bitset.BitSet
}



