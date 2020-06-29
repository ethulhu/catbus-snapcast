// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"log"

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
		snapserverAddr := fmt.Sprintf("%v:%v", *snapserverHost, *snapserverPort)

		client = snapcast.NewClient(snapserverAddr)
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
	client.SetDisconnectHandler(func(err error) {
		log.Printf("disconnected: %v", err)
	})
	client.SetConnectHandler(func(client snapcast.Client) {
		log.Print("connected!")
	})

	client.Connect()
}
