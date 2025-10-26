package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SetRequestHeaders appends the required headers to the request
//
// req: the request to append headers to
//
// token: the token to use for the request
//
// Example: SetRequestHeaders(req, "bearer-token-here")
func SetRequestHeaders(req *http.Request, token string) {
	req.Header.Set("locale", "en_US")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("content-type", "application/json; charset=UTF-8")
}

type CommandResponse struct {
	Code       int    `json:"code"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Complete   bool   `json:"complete"`
}

// PollCommand will repeatedly poll the command URL with the provided token
//
// ctx: the context to use for the command
//
// url: the URL to poll
//
// token: the token to use for the request
//
// pollInterval: the interval to wait between polls in seconds
//
// Example: go PollCommand("https://example.com", "api-token-here", 10)
func PollCommand(ctx context.Context, url string, token string, pollInterval int) error {
	ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return err
			}

			SetRequestHeaders(req, token)

			client := &http.Client{Timeout: time.Second * 10}
			resp, err := client.Do(req)
			if resp.StatusCode != http.StatusOK || err != nil {
				return fmt.Errorf("error polling command. HTTP Status Code %d", resp.StatusCode)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			result := CommandResponse{}
			if err != nil {
				return err
			}

			err = json.Unmarshal(body, &result)
			if err != nil {
				return err
			}

			if result.Complete {
				return fmt.Errorf("command marked as complete. Cannot poll further")
			}
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
//
// url: the URL to send the liveview request to
//
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
		return nil, fmt.Errorf("error starting liveview. HTTP Status Code %d", resp.StatusCode)
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

// StopCommand marks the command (liveview) as completed
//
// url: the URL to send the liveview request to
//
// token: the token to use for the request
//
// Example: StopCommand("https://example.com", "api-token-here")
func StopCommand(url string, token string) error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	SetRequestHeaders(req, token)

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK || err != nil {
		return fmt.Errorf("cannot stop command. HTTP Status Code %d", resp.StatusCode)
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
		return fmt.Errorf("cannot stop command. API Code %d with message %s", result.Code, result.Message)
	}

	return nil
}

type LoginBody struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	ClientId   string `json:"client_id"`
	Scope      string `json:"scope"`
	GrantType  string `json:"grant_type"`
	ClientName string `json:"client_name"`
}

type LoginResponse struct {
	// First Authentication Response
	NextTimeInSeconds   int    `json:"next_time_in_seconds"`
	Phone               string `json:"phone"`
	TwoStepVerification string `json:"tsv_state"`

	// Second Authentication Response
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// Login logs in to the Blink API using the provided credentials
//
// email: the email address to use for login
//
// password: the password to use for login
//
// code: the 2FA code to use for login (if applicable)
//
// fp: the fingerprint to use for login
//
// Example: Login("x", "y", "123456", fingerprint)
func Login(email string, password string, code string, fp *Fingerprint) (*LoginResponse, error) {
	jsonBody, _ := json.Marshal(&LoginBody{
		Username:   email,
		Password:   password,
		ClientId:   "android",
		Scope:      "client",
		GrantType:  "password",
		ClientName: "blink-liveview-middleware",
	})

	req, err := http.NewRequest("POST", "https://api.oauth.blink.com/oauth/token", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json; charset=UTF-8")
	req.Header.Set("hardware_id", fp.Value)
	if code != "" {
		req.Header.Set("2fa-code", code)
	}

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result LoginResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type TierInfoResponse struct {
	Tier      string `json:"tier"`
	AccountId int    `json:"account_id"`
}

func GetTierInfo(token string) (*TierInfoResponse, error) {
	req, err := http.NewRequest("GET", GetApiUrl("")+"/api/v1/users/tier_info", nil)
	if err != nil {
		return nil, err
	}

	SetRequestHeaders(req, token)

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK || err != nil {
		return nil, fmt.Errorf("HTTP Status Code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TierInfoResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type BaseCameraDevice struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	NetworkId int    `json:"network_id"`
}

type BaseNetwork struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type HomescreenResponse struct {
	Networks  []BaseNetwork      `json:"networks"`
	Owls      []BaseCameraDevice `json:"owls"`
	Doorbells []BaseCameraDevice `json:"doorbells"`
}

// Homescreen retrieves the homescreen information from the Blink API
//
// url: the URL to send the homescreen request to
//
// token: the API token to use for the request
//
// Example: Homescreen("https://example.com", "api-token-here")
func Homescreen(url string, token string) (*HomescreenResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	SetRequestHeaders(req, token)

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK || err != nil {
		return nil, fmt.Errorf("HTTP Status Code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result HomescreenResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
