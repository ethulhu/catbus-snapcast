// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package snapcast

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/mdns"
	"go.eth.moe/catbus-snapcast/jsonrpc2"
)

type (
	client struct {
		jsonrpc2.Client

		groupStreamChangedHandler func(string, StreamID)
	}
)

const (
	mdnsService = "_snapcast-jsonrpc._tcp"
)

func Discover() (Client, error) {
	ch := make(chan *mdns.ServiceEntry)
	defer close(ch)

	var serviceEntry *mdns.ServiceEntry
	go func() {
		serviceEntry = <-ch
	}()

	if err := mdns.Lookup(mdnsService, ch); err != nil {
		return nil, fmt.Errorf("could not discover via mDNS: %w", err)
	}
	if serviceEntry == nil {
		return nil, fmt.Errorf("found no %s services", mdnsService)
	}

	addr := fmt.Sprintf("%v:%v", serviceEntry.Host, serviceEntry.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}

// NewClient returns a Snapcast Snapserver client.
func NewClient(conn net.Conn) Client {
	c := &client{
		Client: jsonrpc2.NewClient(conn),
	}

	c.SetNotificationHandler(func(method string, payload json.RawMessage) {
		switch method {
		case groupStreamChanged:
			if c.groupStreamChangedHandler != nil {
				rsp := &groupStreamChangedNotification{}
				if err := json.Unmarshal(payload, rsp); err != nil {
					log.Printf("could not unmarshal %s notification: %v", groupStreamChanged, err)
					return
				}
				c.groupStreamChangedHandler(rsp.ID, rsp.Stream)
			}
		}
	})

	return c
}

func (c *client) SetGroupStreamChangedHandler(f func(string, StreamID)) {
	c.groupStreamChangedHandler = f
}

func (c *client) Host(ctx context.Context) (string, error) {
	rsp := serverGetStatusResponse{}
	if err := c.Call(ctx, serverGetStatus, nil, &rsp); err != nil {
		return "", fmt.Errorf("could not get server status: %w", err)
	}

	name := rsp.Server.Server.Host.Name
	if name == "" {
		name = rsp.Server.Server.Host.IP
	}

	return name, nil
}

func (c *client) Groups(ctx context.Context) (map[string]Group, error) {
	rsp := serverGetStatusResponse{}
	if err := c.Call(ctx, serverGetStatus, nil, &rsp); err != nil {
		return nil, fmt.Errorf("could not get server status: %w", err)
	}

	groups := map[string]Group{}
	for _, g := range rsp.Server.Groups {
		var clients []Speaker
		for _, c := range g.Clients {
			clients = append(clients, Speaker{
				Name:      c.ID,
				Connected: c.Connected,
				Volume: Volume{
					Percent: c.Config.Volume.Percent,
					Muted:   c.Config.Volume.Muted,
				},
			})
		}
		groups[g.ID] = Group{
			ID:       g.ID,
			Name:     g.Name,
			Stream:   g.Stream,
			Speakers: clients,
		}
	}
	return groups, nil
}

func (c *client) Streams(ctx context.Context) ([]Stream, error) {
	rsp := serverGetStatusResponse{}
	if err := c.Call(ctx, serverGetStatus, nil, &rsp); err != nil {
		return nil, fmt.Errorf("could not get server status: %w", err)
	}

	var streams []Stream
	for _, s := range rsp.Server.Streams {
		streams = append(streams, Stream{StreamID(s.ID), s.Status})
	}
	return streams, nil
}

func (c *client) SetGroupName(ctx context.Context, id, name string) error {
	req := groupSetNameRequest{
		ID:   id,
		Name: name,
	}
	rsp := groupSetNameResponse{}
	if err := c.Call(ctx, groupSetName, req, &rsp); err != nil {
		return fmt.Errorf("could not set group name: %w", err)
	}
	if rsp.Name != name {
		return fmt.Errorf("tried to set group name to %v, but got %v instead", name, rsp.Name)
	}
	return nil
}

func (c *client) SetGroupStream(ctx context.Context, groupID string, stream StreamID) error {
	req := groupSetStreamRequest{
		ID:     groupID,
		Stream: stream,
	}
	rsp := groupSetStreamResponse{}
	if err := c.Call(ctx, groupSetStream, req, &rsp); err != nil {
		return fmt.Errorf("could not set stream: %w", err)
	}
	if rsp.Stream != stream {
		return fmt.Errorf("tried to set stream to %v, but got %v instead", stream, rsp.Stream)
	}
	return nil
}
