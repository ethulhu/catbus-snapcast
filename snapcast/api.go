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
		// Connect connects to the Snapserver and blocks forever, reconnecting as needed.
		Connect()
		// SetConnectHandler sets the handler that will be called for each successful connection.
		SetConnectHandler(func(Client))
		// SetDisconnectHandler sets the handler that will be called after each disconnection.
		SetDisconnectHandler(func(error))

		// Groups returns the list of groups managed by the Snapserver.
		Groups(context.Context) ([]Group, error)

		// Streams returns the list of streams managed by the Snapserver.
		Streams(context.Context) ([]Stream, error)

		// SetGroupStream sets a given Group's stream to the given Stream.
		SetGroupStream(ctx context.Context, groupID string, stream StreamID) error

		// SetGroupStreamChangedHandler sets the handler that is called when a group's stream changes.
		SetGroupStreamChangedHandler(func(groupID string, stream StreamID))
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
