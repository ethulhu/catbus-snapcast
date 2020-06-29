// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
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

	client.SetGroupStreamChangedHandler(func(groupID string, stream snapcast.StreamID) {
		log.Printf("group %v changed to stream %v", groupID, stream)
	})

	if err := client.Wait(); err != nil {
		log.Fatalf("disconnected from Snapserver: %v", err)
	}
}
