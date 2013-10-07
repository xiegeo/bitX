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
	complete      bool
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
		t.complete = true
	}

	return t
}

type TaskManager struct {
	tasks map[string]*DownloadTask
	d     Database
}

func (tm *TaskManager) AddDownload(id network.StaticId, sources []Source) {
	t, ok := tm.tasks[id.CompactId()]
	if !ok {
		t = newDownlaodTask(tm.d, id, sources)
		tm.tasks[id.CompactId()] = t
	} else {
		//todo just add sources to t
	}
}

// Request the amount of data up to maxAmount now (expected to be used per tic)
// Only data and inner-hashes are counted, other overhead aren't,
// In the future, there maybe more than one maxAmount (ie, one for local, one for world)
// Returns the actual amount request, which should be the same as maxAmount unless:
// - sourcesFull: all sourcess have there own bandwidth/request queue satuated
// - allRequested: all parts of all downloading files are requested
// - maxAmount - requested < the smallest size requestable (1024)
func (tm *TaskManager) doRequest(maxAmount hashtree.Bytes) (requested hashtree.Bytes, sourcesFull bool, allRequested bool) {
	reqLeft := maxAmount
	allRequested = true
	sourcesFull = true
	for _, t := range tm.tasks {
		if reqLeft <= 1024 {
			sourcesFull = false
			allRequested = false
			break
		}
		r, s, a := t.doRequest(reqLeft)
		reqLeft -= r
		sourcesFull = sourcesFull && s
		allRequested = allRequested && a
	}
	requested = maxAmount - reqLeft
	return
}

func (t *DownloadTask) doRequest(maxAmount hashtree.Bytes) (requested hashtree.Bytes, sourcesFull bool, allRequested bool) {
	//todo
	return 0, true, false
}
