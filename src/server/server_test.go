package server

import (
	"../hashtree"
	"../network"
	"fmt"
	"net"
	"testing"
	"time"
)

func timeOutPanic() {
	time.Sleep(10 * time.Second)
	panic("dead lock?")
}

func TestServerDataProvider(t *testing.T) {

	go timeOutPanic()

	set1 := Setting{network.ServerHello{}, ".testServer1", "simple", 30011, "127.0.0.1"}
	set2 := Setting{network.ServerHello{}, ".testServer2", "simple", 30012, "127.0.0.1"}
	s1 := NewServer(set1)
	s1.conn = &network.ConnMonkey{s1.conn, 0.5}
	s2 := NewServer(set2)

	block := hashtree.Bytes(hashtree.FILE_BLOCK_SIZE)

	fileSizes := []hashtree.Bytes{0, 1, block - 1, block, block + 1, block * 2, block * 3, block * 4, block * 5}

	conn := s2.conn
	toAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", set1.IP, set1.Port))
	if err != nil {
		panic(err)
	}

	files := []network.StaticId{}

	tm := NewTaskManager(s2)
	go tm.runLoop()
	sources := []Source{newUDPSource(conn, toAddr)}
	for _, s := range sources {
		log.Printf("source url: %v", s.GetUrl().String())
	}

	for _, size := range fileSizes {
		id := s1.ImportFromReader(&testFile{length: size})
		files = append(files, id)

		s2.Remove(id)

		if s2.GetState(id) != FILE_UNKNOW {
			if size == 0 {
				if s2.GetState(id) != FILE_COMPLETE {
					t.Fatalf("the zero file should always be complete")
				}
			} else {
				t.Fatalf("file of length %v not unknown", id.GetLength())
			}
		}
		tm.AddDownload(id, sources)
	}

	for _, t := range tm.tasks {
		t.stageFullWait = time.Second / 10 //make the test faster
	}

	for _, id := range files {
		s2.WaitFor(id, FILE_COMPLETE, time.Second)
	}

	for _, id := range files {
		if s2.GetState(id) != FILE_COMPLETE {
			t.Fatalf("file of length %v not complete", id.GetLength())
		}
	}
}
