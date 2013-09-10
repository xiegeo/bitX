package server

import (
	"../network"
	"fmt"
	logPkg "log"
	"os"
)

var log *logPkg.Logger

func init() {
	log = logPkg.New(os.Stdout, "server:", logPkg.LstdFlags)
}

type Server struct {
	setting Setting
	conn    *network.BitXConn
	Database
}

type Setting struct {
	Hello            network.ServerHello
	DatabaseLocation string
	DatabaseType     string
	Port             int
	IP               string
}

func NewServer(s Setting) *Server {
	conn, err := network.ListenUDP(s.IP, s.Port)
	if err != nil {
		panic(err)
	}
	if !conn.StartServerLoop() {
		panic(fmt.Errorf("can't start server:%v", s))
	}
	database := Database(nil)
	switch s.DatabaseType {
	case "simple":
		database = OpenSimpleDatabase(s.DatabaseLocation, 0)
	default:
		panic(fmt.Errorf("unknown database type:%v", s.DatabaseType))
	}

	return &Server{s, conn, database}
}

func (s *Server) consume(ps <-chan network.BitXPacket) {
	for p := range ps {
		s.process(p)
	}
}

func (s *Server) process(bp network.BitXPacket) {
	addr := bp.Addr
	rece := bp.Packet
	if rece.Hello != nil {
		log.Printf("got hello:%v from:%v", rece.Hello, addr)
	}
	if rece.GetHelloRequest() {
		log.Printf("req hello from:%v", addr)
		send := &network.Packet{}
		send.Hello = &s.setting.Hello
		s.conn.Send(send, addr)
	}

	if rece.Files != nil {
		for _, f := range rece.Files {
			id := f.Id
			log.Printf("about:%v", id)
		}
	}

}
