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
	snapserverHost = flag.String("snapserver-host", "", "host of Snapserver")
	snapserverPort = flag.Uint("snapserver-port", snapcast.DefaultPort, "port of Snapserver")
)

func main() {
	flag.Parse()

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

	host, err := client.Host(ctx)
	if err != nil {
		log.Fatalf("could not get host: %v", err)
	}
	fmt.Printf("host: %s\n\n", host)

	groups, err := client.Groups(ctx)
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
			fmt.Printf("  - name: %v\n", c.Name)
			fmt.Printf("    connected: %v\n", c.Connected)
			fmt.Printf("    muted: %v\n", c.Volume.Muted)
			fmt.Printf("    volume: %v%%\n", c.Volume.Percent)
		}
		fmt.Println()
	}

	streams, err := client.Streams(ctx)
	if err != nil {
		log.Fatalf("could not get streams: %v", err)
	}
	fmt.Println("streams:")
	for _, s := range streams {
		fmt.Printf("- id: %s\n", s.ID)
		fmt.Printf("  status: %s\n", s.Status)
	}
}
