package server

import (
	"../network"
	logPkg "log"
)

var log *logPkg.Logger

func init() {
	log = logPkg.New(os.Stdout, "server:", logPkg.LstdFlags)
}

type Server struct {
	Setting Setting
	Conn *network.BitXConn
}

func (s *Server) consume(ps <-chan BitXPacket){
	for p := range ps{
		s.process(p)
	}
}

func (s *Server) process(bp BitXPacket){
	addr := bp.Addr
	p := bp.Packet
	if p.ServerHello != nil{
		log.Printf("got hello:%v from:%v", p.ServerHello, addr)
	}
	if p.GetHelloRequest {
		log.Printf("req hello from:%v", addr)
		s.Conn.Send(S.Setting.Hello,addr)
	}
	if p.File != nil{
		for _,f := range p.File{
			
		}
	}
	
}