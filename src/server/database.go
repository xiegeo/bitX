package server

import (
	"../bitset"
	"../hashtree"
	"../network"
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var ERROR_NOT_LOCAL = errors.New("file is not locally available")
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
	GetAt(b []byte, id network.StaticId, off hashtree.Bytes) (int, error)
	GetInnerHashes(id network.StaticId, req network.InnerHashes) (network.InnerHashes, error)
	StartPart(id network.StaticId) error
	PutAt(b []byte, id network.StaticId, off hashtree.Bytes) error
	PutInnerHashes(id network.StaticId, set network.InnerHashes) error
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

	return d
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
		file.WriteAt(b, off)
	}
	hasher.SetInnerHashListener(listener)
	return file
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
	err = os.Rename(f.Name(), d.fileNameForId(id))
	if err != nil {
		if os.IsExist(err) {
			os.Remove(f.Name())
		} else {
			panic(err)
		}
	}
	err = os.Rename(hashFile.Name(), d.hashFileNameForId(id))
	if err != nil {
		if os.IsExist(err) {
			os.Remove(hashFile.Name())
		} else {
			panic(err)
		}
	}
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

func (d *simpleDatabase) GetAt(b []byte, id network.StaticId, off hashtree.Bytes) (int, error) {
	f, err := os.Open(d.fileNameForId(id))
	if err != nil {
		return 0, ERROR_NOT_LOCAL
	}
	defer f.Close()
	return f.ReadAt(b, int64(off))
}

func (d *simpleDatabase) GetInnerHashes(id network.StaticId, req network.InnerHashes) (network.InnerHashes, error) {
	leaf := refHash.Nodes(hashtree.Bytes(id.GetLength()))
	level := hashtree.Level(req.GetHeight())
	from := hashtree.Nodes(req.GetFrom())
	nodes := hashtree.Nodes(req.GetLength())
	if level < d.lowestInnerHashes {
		return req, ERROR_LEVEL_LOW
	} else if from+nodes > refHash.LevelWidth(leaf, level) {
		return req, ERROR_INDEX_OFF
	}
	f, err := os.Open(d.hashFileNameForId(id))
	if err != nil {
		return req, ERROR_NOT_LOCAL
	}
	defer f.Close()
	off := d.hashPosition(leaf, level, from)
	b := make([]byte, refHash.Size()*int(nodes))
	f.ReadAt(b, off)
	req.Hashes = b
	return req, nil
}
func (d *simpleDatabase) StartPart(id network.StaticId) error {
	_, err := os.Stat(d.hashFileNameForId(id))
	if os.IsNotExist(err) {
		_, err := os.Create(d.partFileNameForId(id))
		if err != nil {
			return err
		}
		_, err = os.Create(d.hashFileNameForId(id))
		if err != nil {
			return err
		}
		return nil
	} else {
		return ERROR_ALREADY_EXIST
	}

}
func (d *simpleDatabase) PutAt(b []byte, id network.StaticId, off hashtree.Bytes) error {
	return nil
}
func (d *simpleDatabase) PutInnerHashes(id network.StaticId, set network.InnerHashes) error {
	leafs := id.Blocks()
	bits := bitset.OpenFileBacked(d.haveHashNameForId(id), int(d.hashTopNumber(leafs)-1))
	defer bits.Close()
	splited := set.SplitLocalSummable(&id)
	for _, hashes := range splited {
		l, n := hashes.LocalRoot()
		key := int(d.hashNumber(leafs, l, n))
		if key == bits.Capacity() {
			//this is root
		} else if !bits.Get(key) {
			continue // this part of hashes can not be verified, skiped
		}
		//sum := hashes.LocalSum()

	}
	return nil
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
