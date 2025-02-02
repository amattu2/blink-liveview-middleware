package common

import (
	"context"
	"fmt"
	"log"
)

type AccountDetails struct {
	// Region to use for the API URL (e.g. "u011")
	Region string
	// API auth token to use for the API requests
	Token string
	// Type of device to use for the liveview path
	DeviceType string
	// Account ID that the camera belongs to
	AccountId int
	// Network ID that the camera is on
	NetworkId int
	// ID of the camera to start the liveview session for
	CameraId int
}

// Livestream coordinates the liveview process for a Blink (Immedia Semiconductor) camera.
// It starts a liveview session, polls the liveview command to keep the connection alive, and connects to the liveview server.
// Returns an error if any of the steps fail.
//
// Refer to TCPStream for the connection process details and output methods.
//
// ctx: the context to use for the liveview session, including cancellation
//
// account: AccountDetails struct containing the necessary information to start a liveview session
//
// Example: Livestream("u011", "example_token", "camera", 1234, 5678, 9012) -> nil
func Livestream(ctx context.Context, account AccountDetails) error {
	baseUrl := GetApiUrl(account.Region)
	liveViewPath, err := GetLiveviewPath(account.DeviceType)
	if err != nil {
		log.Println("Error getting liveview path", err)
		return err
	}

	// Tell Blink we want to start a liveview session
	resp, err := BeginLiveview(fmt.Sprintf(liveViewPath, baseUrl, account.AccountId, account.NetworkId, account.CameraId), account.Token)
	if err != nil {
		log.Println("Error starting liveview session", err)
		return err
	} else if resp == nil || resp.CommandId == 0 {
		log.Println("Error sending liveview command", resp)
		return fmt.Errorf("error sending liveview command")
	}

	// Poll the liveview command to keep the connection alive
	go PollCommand(ctx, fmt.Sprintf("%s/network/%d/command/%d", baseUrl, account.NetworkId, resp.CommandId), account.Token, resp.PollingInterval)
	defer StopCommand(fmt.Sprintf("%s/network/%d/command/%d/done", baseUrl, account.NetworkId, resp.CommandId), account.Token)

	// Get the connection details
	connectionDetails, err := ParseConnectionString(resp.Server)
	if err != nil {
		log.Println("Error parsing connection string", err)
		return err
	}

	// Connect to the liveview server
	if err := TCPStream(ctx, *connectionDetails); err != nil {
		log.Println("TCPStream error", err)
		return err
	}

	<-ctx.Done()
	return nil
}
