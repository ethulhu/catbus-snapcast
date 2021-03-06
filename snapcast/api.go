// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

// Package snapcast is a Snapcast client.
//
// It follows the API as described in https://github.com/badaix/snapcast/blob/master/doc/json_rpc_api/v2_0_0.md.
package snapcast

import "context"

type (
	// Client is a Snapcast Snapserver RPC client.
	Client interface {
		// Host returns either the IP or hostname of the Snapserver.
		Host(context.Context) (string, error)

		// Groups returns a map of groups by group ID managed by the Snapserver.
		Groups(context.Context) (map[string]Group, error)

		// Streams returns the list of streams managed by the Snapserver.
		Streams(context.Context) ([]Stream, error)

		// SetGroupStream sets a given Group's stream to the given Stream.
		SetGroupStream(ctx context.Context, groupID string, stream StreamID) error

		// SetGroupStreamChangedHandler sets the handler that is called when a group's stream changes.
		SetGroupStreamChangedHandler(func(groupID string, stream StreamID))

		// Wait blocks until the connection fails.
		Wait() error

		// Close closes the connection.
		Close() error
	}

	// Group represents a group of speakers.
	Group struct {
		ID     string
		Name   string
		Stream StreamID

		Speakers []Speaker
	}

	// Speaker represents a speaker / sink / Snapclient.
	Speaker struct {
		Name      string
		Connected bool
		Volume    Volume
	}

	// StreamID is a stream identifier.
	StreamID string

	// Stream represents a stream.
	Stream struct {
		ID     StreamID
		Status string
	}

	// Volume is a speaker's volume.
	Volume struct {
		Percent int
		Muted   bool
	}
)

const (
	DefaultPort = 1705
)
