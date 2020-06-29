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
	"path"
	"sort"
	"strings"
	"time"

	"github.com/ethulhu/catbus-snapcast/mqtt"
	"github.com/ethulhu/catbus-snapcast/snapcast"
)

var (
	configPath = flag.String("config-path", "", "path to config.json")

	mqttClientID = flag.String("mqtt-client-id", "catbus-bridge-snapcast", "the client ID passed to the MQTT broker")
)

var host string

func main() {
	flag.Parse()

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "must set --config-path")
		flag.Usage()
		os.Exit(2)
	}

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var snapserver snapcast.Client
	if config.SnapserverHost != "" {
		snapserverPort := config.SnapserverPort
		if snapserverPort == 0 {
			snapserverPort = snapcast.DefaultPort
		}

		snapserverAddr := fmt.Sprintf("%v:%v", config.SnapserverHost, snapserverPort)

		snapserver = snapcast.NewClient(snapserverAddr)
	} else {
		var err error
		snapserver, err = snapcast.Discover()
		if err != nil {
			log.Fatal(err)
		}
	}

	brokerURI := mqtt.URI(config.BrokerHost, config.BrokerPort)
	brokerOptions := mqtt.NewClientOptions()
	brokerOptions.AddBroker(brokerURI)
	brokerOptions.SetAutoReconnect(true)
	brokerOptions.SetClientID(*mqttClientID)
	brokerOptions.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		log.Printf("disconnected from MQTT broker %s: %v", brokerURI, err)
	})
	brokerOptions.SetOnConnectHandler(func(broker mqtt.Client) {
		log.Printf("connected to MQTT broker %s", brokerURI)

		token := broker.Subscribe(config.TopicInput, mqtt.AtLeastOnce, setInput(snapserver, config.SnapcastGroupID))
		if err := token.Error(); err != nil {
			log.Printf("could not subscribe to %v: %v", config.TopicInput, err)
		}
	})
	broker := mqtt.NewClient(brokerOptions)

	snapserver.SetConnectHandler(func(snapserver snapcast.Client) {
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

		var err error
		host, err = snapserver.Host(ctx)
		if err != nil {
			log.Printf("could not get Snapserver host: %v", err)
			return
		}
		log.Printf("connected to Snapserver %q", host)

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

		topicValues := path.Join(config.TopicInput, "values")
		token := broker.Publish(topicValues, mqtt.AtLeastOnce, mqtt.Retain, strings.Join(streamNames, "\n"))
		if err := token.Error(); err != nil {
			log.Printf("could not publish stream values: %v", err)
		}

		groups, err := snapserver.Groups(ctx)
		if err != nil {
			log.Printf("could not get current Snapserver groups: %v", err)
			return
		}
		for _, group := range groups {
			if group.ID != config.SnapcastGroupID {
				return
			}
			log.Printf("publishing stream value %q", group.Stream)
			token := broker.Publish(config.TopicInput, mqtt.AtLeastOnce, mqtt.Retain, string(group.Stream))
			if err := token.Error(); err != nil {
				log.Printf("could not publish stream value %q: %v", group.Stream, err)
			}
		}
	})
	snapserver.SetDisconnectHandler(func(err error) {
		log.Printf("disconnected from Snapserver %q: %v", host, err)
	})
	snapserver.SetGroupStreamChangedHandler(func(groupID string, stream snapcast.StreamID) {
		if groupID != config.SnapcastGroupID {
			return
		}

		log.Printf("publishing stream value %q", stream)
		token := broker.Publish(config.TopicInput, mqtt.AtLeastOnce, mqtt.Retain, string(stream))
		if err := token.Error(); err != nil {
			log.Printf("could not publish stream value %q: %v", stream, err)
		}
	})

	log.Printf("connecting to MQTT broker %v", brokerURI)
	_ = broker.Connect()

	log.Printf("connecting to Snapserver")
	snapserver.Connect()
}

func setInput(snapserver snapcast.Client, groupID string) mqtt.MessageHandler {
	return func(_ mqtt.Client, msg mqtt.Message) {
		log.Printf("setting stream to %q", msg.Payload())
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		if err := snapserver.SetGroupStream(ctx, groupID, snapcast.StreamID(msg.Payload())); err != nil {
			log.Printf("could not set stream to %q: %v", msg.Payload(), err)
			return
		}
	}
}
