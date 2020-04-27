package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ethulhu/mqtt-snapcast-bridge/snapcast"
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

	client := snapcast.NewClient(snapserverAddr)

	client.SetConnectHandler(func(client snapcast.Client) {
		log.Print("connected")

		ctx := context.Background()
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
			fmt.Printf("- %s\n", s)
		}

		os.Exit(0)

	})
	client.SetDisconnectHandler(func(err error) {
		log.Printf("disconnected: %v", err)
	})

	client.Connect()
}
