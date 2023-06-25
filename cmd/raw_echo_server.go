package main

import (
	"flag"
	"fmt"
	"github.com/skiloop/echo-server/raw"
)

var (
	addr     = flag.String("addr", "0.0.0.0:9010", "server address")
	certFile = flag.String("cert", "", "cert file")
	keyFile  = flag.String("key", "", "key file")
)

func main() {
	flag.Parse()
	if *certFile == "" || *keyFile == "" {
		fmt.Println("cert and key file needed")
		return
	}
	raw.StartRawEcho(*addr, *certFile, *keyFile)
}
