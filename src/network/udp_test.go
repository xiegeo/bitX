package network

import (
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"net"
	"testing"
	"time"
)

const (
	t_ip = "127.0.0.1"
	t_p1 = 61701
	t_p2 = 61702
)

func TestStartServer(t *testing.T) {
	conn, err := ListenUDP(t_ip, t_p1)
	if err != nil {
		t.Fatalf(t_ip, ":", t_p1, " ", err)
	}
	if conn.StartServerLoop() {
		if conn.StartServerLoop() {
			t.Fatalf("server should be alread started")
		}
	} else {
		t.Fatalf("can't start server")
	}
	errClose := conn.Close()
	if errClose != nil {
		t.Fatalf("error:", errClose)
	}
	errClose = conn.Close()
	if errClose == nil {
		t.Fatalf("should not close again")
	}
}

func TestSendHello(t *testing.T) {
	s1, _ := ListenUDP(t_ip, t_p1)
	s1.StartServerLoop()
	s2, _ := ListenUDP(t_ip, t_p2)
	s2.StartServerLoop()
	s2Address, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", t_ip, t_p2))
	serverHello := &ServerHello{
		CodeName: proto.String("TestSendHello"),
	}
	packet := &Packet{
		Hello: serverHello,
	}
	s1.Send(packet, s2Address)
	timeout := make(chan bool)
	go func() { time.Sleep(1 * time.Second); timeout <- true }()
	select {
	case ans := <-s2.receive:
		if *ans.Packet.Hello.CodeName != *serverHello.CodeName {
			t.Fatalf("got wrong packet", ans)
		}
	case <-timeout:
		t.Fatalf("timed out, not sent or received")
	}
}
