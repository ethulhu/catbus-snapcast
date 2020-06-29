// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ethulhu/catbus-snapcast/snapcast"
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
		snapserverAddr := fmt.Sprintf("%v:%v", *snapserverHost, *snapserverPort)

		client = snapcast.NewClient(snapserverAddr)
	} else {
		var err error
		client, err = snapcast.Discover()
		if err != nil {
			log.Fatal(err)
		}
	}

	client.SetConnectHandler(func(client snapcast.Client) {
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

		os.Exit(0)
	})
	client.SetDisconnectHandler(func(err error) {
		log.Printf("disconnected: %v", err)
	})

	client.Connect()

}
