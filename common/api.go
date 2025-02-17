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
// Example: SetRequestHeaders(req, "api-token-here")
func SetRequestHeaders(req *http.Request, token string) {
	req.Header.Set("app-build", "ANDROID_28799573")
	req.Header.Set("user-agent", "37.0ANDROID_28799573")
	req.Header.Set("locale", "en_US")
	req.Header.Set("x-blink-time-zone", "America/New_York")
	req.Header.Set("token-auth", token)
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
	Email      string `json:"email"`
	Password   string `json:"password"`
	UniqueId   string `json:"unique_id"`
	ClientType string `json:"client_type"`
	DeviceId   string `json:"device_identifier"`
	OsVersion  string `json:"os_version"`
	ClientName string `json:"client_name"`
	Reauth     bool   `json:"reauth"`
}

type LoginResponse struct {
	Account struct {
		AccountId                  int    `json:"account_id"`
		ClientId                   int    `json:"client_id"`
		Tier                       string `json:"tier"`
		ClientVerificationRequired bool   `json:"client_verification_required"`
	} `json:"account"`
	Auth struct {
		Token string `json:"token"`
	} `json:"auth"`
}

// Login logs in to the Blink API using the provided credentials
//
// email: the email address to use for login
//
// password: the password to use for login
//
// fp: the fingerprint to use for login
//
// Example: Login("x", "y", &Fingerprint{New: true})
func Login(email string, password string, fp *Fingerprint) (*LoginResponse, error) {
	jsonBody, _ := json.Marshal(&LoginBody{
		Email:      email,
		Password:   password,
		UniqueId:   fp.String(),
		ClientType: "android",
		DeviceId:   "Google Pixel 7 Pro, BlinkLiveviewMiddleware",
		OsVersion:  "14.0",
		ClientName: "blink-liveview-middleware",
		Reauth:     !fp.New,
	})

	url := GetApiUrl("") + "/api/v5/account/login"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("content-type", "application/json; charset=UTF-8")

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

	var result LoginResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type VerifyPinBody struct {
	Pin string `json:"pin"`
}

type VerifyPinResponse struct {
}

// VerifyPin will post the provided pin to the verification URL
//
// url: the URL to send the verification request to
//
// token: the API token to use for the request
//
// pin: the pin to verify
//
// Example: VerifyPin("https://example.com", "api-token-here", "123456") // returns nil
func VerifyPin(url string, token string, pin string) error {
	jsonBody, _ := json.Marshal(&VerifyPinBody{
		Pin: pin,
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	SetRequestHeaders(req, token)

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if resp.StatusCode != http.StatusOK || err != nil {
		return fmt.Errorf("HTTP Status Code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	return nil
}
