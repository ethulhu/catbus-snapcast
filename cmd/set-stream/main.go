package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/ethulhu/go-snapcast"
)

var (
	snapserverHost = flag.String("snapserver-host", "", "host of Snapserver")
	snapserverPort = flag.Uint("snapserver-port", 1705, "port of Snapserver")

	groupName = flag.String("group", "", "name of group to set")
	stream    = flag.String("stream", "", "name of stream")
)

func main() {
	flag.Parse()

	if *snapserverHost == "" {
		log.Fatal("must set -snapserver-host")
	}
	snapserverAddr := fmt.Sprintf("%v:%v", *snapserverHost, *snapserverPort)

	if *groupName == "" || *stream == "" {
		log.Fatal("must set -group and -stream")
	}

	conn, err := net.Dial("tcp", snapserverAddr)
	if err != nil {
		log.Fatalf("could not dial %v: %v", snapserverAddr, err)
	}
	defer conn.Close()

	client := snapcast.NewClient(conn)

	groups, err := client.Groups()
	if err != nil {
		log.Fatalf("could not get groups: %v", err)
	}

	id := ""
	for _, group := range groups {
		if group.Name == *groupName {
			id = group.ID
		}
	}
	if id == "" {
		log.Fatalf("could not find group %v", *groupName)
	}

	if err := client.SetStream(id, snapcast.Stream(*stream)); err != nil {
		log.Fatalf("could not set stream: %v", err)
	}
}
