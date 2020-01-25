package main

import (
	"context"
	"flag"
	"fmt"
	"log"

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
	if *groupName == "" || *stream == "" {
		log.Fatal("must set -group and -stream")
	}

	snapserverAddr := fmt.Sprintf("%v:%v", *snapserverHost, *snapserverPort)
	client := snapcast.Dial("tcp", snapserverAddr)
	defer client.Close()

	client.SetConnectHandler(func() {
		log.Print("connected")
	})
	client.SetErrorHandler(func(err error) {
		log.Printf("client error: %v", err)
	})

	ctx := context.Background()
	groups, err := client.Groups(ctx)
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

	if err := client.SetStream(ctx, id, snapcast.Stream(*stream)); err != nil {
		log.Fatalf("could not set stream: %v", err)
	}
}
