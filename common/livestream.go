package common

import (
	"context"
	"fmt"
	"log"
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
		log.Println("Error getting liveview path", err)
		return
	}

	// Tell Blink we want to start a liveview session
	resp, err := BeginLiveview(fmt.Sprintf(liveViewPath, baseUrl, accountId, networkId, cameraId), token)
	if err != nil {
		log.Println("Error starting liveview session", err)
		return
	} else if resp == nil || resp.CommandId == 0 {
		log.Println("Error sending liveview command", resp)
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
		log.Println("Received SIGINT. Stopping liveview session.")
		cancelCtx()
	}()

	// Get the connection details
	connectionDetails, err := ParseConnectionString(resp.Server)
	if err != nil {
		log.Println("Error parsing connection string", err)
		return
	}

	// Connect to the liveview server
	if err := TCPStream(ctx, *connectionDetails); err != nil {
		log.Println("TCPStream error", err)
	}
}
