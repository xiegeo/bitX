package server

import (
	"../bitset"
	"../hashtree"
	"../network"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

var ERROR_NOT_LOCAL = errors.New("file is not locally available")
var ERROR_NOT_PART = errors.New("file is not downloading")
var ERROR_LEVEL_LOW = errors.New("the inner hash level is lower than cached")
var ERROR_INDEX_OFF = errors.New("the index of inner hashes is out of range for file of this size")
var ERROR_ALREADY_EXIST = errors.New("the thing was already added or started")

type FileState int

const (
	FILE_UNKNOW FileState = iota
	FILE_PART
	FILE_COMPLETE
)

type Database interface {
	LowestInnerHashes() hashtree.Level
	ImportFromReader(r io.Reader) network.StaticId
	GetState(id network.StaticId) FileState
	WaitFor(id network.StaticId, toState FileState, timeOut time.Duration) (ok bool, curState FileState)
	GetAt(b []byte, id network.StaticId, off hashtree.Bytes) (hashtree.Bytes, error)
	GetInnerHashes(id network.StaticId, req network.InnerHashes) (network.InnerHashes, error)
	StartPart(id network.StaticId) error
	PutAt(b []byte, id network.StaticId, off hashtree.Bytes) (has hashtree.Nodes, complete bool, err error)
	PutInnerHashes(id network.StaticId, set network.InnerHashes) (has hashtree.Nodes, complete bool, err error)
	Remove(id network.StaticId)
	Close()
}

func ImportLocalFile(d Database, location string) (id network.StaticId) {
	f, _ := os.Open(location)
	r := bufio.NewReader(f)
	id = d.ImportFromReader(r)
	f.Close()
	return
}

type simpleDatabase struct {
	datafolder        *os.File
	dirname           string
	lowestInnerHashes hashtree.Level
	listener          *ListeningDatabase
}

func OpenSimpleDatabase(dirname string, lowestInnerHashes hashtree.Level) Database {
	os.MkdirAll(dirname, 0777)
	dir, err := os.Open(dirname)
	if err != nil {
		panic(err)
	}
	d := &simpleDatabase{
		datafolder:        dir,
		dirname:           dirname,
		lowestInnerHashes: lowestInnerHashes,
	}
	d.listener = NewListeningDatabase(d)
	return d
}

func (d *simpleDatabase) Close() {
	err := d.datafolder.Close()
	if err != nil {
		panic(err)
	}
	d.datafolder = nil
}

func (d *simpleDatabase) LowestInnerHashes() hashtree.Level {
	return d.lowestInnerHashes
}

var refHash = hashtree.NewFile()

func (d *simpleDatabase) hashNumber(leafs hashtree.Nodes, l hashtree.Level, n hashtree.Nodes) int64 {
	sum := hashtree.Nodes(0)
	for i := hashtree.Level(0); i < l; i++ {
		sum += refHash.LevelWidth(leafs, i)
	}
	return int64(sum + n)
}
func (d *simpleDatabase) hashTopNumber(leafs hashtree.Nodes) int64 {
	sum := hashtree.Nodes(0)
	l := hashtree.Levels(leafs)
	for i := hashtree.Level(0); i < l; i++ {
		sum += refHash.LevelWidth(leafs, i)
	}
	return int64(sum)
}

func (d *simpleDatabase) hashPosition(leafs hashtree.Nodes, l hashtree.Level, n hashtree.Nodes) int64 {
	return d.hashNumber(leafs, l, n) * int64(refHash.Size())
}

func (d *simpleDatabase) innerHashListenerFile(hasher hashtree.HashTree, len hashtree.Bytes) *os.File {
	leafs := hasher.Nodes(len)
	top := hasher.Levels(leafs) - 1
	file, err := ioutil.TempFile(d.dirname, "listener-")
	if err != nil {
		panic(err)
	}
	listener := func(l hashtree.Level, i hashtree.Nodes, h *hashtree.H256) {
		//TODO: don't save levels lower than needed
		if l == top {
			return //don't need the root here
		}
		b := h.ToBytes()
		off := d.hashPosition(leafs, l, i)
		if _, err := file.WriteAt(b, off); err != nil {
			panic(err)
		}
	}
	hasher.SetInnerHashListener(listener)
	return file
}

func moveOrRemoveFile(from string, to string) {
	err := os.Rename(from, to)
	if err != nil {
		if os.IsExist(err) {
			os.Remove(from)
		} else {
			panic(err)
		}
	}
}

func (d *simpleDatabase) ImportFromReader(r io.Reader) network.StaticId {
	f, err := ioutil.TempFile(d.dirname, "import-")
	if err != nil {
		panic(err)
	}
	len, err2 := io.Copy(f, r)
	if err2 != nil {
		panic(err2)
	}
	hasher := hashtree.NewFile()
	hashFile := d.innerHashListenerFile(hasher, hashtree.Bytes(len))

	f.Seek(0, os.SEEK_SET)
	io.Copy(hasher, f)
	id := network.StaticId{
		Hash:   hasher.Sum(nil),
		Length: &len,
	}
	f.Close()
	hashFile.Close()

	moveOrRemoveFile(f.Name(), d.fileNameForId(id))
	moveOrRemoveFile(hashFile.Name(), d.hashFileNameForId(id))
	return id
}

func (d *simpleDatabase) GetState(id network.StaticId) FileState {
	_, err := os.Stat(d.fileNameForId(id))
	if os.IsNotExist(err) {
		_, err = os.Stat(d.hashFileNameForId(id))
		if os.IsNotExist(err) {
			return FILE_UNKNOW
		} else {
			return FILE_PART
		}
	} else {
		return FILE_COMPLETE
	}
}

func (d *simpleDatabase) WaitFor(id network.StaticId, toState FileState, timeOut time.Duration) (ok bool, curState FileState) {
	return d.listener.WaitFor(id, toState, timeOut)
}

func (d *simpleDatabase) GetAt(b []byte, id network.StaticId, off hashtree.Bytes) (hashtree.Bytes, error) {
	f, err := os.Open(d.fileNameForId(id))
	if err != nil {
		return 0, ERROR_NOT_LOCAL
	}
	defer f.Close()
	n, err := f.ReadAt(b, int64(off))
	return hashtree.Bytes(n), err
}

func (d *simpleDatabase) GetInnerHashes(id network.StaticId, req network.InnerHashes) (network.InnerHashes, error) {
	leafs := refHash.Nodes(hashtree.Bytes(id.GetLength()))
	level := hashtree.Level(req.GetHeight())
	from := hashtree.Nodes(req.GetFrom())
	nodes := hashtree.Nodes(req.GetLength())
	if nodes == 0 {
		return req, nil //nothing was requested
	} else if level < d.lowestInnerHashes {
		return req, ERROR_LEVEL_LOW
	} else if level >= refHash.Levels(leafs) || from+nodes > refHash.LevelWidth(leafs, level) {
		return req, ERROR_INDEX_OFF
	} else if level == refHash.Levels(leafs)-1 {
		req.Hashes = id.Hash
		return req, nil
	}
	f, err := os.Open(d.hashFileNameForId(id))
	if err != nil {
		return req, ERROR_NOT_LOCAL
	}
	defer f.Close()
	off := d.hashPosition(leafs, level, from)
	b := make([]byte, refHash.Size()*int(nodes))
	if _, err := f.ReadAt(b, off); err != nil {
		panic(err)
	}
	req.Hashes = b
	return req, nil
}

func (d *simpleDatabase) StartPart(id network.StaticId) error {
	_, err := os.Stat(d.hashFileNameForId(id))
	if os.IsNotExist(err) {
		pf, err := os.Create(d.partFileNameForId(id))
		if err != nil {
			return err
		}
		defer pf.Close()
		hf, err2 := os.Create(d.hashFileNameForId(id))
		if err2 != nil {
			return err2
		}
		defer hf.Close()
		return nil
	} else {
		return ERROR_ALREADY_EXIST
	}

}
func (d *simpleDatabase) putAt(b []byte, id network.StaticId, off hashtree.Bytes) (has hashtree.Nodes, complete bool, err error) {
	blocks := id.Blocks()
	f, err := os.OpenFile(d.partFileNameForId(id), os.O_RDWR, 0666)
	if err != nil {
		return 0, false, ERROR_NOT_PART
	}
	defer f.Close()
	bits := bitset.OpenCountingFileBacked(d.havePartNameForId(id), int(blocks))
	defer bits.Close()

	if off%hashtree.FILE_BLOCK_SIZE != 0 {
		panic("offset must start from a block")
	}
	startingNode := hashtree.Nodes(off / hashtree.FILE_BLOCK_SIZE)

	if off >= id.Bytes() {
		b = nil
	} else if off+hashtree.Bytes(len(b)) >= id.Bytes() {
		b = b[:id.Bytes()-off]
	}
	bLen := hashtree.Bytes(len(b))

	hash := hashtree.NewFile()

	for i := hashtree.Bytes(0); i < bLen || off+i == 0; i += hashtree.FILE_BLOCK_SIZE {
		// off+i == 0 is a special case for the empty file
		end := i + hashtree.FILE_BLOCK_SIZE
		if end > bLen {
			end = bLen
		}
		partData := b[i:end]
		hash.Write(partData)
		sum := hash.Sum(nil)
		hash.Reset()
		nthBlock := startingNode + hashtree.Nodes(i/hashtree.FILE_BLOCK_SIZE)
		exp, err := d.GetInnerHashes(id, network.NewInnerHashes(0, nthBlock, 1, nil))
		if err != nil {
			panic(err)
		}
		if bytes.Equal(exp.GetHashes(), sum) {
			_, err := f.WriteAt(partData, int64(nthBlock*hashtree.FILE_BLOCK_SIZE))
			if err != nil {
				panic(err)
			}
			bits.Set(int(nthBlock))
		}
	}
	return hashtree.Nodes(bits.Count()), bits.Full(), nil
}

func (d *simpleDatabase) PutAt(b []byte, id network.StaticId, off hashtree.Bytes) (has hashtree.Nodes, complete bool, err error) {
	has, complete, err = d.putAt(b, id, off)
	if err == nil && complete {
		moveOrRemoveFile(d.partFileNameForId(id), d.fileNameForId(id))
	}
	return
}

func (d *simpleDatabase) PutInnerHashes(id network.StaticId, set network.InnerHashes) (has hashtree.Nodes, complete bool, err error) {
	leafs := id.Blocks()
	f, err := os.OpenFile(d.hashFileNameForId(id), os.O_RDWR, 0666)
	if err != nil {
		return 0, false, ERROR_NOT_PART
	}
	defer f.Close()
	bits := bitset.OpenCountingFileBacked(d.haveHashNameForId(id), int(d.hashTopNumber(leafs)-1))
	defer bits.Close()

	writeHash := func(realL hashtree.Level, realN hashtree.Nodes, b []byte) {
		off := d.hashPosition(leafs, realL, realN)
		if _, err := f.WriteAt(b, off); err != nil {
			panic(err)
		}
		n := int(d.hashNumber(leafs, realL, realN))
		bits.Set(n)
	}

	hashBuffer := make([]byte, refHash.Size())
	splited := set.SplitLocalSummable(&id)
	for _, hashes := range splited {
		rootL, rootN := hashes.LocalRoot(leafs)
		key := int(d.hashNumber(leafs, rootL, rootN))
		if key == bits.Capacity() {
			//this is root
		} else if !bits.Get(key) {
			continue // this part of hashes can not be verified, skiped
		}
		sum := hashes.LocalSum()
		if key == bits.Capacity() {
			copy(hashBuffer, id.GetHash())
		} else {
			off := d.hashPosition(leafs, rootL, rootN)
			if _, err := f.ReadAt(hashBuffer, off); err != nil {
				panic(err)
			}
		}
		if bytes.Equal(sum, hashBuffer) {
			//verified, now save
			listener := func(l hashtree.Level, i hashtree.Nodes, h *hashtree.H256) {
				realL := set.GetHeightL() + l
				realN := i + rootN<<uint32(rootL-l)
				if realL == rootL {
					return //don't need the root here, already verified
				}
				b := h.ToBytes()
				writeHash(realL, realN, b)

				//propagate down nodes with single branch
				for realL > d.LowestInnerHashes() &&
					realN+1 == hashtree.LevelWidth(leafs, realL) &&
					hashtree.LevelWidth(leafs, realL-1)%2 == 1 {

					realL--
					realN = hashtree.LevelWidth(leafs, realL) - 1
					writeHash(realL, realN, b)
				}
			}
			hasher := hashtree.NewTree()
			hasher.SetInnerHashListener(listener)
			hasher.Write(hashes.GetHashes())
			hasher.Sum(nil)
		}
	}
	return hashtree.Nodes(bits.Count()), bits.Full(), nil
}

func remove(filename string) {
	err := os.Remove(filename)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
}

func (d *simpleDatabase) Remove(id network.StaticId) {
	remove(d.havePartNameForId(id))
	remove(d.haveHashNameForId(id))
	remove(d.partFileNameForId(id))
	remove(d.fileNameForId(id))
	remove(d.hashFileNameForId(id))
}

func (d *simpleDatabase) fileNameForId(id network.StaticId) string {
	return fmt.Sprintf("%s/F-%s", d.datafolder.Name(), id.CompactId())
}
func (d *simpleDatabase) hashFileNameForId(id network.StaticId) string {
	return fmt.Sprintf("%s/H-%s", d.datafolder.Name(), id.CompactId())
}
func (d *simpleDatabase) partFileNameForId(id network.StaticId) string {
	return fmt.Sprintf("%s/P-%s", d.datafolder.Name(), id.CompactId())
}
func (d *simpleDatabase) haveHashNameForId(id network.StaticId) string {
	return fmt.Sprintf("%s/hH-%s", d.datafolder.Name(), id.CompactId())
}
func (d *simpleDatabase) havePartNameForId(id network.StaticId) string {
	return fmt.Sprintf("%s/hP-%s", d.datafolder.Name(), id.CompactId())
}
