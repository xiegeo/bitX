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

func (d *simpleDatabase) ImportFromReader(r io.Reader) (id network.StaticId) {
	f, err := ioutil.TempFile(d.dirname, "import-")
	if err != nil {
		panic(err)
	}
	len, err2 := io.Copy(f, r)
	if err2 != nil {
		panic(err2)
	}
	hasher := hashtree.NewFile()
	f.Seek(0, os.SEEK_SET)
	io.Copy(hasher, f)
	id = network.StaticId{
		Hash:   hasher.Sum(nil),
		Length: &len,
	}
	f.Close()
	err = os.Rename(f.Name(), d.fileNameForId(id))
	if err != nil {
		if os.IsExist(err) {
			os.Remove(f.Name())
		} else {
			panic(err)
		}
	}
	return
}
func (d *simpleDatabase) GetAt(b []byte, id network.StaticId, off int64) (int, error) {
	f, err := os.Open(d.fileNameForId(id))
	if err != nil {
		return 0, NOT_LOCAL
	}
	return f.ReadAt(b, off)
}

func (d *simpleDatabase) GetInnerHashes(id network.StaticId, req network.InnerHashes) (network.InnerHashes, error) {
	//TODO: fill in innerhashes
	return req, nil
}

func (d *simpleDatabase) fileNameForId(id network.StaticId) string {
	return fmt.Sprintf("%s/%s", d.datafolder.Name(), id.CompactId())
}
