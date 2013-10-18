package server

import (
	"../hashtree"
	"../network"
	"time"
)

const MIN_REQUEST = 1024

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

// Request the amount of data up to maxAmount now (expected to be used per tic)
// Only data and inner-hashes are counted, other overhead aren't,
// In the future, there maybe more than one maxAmount (ie, one for local, one for world)
// Returns the actual amount request, which should be the same as maxAmount unless:
// - sourcesFull: all sourcess have there own bandwidth/request queue satuated
// - stageFull: all parts of all downloading files (currently downloadable) are requested
// - maxAmount - requested < the smallest size requestable (MIN_REQUEST)
func (tm *TaskManager) doRequest(maxAmount hashtree.Bytes) (requested hashtree.Bytes, sourcesFull bool, stageFull bool) {
	now := time.Now()
	reqLeft := maxAmount
	stageFull = true
	sourcesFull = true
	for _, t := range tm.tasks {
		if reqLeft < MIN_REQUEST {
			sourcesFull = false
			stageFull = false
			break
		}
		r, s, a := t.doRequest(reqLeft, now)
		reqLeft -= r
		sourcesFull = sourcesFull && s
		stageFull = stageFull && a
	}
	requested = maxAmount - reqLeft
	return
}
