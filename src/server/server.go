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
	conn    network.BitXConnecter
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

	server := &Server{s, conn, database}

	go server.consume(conn.Receive)

	return server
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
		send.Hello = &s.setting.Hello
	}

	if rece.Files != nil {
		for _, f := range rece.Files {
			id := f.Id
			log.Printf("about:%v", id.CompactId())
			for _, ha := range f.HashAsk {
				log.Printf("hash ask:%v", ha)
				hashes, err := s.GetInnerHashes(*id, *ha)
				if err != nil {
					log.Printf("err:%v", err)
				} else {
					send.FillHashSend(*id, hashes)
				}
			}
			for _, hs := range f.HashSend {
				log.Printf("hash send:%v", hs)
				number, comp, err := s.PutInnerHashes(*id, *hs)
				if err != nil {
					log.Printf("err:%v", err)
				} else {
					log.Printf("have inner hashes:%v, is completed:%v", number, comp)
				}
			}
			for _, da := range f.DataAsk {
				log.Printf("data ask:%v", da)
				b := make([]byte, da.GetLengthB())
				_, err := s.GetAt(b, *id, da.GetFromB())
				if err != nil {
					log.Printf("err:%v", err)
				} else {
					send.FillDataSend(*id, da.GetFromB(), da.GetLengthB(), b)
				}
			}
			for _, ds := range f.DataSend {
				log.Printf("data send:%v", ds.ShortString())
				number, comp, err := s.PutAt(ds.GetData(), *id, ds.GetFromB())
				if err != nil {
					log.Printf("err:%v", err)
				} else {
					log.Printf("have data blocks:%v, is completed:%v", number, comp)
				}
			}
		}
	}
	if !send.IsEmpty() {
		s.conn.Send(send, addr)
	}

}
