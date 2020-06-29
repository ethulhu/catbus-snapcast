// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package jsonrpc2

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type (
	client struct {
		sync.Mutex

		conn io.ReadWriteCloser

		sequence int

		errorChan           chan error
		requestChan         chan *request
		responseChans       map[int]chan *response
		notificationHandler func(string, json.RawMessage)

		disconnectHandler func(error)
	}
)

const (
	protocolVersion = "2.0"
)

// NewClient returns a new JSON-RPC 2.0 client.
func NewClient(conn net.Conn) Client {
	c := &client{
		conn: conn,

		errorChan:     make(chan error),
		requestChan:   make(chan *request),
		responseChans: map[int]chan *response{},
	}

	connectionClosed := make(chan struct{})
	go c.readLoop(connectionClosed)
	go c.writeLoop(connectionClosed)

	return c
}

func (c *client) Close() error {
	return c.conn.Close()
}
func (c *client) Wait() error {
	return <-c.errorChan
}

func (c *client) SetNotificationHandler(f func(string, json.RawMessage)) {
	c.notificationHandler = f
}

func (c *client) readLoop(connectionClosed chan struct{}) {
	defer close(connectionClosed)
	defer close(c.errorChan)
	defer func() {
		c.Lock()
		defer c.Unlock()
		for id, ch := range c.responseChans {
			close(ch)
			delete(c.responseChans, id)
		}
	}()

	reader := bufio.NewReader(c.conn)
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			// Non-blocking send.
			select {
			case c.errorChan <- err:
			default:
			}
			return
		}

		rsp := &response{}
		if err := json.Unmarshal(data, rsp); err == nil && rsp.ID != nil {
			c.Lock()
			// if noöne requested it, throw it away.
			if ch, ok := c.responseChans[*rsp.ID]; ok {
				ch <- rsp
				close(ch)
				delete(c.responseChans, *rsp.ID)
			}
			c.Unlock()
			continue
		}

		noti := notification{}
		if err := json.Unmarshal(data, &noti); err == nil && noti.Method != "" {
			if c.notificationHandler != nil {
				go c.notificationHandler(noti.Method, noti.Params)
			}
			continue
		}

		log.Printf("unknown inbound message: %s", data)
	}
}

func (c *client) writeLoop(connectionClosed chan struct{}) {
	defer c.conn.Close()
	for {
		select {
		case req := <-c.requestChan:
			packet, err := json.Marshal(req)
			if err != nil {
				panic(fmt.Sprintf("could not marshal request: %v", err))
			}
			packet = append(packet, []byte("\r\n")...)

			if _, err := c.conn.Write(packet); err != nil {
				return
			}
		case <-connectionClosed:
			return
		}
	}
}

func (c *client) Call(ctx context.Context, method string, params interface{}, result interface{}) error {
	id, ch := c.newRequest()

	req := &request{
		ProtocolVersion: protocolVersion,
		ID:              id,
		Method:          method,
		Params:          params,
	}

	c.requestChan <- req

	select {
	case rsp := <-ch:
		if rsp == nil {
			return ErrDisconnected
		}
		if rsp.Error != nil {
			return RemoteError{
				Code:    rsp.Error.Code,
				Message: rsp.Error.Message,
			}
		}
		if result != nil {
			if err := json.Unmarshal(rsp.Result, result); err != nil {
				return fmt.Errorf("could not unmarshal result payload: %w", err)
			}
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (c *client) newRequest() (int, <-chan *response) {
	c.Lock()
	defer c.Unlock()

	id := c.sequence
	c.sequence++

	ch := make(chan *response)
	c.responseChans[id] = ch

	return id, ch
}
