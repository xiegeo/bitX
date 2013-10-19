package server

import (
	"../bitset"
	"../hashtree"
	"../network"
	"time"
)

type DownloadTask struct {
	d             Database
	id            network.StaticId
	sources       []Source
	hashCompleted bool
	complete      bool
	hashesSet     *bitset.NextZeroBitSet
	dataSet       *bitset.NextZeroBitSet
	stageFulled   *time.Time //nil means stage not full
	stageFullWait time.Duration
}

func newDownlaodTask(d Database, id network.StaticId, sources []Source) *DownloadTask {

	t := &DownloadTask{
		d:             d,
		id:            id,
		sources:       sources,
		stageFullWait: time.Second, //the default wait time on finish a stage before repeating requests.(todo: better end game)
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
		t.hashesSet = bitset.NewNextZeroBitSet(t.d.InnerHashesSet(t.id))
	case FILE_PART:
		if !HaveAllInnerHashes(t.d, t.id) {
			t.hashCompleted = false
			t.hashesSet = bitset.NewNextZeroBitSet(t.d.InnerHashesSet(t.id))
		} else {
			t.hashCompleted = true
			t.dataSet = bitset.NewNextZeroBitSet(t.d.DataPartsSet(t.id))
		}
	case FILE_COMPLETE:
		t.complete = true
	}
}

func (t *DownloadTask) doRequest(maxAmount hashtree.Bytes, now time.Time) (requested hashtree.Bytes, sourcesFull bool, stageFull bool) {
	if t.stageFulled != nil {
		if t.stageFulled.Add(t.stageFullWait).Before(now) {
			// stage fulled for a long time
			log.Printf("doRequest renewDatabase stage full timeout:%v", t.id)
			t.stageFulled = nil
			t.renewDatabase()
		} else {
			log.Printf("doRequest skip stage full:%v", t.id)
			// currently full
			return 0, false, true
		}
	}

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
		size := s.RequestableSize(now)
		reqableSize := size
		if size < MIN_REQUEST {
			continue
		} else if size > reqLeft {
			reqableSize = reqLeft
		}
		p, full := t.getNextRequests(reqableSize)
		qsize := p.RequestedPayLoadSize()
		if qsize > 0 {
			s.AddRequest(p, now)
			reqLeft -= qsize
		}
		sourcesFull = sourcesFull && (s.RequestableSize(now) < MIN_REQUEST)
		if full {
			t.stageFulled = &now
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
		return p, true
	} else {
		s, l := t.dataSet.NextRange(int(maxAmount / hashtree.FILE_BLOCK_SIZE))
		if l == 0 {
			return p, true
		} else if s+l < int(t.id.Blocks()) {
			p.FillDataRequest(t.id, hashtree.Bytes(s*hashtree.FILE_BLOCK_SIZE), hashtree.Bytes(l*hashtree.FILE_BLOCK_SIZE))
			return p, false
		} else {
			//requesting last block, which may not be full
			byteLength := t.id.Bytes() - hashtree.Bytes(s*hashtree.FILE_BLOCK_SIZE)
			p.FillDataRequest(t.id, hashtree.Bytes(s*hashtree.FILE_BLOCK_SIZE), byteLength)
			return p, true
		}
	}

}
