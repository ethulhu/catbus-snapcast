package jsonrpc2

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
)

type (
	client struct {
		conn io.ReadWriter
	}

	request struct {
		ProtocolVersion string      `json:"jsonrpc"`
		ID              int         `json:"id"`
		Method          string      `json:"method"`
		Params          interface{} `json:"params,omitempty"`
	}
	response struct {
		ProtocolVersion string          `json:"jsonrpc"`
		ID              int             `json:"id"`
		Result          json.RawMessage `json:"result,omitempty"`
		Error           *responseError  `json:"error,omitempty"`
	}
	responseError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)

const (
	protocolVersion = "2.0"
)

// NewClient returns a new JSON-RPC 2.0 client.
// Users are responsible for closing the connection.
func NewClient(conn io.ReadWriter) Client {
	return &client{
		conn: conn,
	}
}

func (c *client) Call(method string, params interface{}, result interface{}) error {
	req := request{
		ProtocolVersion: protocolVersion,
		ID:              rand.Int(),
		Method:          method,
		Params:          params,
	}

	packet, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("could not marshal request: %w", err)
	}
	packet = append(packet, []byte("\r\n")...)

	if _, err := c.conn.Write(packet); err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}

	reader := bufio.NewReader(c.conn)
	data, err := reader.ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("could not receive response: %w", err)
	}

	rsp := response{}
	if err := json.Unmarshal(data, &rsp); err != nil {
		log.Printf("response: %s", data)
		return fmt.Errorf("could not unmarshal response: %w", err)
	}

	if rsp.Error != nil {
		return &RemoteError{
			Code:    rsp.Error.Code,
			Message: rsp.Error.Message,
		}
	}

	if result != nil {
		if err := json.Unmarshal(rsp.Result, result); err != nil {
			return fmt.Errorf("could not unmarshal result payload: %w", err)
		}
	}

	return nil
}
