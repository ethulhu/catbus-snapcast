package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ethulhu/mqtt-snapcast-bridge/snapcast"
)

var (
	host = flag.String("host", "", "host")
	port = flag.Uint("port", 0, "port")
)

func main() {
	flag.Parse()

	if *host == "" || *port == 0 {
		log.Fatal("must set -host and -port")
	}

	addr := fmt.Sprintf("%v:%v", *host, *port)

	client := snapcast.NewClient(addr)

	client.SetGroupStreamChangedHandler(func(groupID string, stream snapcast.Stream) {
		log.Printf("group %v changed to stream %v", groupID, stream)
	})
	client.SetDisconnectHandler(func(err error) {
		log.Printf("disconnected: %v", err)
	})
	client.SetConnectHandler(func(client snapcast.Client) {
		log.Print("connected!")
	})

	client.Connect()
}
