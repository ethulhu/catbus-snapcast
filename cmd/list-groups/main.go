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
)

func main() {
	flag.Parse()

	if *snapserverHost == "" {
		log.Fatal("must set -snapserver-host")
	}
	snapserverAddr := fmt.Sprintf("%v:%v", *snapserverHost, *snapserverPort)

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

	for _, g := range groups {
		fmt.Printf("%v\t%v\t%v\n", g.Name, g.ID, g.Stream)
		for _, c := range g.Speakers {
			fmt.Printf("\t%v\n", c.Name)
		}
	}
}
