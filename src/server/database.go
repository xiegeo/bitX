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

type SimpleDatabase struct {
	datafolder os.File
}

func (d *SimpleDatabase) ImportFromReader(r io.Reader) network.StaticId {
	//TODO: save and return hash and length
	return network.StaticId{}
}

func (d *SimpleDatabase) Get(id network.StaticId, from, length int64) ([]byte, error) {
	//TODO: return saved
	return make([]byte, length), nil
}
