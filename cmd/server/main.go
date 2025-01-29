package main

import (
	"blink-liveview-websocket/server"
	"flag"
	"fmt"
)

var (
	address = flag.String("address", ":8080", "HTTP server address")
)

func main() {
	// cmd --usage
	flag.Usage = func() {
		fmt.Print("Usage: blink-liveview-websocket [options]\n\nOptions:\n")

		flag.PrintDefaults()
	}

	flag.Parse()

	server.Run(address)
}
