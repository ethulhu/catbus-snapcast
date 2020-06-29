// SPDX-FileCopyrightText: 2020 Ethel Morgan
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"

	"go.eth.moe/catbus-snapcast/jsonrpc2"
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
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("could not dial %v: %v", addr, err)
	}
	defer conn.Close()

	log.Print("connected")

	client := jsonrpc2.NewClient(conn)

	client.SetNotificationHandler(func(method string, payload json.RawMessage) {
		fmt.Printf("method %q, payload %s\n", method, payload)
	})

	if err := client.Wait(); err != nil {
		log.Fatalf("disconnected from %q: %v", addr, err)
	}
}
