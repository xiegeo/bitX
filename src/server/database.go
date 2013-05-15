package server

import (
	"../hashtree"
	"../network"
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

var NOT_LOCAL = errors.New("file is not locally available")
var LEVEL_LOW = errors.New("the inner hash level is lower than cached")
var INDEX_OFF = errors.New("the inner hashes does not exist for file of this size")

type Database interface {
	LowestInnerHashes() hashtree.Level
	ImportFromReader(r io.Reader) network.StaticId
	GetAt(b []byte, id network.StaticId, off hashtree.Bytes) (int, error)
	GetInnerHashes(id network.StaticId, req network.InnerHashes) (network.InnerHashes, error)
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

func (d *simpleDatabase) hashPosition(leafs hashtree.Nodes, l hashtree.Level, n hashtree.Nodes) int64 {
	sum := hashtree.Nodes(0)
	for i := hashtree.Level(0); i < l; i++ {
		sum += refHash.LevelWidth(leafs, i)
	}
	return int64(sum+n) * int64(refHash.Size())
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
func (d *simpleDatabase) GetAt(b []byte, id network.StaticId, off hashtree.Bytes) (int, error) {
	f, err := os.Open(d.fileNameForId(id))
	if err != nil {
		return 0, NOT_LOCAL
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
		return req, LEVEL_LOW
	} else if from+nodes > refHash.LevelWidth(leaf, level) {
		return req, INDEX_OFF
	}
	f, err := os.Open(d.hashFileNameForId(id))
	if err != nil {
		return req, NOT_LOCAL
	}
	defer f.Close()
	off := d.hashPosition(leaf, level, from)
	b := make([]byte, refHash.Size()*int(nodes))
	f.ReadAt(b, off)
	req.Hashes = b
	return req, nil
}

func (d *simpleDatabase) PutAt(b []byte, id network.StaticId, off hashtree.Bytes) error {
	return nil
}
func (d *simpleDatabase) PutInnerHashes(id network.StaticId, set network.InnerHashes) error {
	return nil
}

func (d *simpleDatabase) fileNameForId(id network.StaticId) string {
	return fmt.Sprintf("%s/F-%s", d.datafolder.Name(), id.CompactId())
}
func (d *simpleDatabase) hashFileNameForId(id network.StaticId) string {
	return fmt.Sprintf("%s/H-%s", d.datafolder.Name(), id.CompactId())
}
