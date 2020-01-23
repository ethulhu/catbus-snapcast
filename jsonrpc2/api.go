// Package jsonrpc2 is a minimal JSON-RPC 2.0 client.
package jsonrpc2

import "fmt"

type (
	// Client is a JSON-RPC 2.0 client.
	// TODO: support Notifications.
	Client interface {
		Call(method string, input interface{}, output interface{}) error
	}

	// RemoteError is an error returned by the server in response to an RPC.
	RemoteError struct {
		Code    int
		Message string
	}
)

func (e *RemoteError) Error() string {
	return fmt.Sprintf("remote error: %s (code %d)", e.Message, e.Code)
}
