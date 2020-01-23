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

	fmt.Println("groups:")
	for _, g := range groups {
		fmt.Printf("- id: %v\n", g.ID)
		fmt.Printf("  name: %v\n", g.Name)
		fmt.Printf("  stream: %v\n", g.Stream)
		fmt.Printf("  speakers:\n")
		for _, c := range g.Speakers {
			fmt.Printf("  - %v\n", c.Name)
		}
		fmt.Println()
	}

	streams, err := client.Streams()
	if err != nil {
		log.Fatalf("could not get streams: %v", err)
	}
	fmt.Println("streams:")
	for _, s := range streams {
		fmt.Printf("- %s\n", s)
	}
}
