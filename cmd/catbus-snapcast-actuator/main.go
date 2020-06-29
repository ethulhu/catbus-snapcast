// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"log"
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

	catbusOptions := catbus.ClientOptions{
		DisconnectHandler: func(_ catbus.Client, err error) {
			log.Printf("disconnected from MQTT broker %s: %v", config.BrokerURI, err)
		},
		ConnectHandler: func(broker catbus.Client) {
			log.Printf("connected to MQTT broker %s", config.BrokerURI)

			if err := broker.Subscribe(config.Topics.Input, setInput(config)); err != nil {
				log.Printf("could not subscribe to %v: %v", config.Topics.Input, err)
			}
		},
	}
	broker := catbus.NewClient(config.BrokerURI, catbusOptions)

	log.Printf("connecting to Catbus %v", config.BrokerURI)
	if err := broker.Connect(); err != nil {
		log.Fatalf("could not connect to Catbus: %v", err)
	}
}

func setInput(config *config.Config) catbus.MessageHandler {
	return func(_ catbus.Client, msg catbus.Message) {
		stream := snapcast.StreamID(msg.Payload)

		snapserver, err := snapcast.Discover()
		if err != nil {
			log.Printf("could not connect to Snapserver: %v", err)
			return
		}
		defer snapserver.Close()

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		groups, err := snapserver.Groups(ctx)
		if err != nil {
			log.Printf("could not get existing groups: %v", err)
			return
		}

		group, ok := groups[config.Snapcast.GroupID]
		if !ok {
			log.Print("could not find group")
			return
		}

		if group.Stream == stream {
			// Don't set it twice.
			return
		}

		if err := snapserver.SetGroupStream(ctx, config.Snapcast.GroupID, stream); err != nil {
			log.Printf("could not set stream to %q: %v", msg.Payload, err)
			return
		}
		log.Printf("set stream to %q", msg.Payload)
	}
}
