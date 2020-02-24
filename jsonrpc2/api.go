// Package jsonrpc2 is a minimal JSON-RPC 2.0 client.
//
// Message format
//
// All JSON-RPC 2.0 messages contain the field "jsonrpc" with the value "2.0".
//
// Requests contain an "id" (an arbitrary JSON value), a "method" (a string), and "params" (an arbitrary JSON value).
//
// Responses contain the "id" for their Request, and a "result" (an arbitrary JSON value).
//
// Notfications contain a "method" (a string), and "params" (an arbitrary JSON value).
//
// Example messages
//
// Request:
//
//   { "jsonrpc": "2.0", "id": 3, "method": "add", "params": [1, 2] }
//
// Response:
//
//   { "jsonrpc": "2.0", "id": 3, "result": 19 }
//
// Notification:
//
//   { "jsonrpc": "2.0", "method": "update", "params": { "id": 12345, "name": "toasty" } }
package jsonrpc2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type (
	// Client is a JSON-RPC 2.0 client.
	Client interface {
		// Call performs a JSON-RPC 2.0 method call.
		Call(ctx context.Context, method string, input interface{}, output interface{}) error

		// SetNotificationHandler sets the callback JSON-RPC 2.0 notifications.
		SetNotificationHandler(func(method string, payload json.RawMessage))

		// SetConnectHandler sets the handler that will be called for each successful connection.
		SetConnectHandler(func())

		// SetDisconnectHandler sets the handler that will be called for each disconnection.
		SetDisconnectHandler(func(error))

		// Connect connects and blocks forever, reconnecting as needed.
		Connect()

		// Close closes the client, rendering it inert and stopping reconnect attempts.
		Close()
	}

	// NotificationHandler is a callback function for Client.SetNotificationHandler.
	//
	// The argument is the notification's params, encoded as per encoding/json:
	//
	//   bool, for JSON booleans
	//   float64, for JSON numbers
	//   string, for JSON strings
	//   []interface{}, for JSON arrays
	//   map[string]interface{}, for JSON objects
	//   nil for JSON null
	NotificationHandler func(interface{})

	// RemoteError is an error returned by the server in response to an RPC.
	RemoteError struct {
		Code    int
		Message string
	}
)

var (
	// ErrDisconnected is returned when a Call is cancelled by network disconnection.
	ErrDisconnected = errors.New("disconnected while waiting for response")

	// ErrClosed is returned when trying to call Call on a Client that has been closed.
	ErrClosed = errors.New("client has been closed")
)

func (e *RemoteError) Error() string {
	return fmt.Sprintf("remote error: %s (code %d)", e.Message, e.Code)
}
