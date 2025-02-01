package common

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

// Livestream coordinates the liveview process for a Blink (Immedia Semiconductor) camera.
// It starts a liveview session, polls the liveview command to keep the connection alive, and connects to the liveview server.
//
// Refer to TCPStream for the connection process details and output methods.
//
// region: the region to use for the API URL (e.g. "u011")
//
// token: the API auth token to use for the API requests
//
// deviceType: the type of device to use for the liveview path
//
// accountId: the account ID that the camera belongs to
//
// networkId: the network ID that the camera is on
//
// cameraId: the ID of the camera to start the liveview session for
//
// Example: Livestream("u011", "example_token", "camera", 1234, 5678, 9012)
func Livestream(region string, token string, deviceType string, accountId int, networkId int, cameraId int) {
	baseUrl := GetApiUrl(region)
	liveViewPath, err := GetLiveviewPath(deviceType)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Tell Blink we want to start a liveview session
	resp, err := BeginLiveview(fmt.Sprintf(liveViewPath, baseUrl, accountId, networkId, cameraId), token)
	if err != nil {
		fmt.Println(err)
		return
	} else if resp == nil || resp.CommandId == 0 {
		fmt.Println("Error sending liveview command", resp)
		return
	}

	// Poll the liveview command to keep the connection alive
	ctx, cancelCtx := context.WithCancel(context.Background())
	go PollCommand(ctx, fmt.Sprintf("%s/network/%d/command/%d", baseUrl, networkId, resp.CommandId), token, resp.PollingInterval)
	defer StopCommand(fmt.Sprintf("%s/network/%d/command/%d/done", baseUrl, networkId, resp.CommandId), token)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("SIGINT: Stopping...")
		cancelCtx()
	}()

	// Connect to the liveview server
	connectionDetails, err := ParseConnectionString(resp.Server)
	if err != nil {
		fmt.Println(err)
		return
	}

	TCPStream(ctx, *connectionDetails)
}
