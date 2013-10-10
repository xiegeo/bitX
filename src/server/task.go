package server

import (
	"../bitset"
	"../hashtree"
	"../network"
	"time"
)

type Source interface {
	RequestableSize() hashtree.Bytes
	AddRequest(p *network.Packet)
	RemoveRequest(p *network.Packet)
	Reset()
}

type DownloadTask struct {
	d             Database
	id            network.StaticId
	sources       []Source
	hashCompleted bool
	complete      bool
	hashesSet     bitset.BitSet
	dataSet       bitset.BitSet
}

func newDownlaodTask(d Database, id network.StaticId, sources []Source) *DownloadTask {

	t := &DownloadTask{
		d:       d,
		id:      id,
		sources: sources,
	}

	t.renewDatabase()

	return t
}

func (t *DownloadTask) renewDatabase() {
	state := t.d.GetState(t.id)

	switch state {
	case FILE_UNKNOW:
		err := t.d.StartPart(t.id)
		if err != nil {
			log.Printf("starting new DownloadTask err:%v", err)
		}
		t.hashCompleted = false
		t.hashesSet = t.d.InnerHashesSet(t.id)
	case FILE_PART:
		if !HaveAllInnerHashes(t.d, t.id) {
			t.hashCompleted = false
			t.hashesSet = t.d.InnerHashesSet(t.id)
		} else {
			t.hashCompleted = true
			t.dataSet = t.d.DataPartsSet(t.id)
		}
	case FILE_COMPLETE:
		t.complete = true
	}
}

type TaskManager struct {
	tasks map[string]*DownloadTask
	d     Database
}

func NewTaskManager(d Database) *TaskManager {
	return &TaskManager{make(map[string]*DownloadTask), d}
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

func (tm *TaskManager) runLoop() {
	perSec := hashtree.Bytes(1024 * 1024)
	ticksPerSec := 20
	tick := time.Second / time.Duration(ticksPerSec)
	perTick := perSec / hashtree.Bytes(ticksPerSec)
	for {
		tm.doRequest(perTick)
		time.Sleep(tick)
	}
}

const MIN_REQUEST = 1024

// Request the amount of data up to maxAmount now (expected to be used per tic)
// Only data and inner-hashes are counted, other overhead aren't,
// In the future, there maybe more than one maxAmount (ie, one for local, one for world)
// Returns the actual amount request, which should be the same as maxAmount unless:
// - sourcesFull: all sourcess have there own bandwidth/request queue satuated
// - stageFull: all parts of all downloading files (currently downloadable) are requested
// - maxAmount - requested < the smallest size requestable (MIN_REQUEST)
func (tm *TaskManager) doRequest(maxAmount hashtree.Bytes) (requested hashtree.Bytes, sourcesFull bool, stageFull bool) {
	reqLeft := maxAmount
	stageFull = true
	sourcesFull = true
	for _, t := range tm.tasks {
		if reqLeft < MIN_REQUEST {
			sourcesFull = false
			stageFull = false
			break
		}
		r, s, a := t.doRequest(reqLeft)
		reqLeft -= r
		sourcesFull = sourcesFull && s
		stageFull = stageFull && a
	}
	requested = maxAmount - reqLeft
	return
}

func (t *DownloadTask) doRequest(maxAmount hashtree.Bytes) (requested hashtree.Bytes, sourcesFull bool, stageFull bool) {
	t.renewDatabase() //todo: update more efficiently
	if t.complete {
		return 0, false, true
	}
	reqLeft := maxAmount
	stageFull = false
	sourcesFull = true
	for _, s := range t.sources {
		if reqLeft < MIN_REQUEST {
			sourcesFull = false
			stageFull = false
			break
		}
		size := s.RequestableSize()
		reqableSize := size
		if size < MIN_REQUEST {
			continue
		} else if size > reqLeft {
			reqableSize = reqLeft
		}
		p, full := t.getNextRequests(reqableSize)
		qsize := p.RequestedPayLoadSize()
		s.AddRequest(p)
		reqLeft -= qsize
		sourcesFull = sourcesFull && (s.RequestableSize() < MIN_REQUEST)
		if full {
			stageFull = true
			break
		}
	}
	return maxAmount - reqLeft, sourcesFull, stageFull
}

func (t *DownloadTask) getNextRequests(maxAmount hashtree.Bytes) (p *network.Packet, stageFull bool) {
	//todo: splite requests to maxAmount and based on what's already here
	p = &network.Packet{}
	if !t.hashCompleted {
		p.FillHashRequest(t.id, 0, 0, hashtree.FileNodesDefault(t.id.Bytes()))
	} else {
		p.FillDataRequest(t.id, 0, t.id.Bytes())
	}
	return p, true
}
