package server

import (
	"../hashtree"
	"../network"
	//"bytes"
	//"io"
	"os"
	"testing"
)

func TestPart(t *testing.T) {
	source, part := testSetUp(t)
	defer source.Close()
	defer part.Close()

	test := func(size hashtree.Bytes) {
		testPartSize(size, source, part, t)
	}
	//test(0) //don't test the zero file, it's logic is too different
	test(1)
	test(1024)
	test(1025)
	test(2345)
	test(12345)
}

type testHashReq struct {
	req   network.InnerHashes
	count hashtree.Nodes
}

func TestPartOutOfOrder(t *testing.T) {
	source, part := testSetUp(t)
	defer source.Close()
	defer part.Close()

	id := source.ImportFromReader(&testFile{length: 6 * hashtree.FILE_BLOCK_SIZE})
	testPartStart(part, id, t)

	testData := []testHashReq{
		{network.NewInnerHashes(0, 0, 5, nil), 0},         //nothing can be verified
		{network.NewInnerHashes(0, 1, 5, nil), 0},         //nothing can be verified
		{network.NewInnerHashes(2, 0, 2, nil), 3},         //top two, with one propergated down
		{network.NewInnerHashes(0, 1, 5, nil), 3 + 2},     //plus the last two, 1 to 3 lost
		{network.NewInnerHashes(0, 0, 3, nil), 3 + 2},     //but nothing here
		{network.NewInnerHashes(0, 0, 4, nil), 3 + 2 + 6}, //finnally
	}
	for _, v := range testData {
		n, _, err := testTransferHash(v.req, id, source, part, t)
		if err != nil {
			t.Fatal(err)
		}
		if n != v.count {
			t.Fatalf("count got:%v, should be %v, for test:%v", n, v.count, v.req.String())
		}
	}
}

func testSetUp(t *testing.T) (source Database, part Database) {
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
	source = OpenSimpleDatabase(sourceDatabase, testLevelLow)
	part = OpenSimpleDatabase(partDatabase, testLevelLow)
	return
}

func testPartSize(size hashtree.Bytes, source Database, part Database, t *testing.T) {
	//create file in source and get it's link
	id := source.ImportFromReader(&testFile{length: size})
	t.Logf("id:%v", id.String())
	if source.GetState(id) != FILE_COMPLETE {
		t.Fatalf("The source should have this file:%v", id.CompactId())
	}

	//first test by Import
	testPartStart(part, id, t)
	part.ImportFromReader(&testFile{length: size})
	if part.GetState(id) != FILE_COMPLETE {
		t.Fatalf("The file should be complete:%v", id.CompactId())
	}

	//then test by Put...
	part.Remove(id)
	testPartStart(part, id, t)
	leafs := hashtree.FileNodesDefault(size)
	req := network.NewInnerHashes(0, 0, leafs, nil)
	if leafs == 1 {
		req = network.NewInnerHashes(0, 0, 0, nil)
	}
	_, complete, _ := testTransferHash(req, id, source, part, t)
	if !complete {
		t.Fatal("should have put all inner hashes in")
	}

	data := make([]byte, id.GetLength())
	_, err := source.GetAt(data, id, 0)
	if err != nil {
		t.Fatalf("can't get from source:%v", err)
	}

	has, comp, err := part.PutAt(data, id, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !comp || has != leafs {
		t.Fatalf("complete:%v, %v/%v", comp, has, leafs)
	}
	if part.GetState(id) != FILE_COMPLETE {
		t.Fatalf("The file should be complete:%v", id.CompactId())
	}
}

func testTransferHash(req network.InnerHashes, id network.StaticId, source Database, part Database, t *testing.T) (has hashtree.Nodes, complete bool, err error) {
	t.Logf("req:%v", req.String())
	hashes, err := source.GetInnerHashes(id, req)
	if err != nil {
		t.Error(err)
		return -1, false, err
	}
	t.Logf("hashes:%v", hashes.String())
	n, complete, err2 := part.PutInnerHashes(id, hashes)
	if err2 != nil {
		t.Error(err2)
	}
	return n, complete, err2
}

func testPartStart(part Database, id network.StaticId, t *testing.T) {
	//part database should start with unknow for the link
	//in the furture, a smart database does not have to follow this strictly if it can compute the data, such as for the empty file.
	if part.GetState(id) != FILE_UNKNOW {
		t.Fatalf("Can't test file that already exist:%v", id.CompactId())
	}
	//StartPart, now hashes/parts of that file can be added in database
	err := part.StartPart(id)
	if err != nil {
		t.Fatalf("StartPart error:%v", err)
	}
	if part.GetState(id) != FILE_PART {
		t.Fatalf("should have started saving parts for:%v", id.CompactId())
	}
}
