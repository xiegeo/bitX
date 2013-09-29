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
	on       bool
	Receive  <-chan BitXPacket
	listener PacketListener
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
	return &BitXConn{conn: pudp, on: false, listener: newPacketListener()}, nil
}

func (b *BitXConn) StartServerLoop() bool {
	if b.on || b.Receive != nil {
		return false
	}
	b.on = true
	log.Printf("Server started on:%s\n", b.conn.LocalAddr())
	receive := make(chan BitXPacket, PacketBufferSize)
	b.Receive = receive
	go func() {
		buf := make([]byte, 65507)
		for b.on {
			n, addr, err := b.conn.ReadFromUDP(buf)
			if err != nil {
				if b.on {
					log.Println("error ReadFromUDP:", err)
				}
				continue
			}
			log.Println("from:", addr, "length", n)
			p := new(Packet)
			err = proto.Unmarshal(buf[:n], p)
			if err != nil {
				log.Println("error Unmarshal:", err)
				continue
			}
			bp := BitXPacket{addr, p}
			b.listener.receive <- bp
			receive <- bp
		}
		close(receive)
	}()
	return true
}

func (b *BitXConn) GetListener() *PacketListener {
	return &b.listener
}

func (b *BitXConn) Close() error {
	b.on = false
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
