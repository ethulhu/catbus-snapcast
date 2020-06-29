// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"log"
	"sort"
	"strings"
	"time"

	"go.eth.moe/catbus"
	"go.eth.moe/catbus-snapcast/config"
	"go.eth.moe/catbus-snapcast/snapcast"
	"go.eth.moe/flag"
)

var (
	configPath = flag.Custom("config-path", "", "path to config.json", flag.RequiredString)
)

var host string

func main() {
	flag.Parse()

	configPath := (*configPath).(string)

	config, err := config.ParseFile(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	snapserver, err := snapcast.Discover()
	if err != nil {
		log.Fatal(err)
	}

	catbusOptions := catbus.ClientOptions{
		DisconnectHandler: func(_ catbus.Client, err error) {
			log.Printf("disconnected from MQTT broker %s: %v", config.BrokerURI, err)
		},
		ConnectHandler: func(broker catbus.Client) {
			log.Printf("connected to MQTT broker %s", config.BrokerURI)

			if err := broker.Subscribe(config.Topics.Input, setInput(snapserver, config.Snapcast.GroupID)); err != nil {
				log.Printf("could not subscribe to %v: %v", config.Topics.Input, err)
			}
		},
	}
	broker := catbus.NewClient(config.BrokerURI, catbusOptions)

	snapserver.SetConnectHandler(func(snapserver snapcast.Client) {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		var err error
		host, err = snapserver.Host(ctx)
		if err != nil {
			log.Printf("could not get Snapserver host: %v", err)
			return
		}
		log.Printf("connected to Snapserver: %v", host)

		streams, err := snapserver.Streams(ctx)
		if err != nil {
			log.Printf("could not list Snapserver streams: %v", err)
			return
		}

		streamNames := make([]string, len(streams))
		for i, stream := range streams {
			streamNames[i] = string(stream.ID)
		}
		sort.Strings(streamNames)

		if err := broker.Publish(config.Topics.InputValues, catbus.Retain, strings.Join(streamNames, "\n")); err != nil {
			log.Printf("could not publish stream values: %v", err)
		}

		groups, err := snapserver.Groups(ctx)
		if err != nil {
			log.Printf("could not get current Snapserver groups: %v", err)
			return
		}
		for _, group := range groups {
			if group.ID != config.Snapcast.GroupID {
				return
			}
			log.Printf("publishing stream value %q", group.Stream)
			if err := broker.Publish(config.Topics.Input, catbus.Retain, string(group.Stream)); err != nil {
				log.Printf("could not publish stream value %q: %v", group.Stream, err)
			}
		}
	})
	snapserver.SetDisconnectHandler(func(err error) {
		log.Printf("disconnected from Snapserver %q: %v", host, err)
	})
	snapserver.SetGroupStreamChangedHandler(func(groupID string, stream snapcast.StreamID) {
		if groupID != config.Snapcast.GroupID {
			return
		}

		log.Printf("publishing stream value %q", stream)
		if err := broker.Publish(config.Topics.Input, catbus.Retain, string(stream)); err != nil {
			log.Printf("could not publish stream value %q: %v", stream, err)
		}
	})

	go func() {
		log.Printf("connecting to Snapserver")
		snapserver.Connect()
	}()

	log.Printf("connecting to MQTT broker %v", config.BrokerURI)
	if err := broker.Connect(); err != nil {
		log.Fatalf("could not connect to Catbus: %v", err)
	}
}

func setInput(snapserver snapcast.Client, groupID string) catbus.MessageHandler {
	return func(_ catbus.Client, msg catbus.Message) {
		stream := snapcast.StreamID(msg.Payload)

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		groups, err := snapserver.Groups(ctx)
		if err != nil {
			log.Printf("could not get existing groups: %v", err)
			return
		}

		group, ok := groups[groupID]
		if !ok {
			log.Print("could not find group")
			return
		}

		if group.Stream == stream {
			// Don't set it twice.
			return
		}

		if err := snapserver.SetGroupStream(ctx, groupID, stream); err != nil {
			log.Printf("could not set stream to %q: %v", msg.Payload, err)
			return
		}
		log.Printf("set stream to %q", msg.Payload)
	}
}
