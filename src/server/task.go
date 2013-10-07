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
	id            network.StaticId
	sources       []Source
	hashCompleted bool
	hashesSet     bitset.BitSet
	dataSet       bitset.BitSet
}

func newDownlaodTask(d Database, id network.StaticId, sources []Source) *DownloadTask {

	t := &DownloadTask{
		id:      id,
		sources: sources,
	}

	state := d.GetState(id)

	switch state {
	case FILE_UNKNOW:
		err := d.StartPart(id)
		if err != nil {
			log.Printf("starting new DownloadTask err:%v", err)
		}
		t.hashCompleted = false
		t.hashesSet = d.InnerHashesSet(id)
	case FILE_PART:
		if !HaveAllInnerHashes(d, id) {
			t.hashCompleted = false
			t.hashesSet = d.InnerHashesSet(id)
		} else {
			t.hashCompleted = true
			t.dataSet = d.DataPartsSet(id)
		}
	case FILE_COMPLETE:
		//...
	}

	return t
}
