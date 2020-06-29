// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"go.eth.moe/catbus-snapcast/snapcast"
)

var (
	snapserverHost = flag.String("snapserver-host", "", "host of Snapserver (optional)")
	snapserverPort = flag.Uint("snapserver-port", snapcast.DefaultPort, "port of Snapserver")

	groupName = flag.String("group", "", "name of group to set")
	stream    = flag.String("stream", "", "name of stream")
)

func main() {
	flag.Parse()

	if *groupName == "" || *stream == "" {
		log.Fatal("must set -group and -stream")
	}

	var client snapcast.Client
	if *snapserverHost != "" {
		addr := fmt.Sprintf("%v:%v", *snapserverHost, *snapserverPort)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Fatalf("could not dial %v: %v", addr, err)
		}
		defer conn.Close()

		client = snapcast.NewClient(conn)
	} else {
		var err error
		client, err = snapcast.Discover()
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Print("connected")

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

	if err := client.SetGroupStream(ctx, id, snapcast.StreamID(*stream)); err != nil {
		log.Fatalf("could not set stream: %v", err)
	}
}
