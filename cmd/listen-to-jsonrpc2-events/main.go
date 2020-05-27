// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/ethulhu/catbus-snapcast/jsonrpc2"
)

var (
	host = flag.String("host", "", "host of Snapserver")
	port = flag.Uint("port", 0, "port of Snapserver")
)

func main() {
	flag.Parse()
	if *host == "" || *port == 0 {
		log.Fatal("must set -host and -port")
	}
	addr := fmt.Sprintf("%v:%v", *host, *port)

	client := jsonrpc2.NewClient(addr)

	client.SetConnectHandler(func() {
		log.Print("connected")
	})
	client.SetDisconnectHandler(func(err error) {
		log.Printf("disconnected: %v", err)
	})

	client.SetNotificationHandler(func(method string, payload json.RawMessage) {
		fmt.Printf("method %q, payload %s\n", method, payload)
	})

	client.Connect()
}
