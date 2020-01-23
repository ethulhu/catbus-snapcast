package snapcast

import (
	"fmt"
	"io"

	"github.com/ethulhu/go-snapcast/jsonrpc2"
)

type (
	client struct {
		rpcClient jsonrpc2.Client
	}
)

// NewClient returns a Snapcast Snapserver client.
func NewClient(conn io.ReadWriter) Client {
	return &client{
		rpcClient: jsonrpc2.NewClient(conn),
	}
}

func (c *client) Groups() ([]Group, error) {
	rsp := serverGetStatusResponse{}
	if err := c.rpcClient.Call(serverGetStatus, nil, &rsp); err != nil {
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

func (c *client) Streams() ([]Stream, error) {
	rsp := serverGetStatusResponse{}
	if err := c.rpcClient.Call(serverGetStatus, nil, &rsp); err != nil {
		return nil, fmt.Errorf("could not get server status: %w", err)
	}

	var streams []Stream
	for _, s := range rsp.Server.Streams {
		streams = append(streams, Stream(s.ID))
	}
	return streams, nil
}

func (c *client) SetGroupName(id, name string) error {
	req := groupSetNameRequest{
		ID:   id,
		Name: name,
	}
	rsp := groupSetNameResponse{}
	if err := c.rpcClient.Call(groupSetName, req, &rsp); err != nil {
		return fmt.Errorf("could not set group name: %w", err)
	}
	if rsp.Name != name {
		return fmt.Errorf("tried to set group name to %v, but got %v instead", name, rsp.Name)
	}
	return nil
}

func (c *client) SetStream(groupID string, stream Stream) error {
	req := groupSetStreamRequest{
		ID:     groupID,
		Stream: stream,
	}
	rsp := groupSetStreamResponse{}
	if err := c.rpcClient.Call(groupSetStream, req, &rsp); err != nil {
		return fmt.Errorf("could not set stream: %w", err)
	}
	if rsp.Stream != stream {
		return fmt.Errorf("tried to set stream to %v, but got %v instead", stream, rsp.Stream)
	}
	return nil
}
