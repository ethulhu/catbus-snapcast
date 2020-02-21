package main

import (
	"fmt"
	"log"

	"github.com/ethulhu/mqtt-snapcast-bridge/jsonrpc2"
)

var (
	host = flag.String("host", "", "host of Snapserver")
	port = flag.Uint("port", 0, "port of Snapserver")
)

func main() {
	if *host == "" || *port == 0 {
		log.Fatal("must set -host and -port")
	}
	addr := fmt.Sprintf("%v:%v", *host, *port)

	client := jsonrpc2.Dial("tcp", addr)
	defer client.Close()

	client.SetConnectHandler(func() {
		log.Print("connected")
	})
	client.SetErrorHandler(func(err error) {
		log.Printf("client error: %v", err)
	})

	client.SetNotificationHandler("Group.OnStreamChanged", func(params interface{}) {
		fmt.Printf("%+v\n", params)
	})

	select {}
}
