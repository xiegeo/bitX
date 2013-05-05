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
	p := bp.Packet
	if p.Hello != nil {
		log.Printf("got hello:%v from:%v", p.Hello, addr)
	}
	if p.GetHelloRequest() {
		log.Printf("req hello from:%v", addr)
		packet := &network.Packet{
			Hello: &s.Setting.Hello,
		}
		s.Conn.Send(packet, addr)
	}
	if p.File != nil {
		for _, f := range p.File {

		}
	}

}
