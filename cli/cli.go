package cli

import (
	"blink-liveview-websocket/common"
	"fmt"
)

func Run(region *string, token *string, accountId *int, networkId *int, cameraId *int) {
	fmt.Println("Running the CLI")

	if *region == "" {
		panic("Region is required")
	}

	if *token == "" {
		panic("Token is required")
	}

	if *accountId == 0 {
		panic("Account ID is required")
	}

	if *networkId == 0 {
		panic("Network ID is required")
	}

	if *cameraId == 0 {
		panic("Camera ID is required")
	}

	baseUrl := common.GetApiUrl(*region)

	common.Livestream(baseUrl, *token, *accountId, *networkId, *cameraId)
}
