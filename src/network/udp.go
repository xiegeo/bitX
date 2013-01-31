package network

import (
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	logPkg "log"
	"net"
	"os"
)

const (
	PacketBufferSize = 10
)

var log *logPkg.Logger

func init() {
	log = logPkg.New(os.Stdout, "network:", logPkg.LstdFlags)
}

type BitXConn struct {
	conn     *net.UDPConn
	serverOn bool
	send     chan<- BitXPacket
	receive  <-chan BitXPacket
}

type BitXPacket struct {
	Addr   *net.UDPAddr
	Packet *Packet
}

func ListenUDP(ip string, port int) (*BitXConn, error) {
	addr := fmt.Sprintf("%v:%v", ip, port)
	laddr, errResolve := net.ResolveUDPAddr("udp", addr)
	if errResolve != nil {
		return nil, errResolve
	}
	pudp, errListen := net.ListenUDP("udp", laddr)
	if errListen != nil {
		return nil, errListen
	}
	log.Printf("Server started on:%s\n", pudp.LocalAddr())
	return &BitXConn{pudp, false, nil, nil}, nil
}

func (b *BitXConn) StartServerLoop() bool {
	if b.serverOn || b.send != nil || b.receive != nil {
		return false
	}
	b.serverOn = true
	send := make(chan BitXPacket, PacketBufferSize)
	receive := make(chan BitXPacket, PacketBufferSize)
	b.send = send
	b.receive = receive
	loop := func() {
		buf := make([]byte, 65507)
		for b.serverOn {
			n, addr, err := b.conn.ReadFromUDP(buf)
			if err != nil {
				log.Println("error:", err)
				continue
			}
			log.Println("from:", addr, "length", n, "data:", buf)
			p := new(Packet)
			err = proto.Unmarshal(buf, p)
			if err != nil {
				log.Println("error:", err)
				continue
			}
			receive <- BitXPacket{addr, p}
		}
		close(receive)
	}
	go loop()
	return true
}
func (b *BitXConn) Close() error {
	b.serverOn = false
	return b.conn.Close()
}

func (b *BitXConn) Send(p *Packet, addr *net.UDPAddr) (int, error) {
	data, err := proto.Marshal(p)
	if err != nil {
		return 0, err
	}
	return b.conn.WriteToUDP(data, addr)
}
