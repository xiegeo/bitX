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
	return &BitXConn{pudp, false, nil, nil}, nil
}

func (b *BitXConn) StartServerLoop() bool {
	if b.serverOn || b.send != nil || b.receive != nil {
		return false
	}
	b.serverOn = true
	log.Printf("Server started on:%s\n", b.conn.LocalAddr())
	send := make(chan BitXPacket, PacketBufferSize)
	receive := make(chan BitXPacket, PacketBufferSize)
	b.send = send
	b.receive = receive
	go func() {
		buf := make([]byte, 65507)
		for b.serverOn {
			n, addr, err := b.conn.ReadFromUDP(buf)
			if err != nil {
				log.Println("error ReadFromUDP:", err)
				continue
			}
			log.Println("from:", addr, "length", n)
			p := new(Packet)
			err = proto.Unmarshal(buf[:n], p)
			if err != nil {
				log.Println("error Unmarshal:", err)
				continue
			}
			receive <- BitXPacket{addr, p}
		}
		close(receive)
	}()
	return true
}
func (b *BitXConn) Close() error {
	b.serverOn = false
	err := b.conn.Close()
	if err == nil {
		log.Printf("Server stopped on:%s\n", b.conn.LocalAddr())
	}
	return err
}

func (b *BitXConn) Send(p *Packet, addr *net.UDPAddr) (int, error) {
	data, err := proto.Marshal(p)
	if err != nil {
		return 0, err
	}
	return b.conn.WriteToUDP(data, addr)
}