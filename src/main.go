package main

import (
	"./network"
	proto "code.google.com/p/goprotobuf/proto"
	"fmt"
	"strconv"
)

var serverHello = &network.ServerHello{
	CodeName:      proto.String("BX"),
	VersionNumber: proto.String("0.0 pre-alpha"),
}

func main() {
	fmt.Println(proto.MarshalTextString(serverHello))
	data, error := proto.Marshal(serverHello)
	fmt.Println(data, error)
	fmt.Println(strconv.QuoteToASCII(string(data)), error)
}
