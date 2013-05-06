package server

import (
	"../network"
	"bufio"
	"io"
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
}

func OpenSimpleDatabase(dirname string) Database {
	os.MkdirAll(dirname, 0777)
	dir, err := os.Open(dirname)
	if err != nil {
		panic(err)
	}
	d := &simpleDatabase{
		datafolder: dir,
	}

	return d
}

func (d *simpleDatabase) ImportFromReader(r io.Reader) network.StaticId {
	//TODO: save and return hash and length
	return network.StaticId{}
}

func (d *simpleDatabase) Get(id network.StaticId, from, length int64) ([]byte, error) {
	//TODO: return saved
	return make([]byte, length), nil
}
