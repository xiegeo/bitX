package server

import (
	"../hashtree"
	"../network"
	"fmt"
	"net"
	"testing"
)

func TestServerDataProvider(t *testing.T) {

	set1 := Setting{network.ServerHello{}, ".testServer1", "simple", 30001, "127.0.0.1"}
	set2 := Setting{network.ServerHello{}, ".testServer1", "simple", 30002, "127.0.0.1"}
	s1 := NewServer(set1)
	s2 := NewServer(set2)

	block := hashtree.Bytes(hashtree.FILE_BLOCK_SIZE)

	fileSizes := []hashtree.Bytes{0, 1, block - 1, block, block + 1, block * 2, block * 3, block * 4, block * 5}

	conn := s2.conn
	toAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", set1.IP, set1.Port))
	if err != nil {
		panic(err)
	}

	for _, size := range fileSizes {
		/*id :=*/ s1.ImportFromReader(&testFile{length: size})

		p := &network.Packet{}

		//todo fill p

		conn.Send(p, toAddr)

	}

}
