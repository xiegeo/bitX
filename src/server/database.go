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
var LEVEL_HIGH = errors.New("the inner hash level is higher than root")

type Database interface {
	LowestInnerHashes() int
	ImportFromReader(r io.Reader) network.StaticId
	GetAt(b []byte, id network.StaticId, off int64) (n int, err error)
	GetInnerHashes(id network.StaticId, req network.InnerHashes) (network.InnerHashes, error)
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
	lowestInnerHashes int
}

func OpenSimpleDatabase(dirname string, lowestInnerHashes int) Database {
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

func (d *simpleDatabase) LowestInnerHashes() int {
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
func (d *simpleDatabase) GetAt(b []byte, id network.StaticId, off int64) (int, error) {
	f, err := os.Open(d.fileNameForId(id))
	if err != nil {
		return 0, NOT_LOCAL
	}
	defer f.Close()
	return f.ReadAt(b, off)
}

func (d *simpleDatabase) GetInnerHashes(id network.StaticId, req network.InnerHashes) (network.InnerHashes, error) {
	f, err := os.Open(d.hashFileNameForId(id))
	if err != nil {
		return network.InnerHashes{}, NOT_LOCAL
	}
	defer f.Close()
	off := d.hashPosition(refHash.Nodes(hashtree.Bytes(id.GetLength())), hashtree.Level(req.GetHeight()), hashtree.Nodes(req.GetFrom()))
	b := make([]byte, refHash.Size()*int(req.GetLength()))
	f.ReadAt(b, off)
	req.Hashes = b
	return req, nil
}

func (d *simpleDatabase) fileNameForId(id network.StaticId) string {
	return fmt.Sprintf("%s/F-%s", d.datafolder.Name(), id.CompactId())
}
func (d *simpleDatabase) hashFileNameForId(id network.StaticId) string {
	return fmt.Sprintf("%s/H-%s", d.datafolder.Name(), id.CompactId())
}
