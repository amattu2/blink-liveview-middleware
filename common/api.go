package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// GetApiUrl builds the Blink API URL based on the region if provided
// region: region to build the URL for
//
// Example: GetApiUrl("u011") = "https://rest-u011.immedia-semi.com"
// Example: GetApiUrl("") = "https://rest-prod.immedia-semi.com"
func GetApiUrl(region string) string {
	if region == "" {
		region = "prod"
	}

	return fmt.Sprintf("https://rest-%s.immedia-semi.com", region)
}

// SetRequestHeaders appends the required headers to the request
// req: the request to append headers to
// token: the token to use for the request
//
// Example: SetRequestHeaders(req, "api-token-here")
func SetRequestHeaders(req *http.Request, token string) {
	req.Header.Set("app-build", "ANDROID_28799573")
	req.Header.Set("user-agent", "37.0ANDROID_28799573")
	req.Header.Set("locale", "en_US")
	req.Header.Set("x-blink-time-zone", "America/New_York")
	req.Header.Set("token-auth", token)
	req.Header.Set("content-type", "application/json; charset=UTF-8")
}

type ConnectionDetails struct {
	Host         string
	Port         string
	ClientId     int
	ConnectionId string
}

// ParseConnectionString parses the connection string to extract the connection details
// url: the connection string to parse
//
// Example: ParseConnectionString("TODO")
func ParseConnectionString(server string) (*ConnectionDetails, error) {
	if server == "" {
		return nil, fmt.Errorf("invalid connection URL")
	}

	parsedUrl, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	host := strings.Split(strings.Split(server, "/")[2], ":")
	if len(host) <= 1 || host[0] == "" {
		return nil, fmt.Errorf("invalid host")
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
		Host:         host[0],
		Port:         "443",
		ClientId:     clientID,
		ConnectionId: connID[0],
	}, nil
}

type CommandResponse struct {
	Code       int    `json:"code"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Complete   bool   `json:"complete"`
}

// PollCommand will repeatedly poll the command URL with the provided token
// ctx: the context to use for the command
// url: the URL to poll
// token: the token to use for the request
// pollInterval: the interval to wait between polls in seconds
//
// Example: go PollCommand("https://example.com", "api-token-here", 10)
func PollCommand(ctx context.Context, url string, token string, pollInterval int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Println(err)
				return
			}

			SetRequestHeaders(req, token)

			client := &http.Client{Timeout: time.Second * 10}
			resp, err := client.Do(req)
			if resp.StatusCode != http.StatusAccepted || err != nil {
				log.Println("Error polling API", resp.StatusCode, err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			result := CommandResponse{}
			if err != nil {
				log.Println(err)
				return
			}

			err = json.Unmarshal(body, &result)
			if err != nil {
				log.Println(err)
				return
			}

			if result.Complete {
				return
			}

			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}
}

type LiveviewInput struct {
	Intent string `json:"intent"`
}

type LiveviewResponse struct {
	CommandId       int    `json:"command_id"`
	PollingInterval int    `json:"polling_interval"`
	Server          string `json:"server"`
}

// BeginLiveview starts the liveview intention for the camera
// url: the URL to send the liveview request to
// token: the token to use for the request
//
// Example: BeginLiveview("https://example.com", "api-token-here")
func BeginLiveview(url string, token string) (*LiveviewResponse, error) {
	jsonBody, _ := json.Marshal(&LiveviewInput{
		Intent: "liveview",
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	SetRequestHeaders(req, token)

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK || err != nil {
		return nil, fmt.Errorf("error starting liveview. Status Code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result LiveviewResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// StopLiveview stops the liveview session for the camera
// url: the URL to send the liveview request to
// token: the token to use for the request
//
// Example: StopLiveview("https://example.com", "api-token-here")
func StopLiveview(url string, token string) error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	SetRequestHeaders(req, token)

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK || err != nil {
		return fmt.Errorf("unable to stop liveview. Status Code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result CommandResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	if result.Code != 902 {
		return fmt.Errorf("unable to stop liveview. Code %d with message %s", result.Code, result.Message)
	}

	return nil
}
