package server

import (
	"../network"
	"bufio"
	"io"
	"io/ioutil"
	"os"
)

type Database interface {
	ImportFromReader(r io.Reader) network.StaticId
	Get(id network.StaticId, from, length int64) ([]byte, error)
}

func ImportLocalFile(d Database, location string) network.StaticId {
	f, _ := os.Open(location)
	r := bufio.NewReader(f)
	id := d.ImportFromReader(r)
	f.Close()
	return id
}

type simpleDatabase struct {
	datafolder *os.File
	dirname    string
}

func OpenSimpleDatabase(dirname string) Database {
	os.MkdirAll(dirname, 0777)
	dir, err := os.Open(dirname)
	if err != nil {
		panic(err)
	}
	d := &simpleDatabase{
		datafolder: dir,
		dirname:    dirname,
	}

	return d
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
	ulen := uint64(len)
	return network.StaticId{Length:&ulen}
}

func (d *simpleDatabase) Get(id network.StaticId, from, length int64) ([]byte, error) {
	//TODO: return saved
	return make([]byte, length), nil
}
