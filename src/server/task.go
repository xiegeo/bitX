package server

import (
	"../bitset"
	"../hashtree"
	"../network"
)

type Source interface {
	RequestableSize() hashtree.Bytes
	AddRequest(file *network.File)
	RemoveRequest(file *network.File)
	Reset()
}

type DownloadTask struct {
	id              network.StaticId
	sources         []Source
	hashCompleted   bool
	requestedHashes bitset.BitSet
	requestedData   bitset.BitSet
}

func newDownlaodTask(d Database, id network.StaticId, sources []Source) *DownloadTask {
	return &DownloadTask{
		id:            id,
		sources:       sources,
		hashCompleted: false,
		//requested:
	}
}
