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
	"time"
)

type (
	client struct {
		sync.Mutex

		addr string

		connected bool
		conn      io.ReadWriteCloser

		sequence int

		requestChan         chan *request
		responseChans       map[int]chan *response
		notificationHandler func(string, json.RawMessage)

		connectHandler    func()
		disconnectHandler func(error)
	}
)

const (
	protocolVersion = "2.0"

	reconnectionDelay = 5 * time.Second
)

// Dial returns a new JSON-RPC 2.0 client.
//
// It is non-blocking, as it handles connecting and re-connecting itself.
// To close the client, explicitly call Close().
func NewClient(addr string) Client {
	return &client{
		addr: addr,

		requestChan:   make(chan *request),
		responseChans: map[int]chan *response{},
	}
}

func (c *client) SetConnectHandler(f func())         { c.connectHandler = f }
func (c *client) SetDisconnectHandler(f func(error)) { c.disconnectHandler = f }
func (c *client) SetNotificationHandler(f func(string, json.RawMessage)) {
	c.notificationHandler = f
}

func (c *client) Connect() {
	c.connected = true

	for {
		conn, err := net.Dial("tcp", c.addr)
		if err != nil {
			log.Printf("could not connect, sleeping for %v and trying again", reconnectionDelay)
			time.Sleep(reconnectionDelay)
			continue
		}
		c.conn = conn

		var wg sync.WaitGroup
		wg.Add(2)

		connectionClosed := make(chan struct{})
		go func() {
			c.readLoop(connectionClosed)
			wg.Done()
		}()
		go func() {
			c.writeLoop(connectionClosed)
			wg.Done()
		}()

		if c.connectHandler != nil {
			go c.connectHandler()
		}

		wg.Wait()
	}
}

func (c *client) readLoop(connectionClosed chan struct{}) {
	defer close(connectionClosed)
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
			go c.disconnectHandler(err)
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
	if !c.connected {
		panic("called jsonrpc2.Client.Call before jsonrpc2.Client.Connect")
	}

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