package main

import (
	"./interface/ihttp"
	"./network"
	proto "code.google.com/p/goprotobuf/proto"
	"flag"
	"fmt"
)

var serverHello = &network.ServerHello{
	CodeName:      proto.String("BX"),
	VersionNumber: proto.String("0.0 pre-alpha"),
}

var about = flag.Bool("about", false, "shows server information")
var ip = flag.String("ip", "0.0.0.0", "the ip address to listen on")
var port = flag.Int("port", 6170, "the UDP port to listen on")

var httpOn = flag.Bool("http", false, "runs the http interface")
var httpListen = flag.String("httpL", "localhost:8088", "the address and port of http interface")

func main() {
	flag.Parse()
	if *about {
		fmt.Println(proto.MarshalTextString(serverHello))
	}
	if *httpOn {
		fmt.Printf("starting server: %v\n", *httpListen)
		ihttp.StartServer(*httpListen)
	}
}
