package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/ethulhu/mqtt-snapcast-bridge/mqtt"
	"github.com/ethulhu/mqtt-snapcast-bridge/snapcast"
)

var (
	configPath = flag.String("config-path", "", "path to config.json")
)

func main() {
	flag.Parse()

	if *configPath == "" {
		log.Fatal("must set --config-path")
	}

	config, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	snapserverAddr := fmt.Sprintf("%v:%v", config.SnapserverHost, config.SnapserverPort)
	snapserver := snapcast.NewClient(snapserverAddr)

	brokerURI := mqtt.URI(config.BrokerHost, config.BrokerPort)
	brokerOptions := mqtt.NewClientOptions()
	brokerOptions.AddBroker(brokerURI)
	brokerOptions.SetAutoReconnect(true)
	brokerOptions.SetClientID("mqtt-snapcast-bridge")
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
		log.Printf("connected to Snapserver %s", snapserverAddr)

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
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
		log.Printf("disconnected from Snapserver %s: %v", snapserverAddr, err)
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

	log.Printf("connecting to Snapserver %v", snapserverAddr)
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
