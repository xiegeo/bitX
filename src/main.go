package main

import (
	"net"
	"./network"
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"flag"
)

var serverHello = &network.ServerHello{
	CodeName:      proto.String("BX"),
	VersionNumber: proto.String("0.0 pre-alpha"),
}

var about = flag.Bool("about", false, "shows server information")
var ip = flag.String("ip", "0.0.0.0", "the ip address to listen on")
var port = flag.Int("port", 6170, "the UDP port to listen on")

func main() {
	flag.Parse()
	if(*about){
		fmt.Println(proto.MarshalTextString(serverHello))
	}
}


