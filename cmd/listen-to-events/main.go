package main

import (
	"fmt"
	"log"

	"github.com/ethulhu/go-snapcast/jsonrpc2"
)

func main() {
	client := jsonrpc2.Dial("tcp", "valkyrie.local:1705")
	defer client.Close()

	client.SetConnectHandler(func() {
		log.Print("connected")
	})
	client.SetErrorHandler(func(err error) {
		log.Printf("client error: %v", err)
	})

	client.SetNotificationHandler("Group.OnStreamChanged", func(params interface{}) {
		fmt.Printf("%+v\n", params)
	})

	select {}
}
