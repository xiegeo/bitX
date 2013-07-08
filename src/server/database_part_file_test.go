package server

import (
	"../hashtree"
	//"../network"
	//"bytes"
	//"io"
	"os"
	"testing"
)

func TestPart(t *testing.T) {
	sourceDatabase := ".testSourceDatabase"
	partDatabase := ".testPartDatabase"
	err := os.RemoveAll(sourceDatabase)
	if err != nil {
		t.Fatal(err)
	}
	err = os.RemoveAll(partDatabase)
	if err != nil {
		t.Fatal(err)
	}
	source := OpenSimpleDatabase(sourceDatabase, testLevelLow)
	part := OpenSimpleDatabase(partDatabase, testLevelLow)
	test := func(size hashtree.Bytes) {
		testPartSize(size, source, part, t)
	}

	test(0)
	test(1)
	test(1024)
	test(1025)
	test(2345)
	test(12345)
}
func testPartSize(size hashtree.Bytes, source Database, part Database, t *testing.T) {
	//create file in source and get it's link
	id := source.ImportFromReader(&testFile{length: size})
	if source.GetState(id) != FILE_COMPLETE {
		t.Fatalf("The source should have this file:%s", id.CompactId())
	}
	//part database should start with unknow for the link
	//in the furture, a smart database does not have to follow this strictly if it can compute the data, such as for the empty file.
	if part.GetState(id) != FILE_UNKNOW {
		t.Fatalf("Can't test file that already exist:%s", id.CompactId())
	}
	//StartPart, now parts of that file can be added in database
	err := part.StartPart(id)
	if err != nil {
		t.Fatalf("StartPart error:%x", err)
	}
	if part.GetState(id) != FILE_PART {
		t.Fatalf("should have started saving parts for:%s", id.CompactId())
	}
	//after all parts are saved, it should be complete
	part.ImportFromReader(&testFile{length: size})
	if part.GetState(id) != FILE_COMPLETE {
		t.Fatalf("The file should be complete:%s", id.CompactId())
	}
}
