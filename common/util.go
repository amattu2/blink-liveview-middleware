package common

import (
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// GetApiUrl builds the Blink API URL based on the region if provided
// region: region to build the URL for
//
// Example: GetApiUrl("u011") = "https://rest-u011.immedia-semi.com"
//
// Example: GetApiUrl("") = "https://rest-prod.immedia-semi.com"
func GetApiUrl(region string) string {
	if region == "" {
		region = "prod"
	}

	return fmt.Sprintf("https://rest-%s.immedia-semi.com", region)
}

// GetLiveviewPath returns the liveview path based on the device type
//
// deviceType: the type of device to get the liveview path for
//
// Example: GetLiveviewPath("camera") = "%s/api/v5/accounts/%d/networks/%d/cameras/%d/liveview"
func GetLiveviewPath(deviceType string) (string, error) {
	switch deviceType {
	case "camera":
		return "%s/api/v5/accounts/%d/networks/%d/cameras/%d/liveview", nil
	case "owl":
		return "%s/api/v2/accounts/%d/networks/%d/owls/%d/liveview", nil
	case "doorbell":
	case "lotus":
		return "%s/api/v2/accounts/%d/networks/%d/doorbells/%d/liveview", nil
	}

	return "", fmt.Errorf("cannot build path for unknown device type: %s", deviceType)
}

// ParseConnectionString parses the connection string to extract the connection details
//
// url: the connection string to parse
//
// Example: ParseConnectionString("TODO")
func ParseConnectionString(server string) (*ConnectionDetails, error) {
	parsedUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	if parsedUrl.Hostname() == "" {
		return nil, fmt.Errorf("invalid host")
	}

	if parsedUrl.Port() != "443" {
		return nil, fmt.Errorf("unexpected port %s. Expecting 443", parsedUrl.Port())
	}

	connID := strings.Split(strings.Split(server, "/")[len(strings.Split(server, "/"))-1], "_")
	if len(connID) <= 1 || connID[0] == "" {
		return nil, fmt.Errorf("invalid connection ID")
	}

	clientID, err := strconv.Atoi(parsedUrl.Query().Get("client_id"))
	if clientID == 0 || err != nil {
		return nil, fmt.Errorf("invalid client ID")
	}

	return &ConnectionDetails{
		Host:         parsedUrl.Hostname(),
		Port:         parsedUrl.Port(),
		ClientId:     clientID,
		ConnectionId: connID[0],
	}, nil
}

// GetTCPAuthFrames returns the header payload for the TCP connection
//
// connectionId: the connection ID to use in the header
//
// clientId: the client ID to use in the header
//
// Example: GetTCPAuthFrames("connection-id", 123)
func GetTCPAuthFrames(connectionId string, clientId int) [][]byte {
	// Frame 1 (unknown)
	frame1 := []byte{
		0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	// Frame 2 (Client ID)
	clientIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(clientIDBytes, uint32(clientId))
	frame2 := []byte{
		clientIDBytes[0], clientIDBytes[1], clientIDBytes[2], clientIDBytes[3],
	}

	// Frame 3 (unknown)
	frame3 := []byte{
		0x01, 0x08, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x10,
	}

	// Frame 4 (Connection ID)
	frame4 := []byte(connectionId)

	// Frame 5 (unknown)
	frame5 := []byte{
		0x00, 0x00, 0x00, 0x01, 0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	return [][]byte{
		frame1,
		frame2,
		frame3,
		frame4,
		frame5,
	}
}
