package cli

import (
	"blink-liveview-websocket/common"
	"fmt"
	"os"
)

func Run(region *string, token *string, deviceType *string, accountId *int, networkId *int, cameraId *int) {
	if *region == "" {
		fmt.Fprintf(os.Stderr, "No region parameter provided. Please specify via --region=<region>\n")
		os.Exit(1)
	}

	if *token == "" {
		fmt.Fprintf(os.Stderr, "No token parameter provided. Please specify via --token=<token>\n")
		os.Exit(1)
	}

	if *deviceType == "" {
		fmt.Fprintf(os.Stderr, "No device type parameter provided. Please specify via --device-type=<device-type>\n")
		os.Exit(1)
	}

	if *accountId == 0 {
		fmt.Fprintf(os.Stderr, "No account ID parameter provided. Please specify via --account-id=<account-id>\n")
		os.Exit(1)
	}

	if *networkId == 0 {
		fmt.Fprintf(os.Stderr, "No network ID parameter provided. Please specify via --network-id=<network-id>\n")
		os.Exit(1)
	}

	if *cameraId == 0 {
		fmt.Fprintf(os.Stderr, "No camera ID parameter provided. Please specify via --camera-id=<camera-id>\n")
		os.Exit(1)
	}

	common.Livestream(*region, *token, *deviceType, *accountId, *networkId, *cameraId)
}
