package jsonrpc2

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type (
	client struct {
		sync.Mutex

		closed bool
		conn   io.ReadWriteCloser

		sequence int

		errorChan            chan error
		requestChan          chan *request
		responseChans        map[int]chan *response
		notificationHandlers map[string]NotificationHandler

		connectHandler func()
		errorHandler   func(error)
	}
)

const (
	protocolVersion = "2.0"
)

// Dial returns a new JSON-RPC 2.0 client.
//
// It is non-blocking, as it handles connecting and re-connecting itself.
// To close the client, explicitly call Close().
func Dial(network, addr string) Client {
	c := &client{
		errorChan:            make(chan error),
		requestChan:          make(chan *request),
		responseChans:        map[int]chan *response{},
		notificationHandlers: map[string]NotificationHandler{},
	}

	go func() {
		for err := range c.errorChan {
			if c.errorHandler != nil {
				c.errorHandler(err)
			}
		}
	}()

	go func() {
		for !c.closed {
			conn, err := net.Dial(network, addr)
			if err != nil {
				c.errorChan <- err
				time.Sleep(5 * time.Second)
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
			c.conn.Close()

			if c.errorHandler != nil {
				go c.errorHandler(nil)
			}
		}
	}()

	return c
}

func (c *client) SetConnectHandler(f func())    { c.connectHandler = f }
func (c *client) SetErrorHandler(f func(error)) { c.errorHandler = f }

func (c *client) Close() {
	c.closed = true
	close(c.errorChan)
	close(c.requestChan)
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
			c.errorChan <- fmt.Errorf("could not receive response: %w", err)
			return
		}

		rsp := &response{}
		if err := json.Unmarshal(data, rsp); err == nil && rsp.ID != nil {
			c.Lock()
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
			c.Lock()
			if f, ok := c.notificationHandlers[noti.Method]; ok {
				go f(noti.Params)
			}
			c.Unlock()
			continue
		}

		c.errorChan <- fmt.Errorf("unknown inbound message: %s", data)
	}
}

func (c *client) writeLoop(connectionClosed chan struct{}) {
	defer c.conn.Close()
	for {
		select {
		case req := <-c.requestChan:
			if req == nil {
				return // c.Close() was called.
			}

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
	if c.closed {
		return ErrClosed
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

func (c *client) SetNotificationHandler(method string, f NotificationHandler) {
	c.Lock()
	defer c.Unlock()

	if f == nil {
		delete(c.notificationHandlers, method)
	} else {
		c.notificationHandlers[method] = f
	}
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
