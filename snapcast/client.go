package snapcast

import (
	"context"
	"fmt"

	"github.com/ethulhu/mqtt-snapcast-bridge/jsonrpc2"
)

type (
	client struct {
		jsonrpc2.Client
	}
)

// Dial returns a Snapcast Snapserver client.
//
// It is non-blocking, as it handles connecting and re-connecting itself.
// To close the client, explicitly call Close().
func Dial(network, addr string) Client {
	return &client{
		jsonrpc2.Dial(network, addr),
	}
}

func (c *client) Groups(ctx context.Context) ([]Group, error) {
	rsp := serverGetStatusResponse{}
	if err := c.Client.Call(ctx, serverGetStatus, nil, &rsp); err != nil {
		return nil, fmt.Errorf("could not get server status: %w", err)
	}

	var groups []Group
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
		groups = append(groups, Group{
			ID:       g.ID,
			Name:     g.Name,
			Stream:   g.Stream,
			Speakers: clients,
		})
	}
	return groups, nil
}

func (c *client) Streams(ctx context.Context) ([]Stream, error) {
	rsp := serverGetStatusResponse{}
	if err := c.Client.Call(ctx, serverGetStatus, nil, &rsp); err != nil {
		return nil, fmt.Errorf("could not get server status: %w", err)
	}

	var streams []Stream
	for _, s := range rsp.Server.Streams {
		streams = append(streams, Stream(s.ID))
	}
	return streams, nil
}

func (c *client) SetGroupName(ctx context.Context, id, name string) error {
	req := groupSetNameRequest{
		ID:   id,
		Name: name,
	}
	rsp := groupSetNameResponse{}
	if err := c.Client.Call(ctx, groupSetName, req, &rsp); err != nil {
		return fmt.Errorf("could not set group name: %w", err)
	}
	if rsp.Name != name {
		return fmt.Errorf("tried to set group name to %v, but got %v instead", name, rsp.Name)
	}
	return nil
}

func (c *client) SetStream(ctx context.Context, groupID string, stream Stream) error {
	req := groupSetStreamRequest{
		ID:     groupID,
		Stream: stream,
	}
	rsp := groupSetStreamResponse{}
	if err := c.Client.Call(ctx, groupSetStream, req, &rsp); err != nil {
		return fmt.Errorf("could not set stream: %w", err)
	}
	if rsp.Stream != stream {
		return fmt.Errorf("tried to set stream to %v, but got %v instead", stream, rsp.Stream)
	}
	return nil
}
