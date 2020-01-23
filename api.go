// Package snapcast is a Snapcast client.
package snapcast

type (
	// Client is a Snapcast Snapserver RPC client.
	Client interface {
		// Groups returns the list of groups managed by the Snapserver.
		Groups() ([]Group, error)

		// Streams returns the list of streams managed by the Snapserver.
		Streams() ([]Stream, error)

		// SetStream sets a given Group's stream to the given Stream.
		SetStream(groupID string, stream Stream) error
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
