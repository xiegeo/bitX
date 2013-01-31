package network

import (
	"testing"
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
