package main

import (
	"blink-liveview-websocket/cli"
	"flag"
	"fmt"
)

var (
	region    = flag.String("region", "", "Blink API region")
	token     = flag.String("token", "", "Blink API token")
	accountId = flag.Int("account-id", 0, "Blink account ID")
	networkId = flag.Int("network-id", 0, "Blink network ID")
	cameraId  = flag.Int("camera-id", 0, "Blink camera ID")
)

func main() {
	// cmd --usage
	flag.Usage = func() {
		fmt.Print("Usage: blink-liveview-websocket [options]\n\nOptions:\n")

		flag.PrintDefaults()
	}

	flag.Parse()

	cli.Run(region, token, accountId, networkId, cameraId)
}
