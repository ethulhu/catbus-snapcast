// Package snapcast is a Snapcast client.
//
// It follows the API as described in https://github.com/badaix/snapcast/blob/master/doc/json_rpc_api/v2_0_0.md.
package snapcast

import "context"

type (
	// Client is a Snapcast Snapserver RPC client.
	Client interface {
		// Groups returns the list of groups managed by the Snapserver.
		Groups(context.Context) ([]Group, error)

		// Streams returns the list of streams managed by the Snapserver.
		Streams(context.Context) ([]Stream, error)

		// SetStream sets a given Group's stream to the given Stream.
		SetStream(ctx context.Context, groupID string, stream Stream) error

		// SetConnectHandler sets the handler that will be called for each successful connection.
		SetConnectHandler(func())
		// SetErrorHandler sets the handler that will be called for connection errors (not RPC errors).
		SetErrorHandler(func(error))

		// Close closes the client, rendering it inert and stopping reconnect attempts.
		Close()
	}

	// Group represents a group of speakers.
	Group struct {
		ID     string
		Name   string
		Stream Stream

		Speakers []Speaker
	}

	// Speaker represents a speaker / sink / Snapclient.
	Speaker struct {
		Name      string
		Connected bool
		Volume    Volume
	}

	// Stream is a stream identifier.
	Stream string

	// Volume is a speaker's volume.
	Volume struct {
		Percent int
		Muted   bool
	}
)
