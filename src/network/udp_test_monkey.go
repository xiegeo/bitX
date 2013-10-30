package network

import (
	"math/rand"
	"net"
)

//simulate network conditions for testing by randomizing send
type ConnMonkey struct {
	BitXConnecter
	LossChance float32
}

func (m *ConnMonkey) Send(p *Packet, addr *net.UDPAddr) {
	if rand.Float32() < m.LossChance {
		log.Printf("Monkey losses a packet\n")
		return
	}
	m.BitXConnecter.Send(p, addr)
}
