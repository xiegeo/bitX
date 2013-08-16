package server

import (
	"../network"
	logPkg "log"
	"os"
)

var log *logPkg.Logger

func init() {
	log = logPkg.New(os.Stdout, "server:", logPkg.LstdFlags)
}

type Server struct {
	Setting Setting
	Conn    *network.BitXConn
}

func (s *Server) consume(ps <-chan network.BitXPacket) {
	for p := range ps {
		s.process(p)
	}
}

func (s *Server) process(bp network.BitXPacket) {
	addr := bp.Addr
	rece := bp.Packet
	send := &network.Packet{}
	if rece.Hello != nil {
		log.Printf("got hello:%v from:%v", rece.Hello, addr)
	}
	if rece.GetHelloRequest() {
		log.Printf("req hello from:%v", addr)
		send.Hello = &s.Setting.Hello
	}

	if rece.File != nil {
		for _, f := range rece.File {
			id := f.Id
			log.Printf("about:%v", id)
		}
	}

	s.Conn.Send(send, addr)

}
