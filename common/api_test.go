package common_test

import (
	"blink-liveview-websocket/common"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func TestSetRequestHeaders(t *testing.T) {
	// Create a new request
	req, _ := http.NewRequest("GET", "https://example.com", nil)

	// Set the request headers
	common.SetRequestHeaders(req, "xyz-auth-token")

	// Check the headers
	assert.Equal(t, "en_US", req.Header.Get("locale"))
	assert.Equal(t, "Bearer xyz-auth-token", req.Header.Get("Authorization"))
	assert.Equal(t, "application/json; charset=UTF-8", req.Header.Get("content-type"))
}

func TestPollCommandInterval(t *testing.T) {
	var called bool
	var mu sync.Mutex

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		mu.Lock()
		called = true
		mu.Unlock()
	}))
	defer mockServer.Close()

	mockCtx, mockCancel := context.WithCancel(context.Background())
	defer mockCancel()

	go common.PollCommand(mockCtx, mockServer.URL, "xyz-auth-token", 1)

	// Sleep for 2 seconds to allow the poll to occur
	<-time.After(2 * time.Second)

	mu.Lock()
	assert.Equal(t, true, called)
	mu.Unlock()
}

func TestPollCommandCancel(t *testing.T) {
	var called bool
	var mu sync.Mutex

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		mu.Lock()
		called = true
		mu.Unlock()
	}))
	defer mockServer.Close()

	mockCtx, mockCancel := context.WithCancel(context.Background())

	go common.PollCommand(mockCtx, mockServer.URL, "xyz-auth-token", 1)

	time.Sleep(2 * time.Second)

	mockCancel()

	mu.Lock()
	assert.Equal(t, true, called)
	mu.Unlock()
}

func TestPollCommandHttpError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	mockCtx, mockCancel := context.WithCancel(context.Background())
	defer mockCancel()

	err := common.PollCommand(mockCtx, mockServer.URL, "xyz-auth-token", 1)

	assert.Equal(t, "error polling command. HTTP Status Code 500", err.Error())
}

func TestPollCommandComplete(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": 200, "status_code": 200, "message": "OK", "complete": true}`))
	}))
	defer mockServer.Close()

	mockCtx, mockCancel := context.WithCancel(context.Background())
	defer mockCancel()

	err := common.PollCommand(mockCtx, mockServer.URL, "xyz-auth-token", 1)

	assert.Equal(t, "command marked as complete. Cannot poll further", err.Error())
}

func TestBeginLiveviewNominal(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"command_id": 75888, "polling_interval": 15, "server": "immis://93.93.93.93:443"}`))
	}))
	defer mockServer.Close()

	resp, err := common.BeginLiveview(mockServer.URL, "xyz-auth-token")

	assert.Equal(t, nil, err)
	assert.Equal(t, 75888, resp.CommandId)
	assert.Equal(t, 15, resp.PollingInterval)
	assert.Equal(t, "immis://93.93.93.93:443", resp.Server)
}

func TestBeginLiveviewHttpError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	resp, err := common.BeginLiveview(mockServer.URL, "xyz-auth-token")

	assert.Equal(t, nil, resp)
	assert.Equal(t, "error starting liveview. HTTP Status Code 500", err.Error())
}

func TestStopCommandNominal(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": 902, "status_code": 200, "message": "OK"}`))
	}))
	defer mockServer.Close()

	err := common.StopCommand(mockServer.URL, "xyz-auth-token")

	assert.Equal(t, nil, err)
}

func TestStopCommandAPIError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code": 800, "status_code": 400, "message": "Some error"}`))
	}))
	defer mockServer.Close()

	err := common.StopCommand(mockServer.URL, "xyz-auth-token")

	assert.Equal(t, "cannot stop command. API Code 800 with message Some error", err.Error())
}

func TestStopCommandHttpError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	err := common.StopCommand(mockServer.URL, "xyz-auth-token")

	assert.Equal(t, "cannot stop command. HTTP Status Code 500", err.Error())
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestLoginNominal(t *testing.T) {
	orig := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = orig })

	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// Ensure the URL used by Login matches the expected OAuth endpoint
		assert.Equal(t, "https://api.oauth.blink.com/oauth/token", r.URL.String())

		// Return a successful OAuth token response
		body := `{"access_token":"xyz-auth-token","expires_in":3600,"refresh_token":"r1","scope":"client","token_type":"Bearer"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioNopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})

	fp := common.Fingerprint{Value: "mock-fingerprint", New: false}
	resp, err := common.Login("mock-email", "mock-password", "", &fp)

	assert.Equal(t, nil, err)
	assert.Equal(t, "xyz-auth-token", resp.AccessToken)
	assert.Equal(t, 3600, resp.ExpiresIn)
	assert.Equal(t, "r1", resp.RefreshToken)
	assert.Equal(t, "client", resp.Scope)
	assert.Equal(t, "Bearer", resp.TokenType)
}

func TestLoginSetsHeaders(t *testing.T) {
	orig := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = orig })

	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// Validate required headers set by Login
		assert.Equal(t, "application/json; charset=UTF-8", r.Header.Get("content-type"))
		assert.Equal(t, "mock-fingerprint", r.Header.Get("hardware_id"))
		assert.Equal(t, "", r.Header.Get("2fa-code"))

		body := `{"access_token":"ok","expires_in":1,"refresh_token":"r","scope":"client","token_type":"Bearer"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioNopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})

	fp := common.Fingerprint{Value: "mock-fingerprint", New: false}
	_, err := common.Login("email", "password", "", &fp)
	assert.Equal(t, nil, err)
}

func TestLoginWithTwoFactorHeader(t *testing.T) {
	orig := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = orig })

	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		assert.Equal(t, "123456", r.Header.Get("2fa-code"))

		body := `{"access_token":"ok","expires_in":1,"refresh_token":"r","scope":"client","token_type":"Bearer"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioNopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})

	fp := common.Fingerprint{Value: "mock-fingerprint", New: false}
	_, err := common.Login("email", "password", "123456", &fp)
	assert.Equal(t, nil, err)
}

func TestLoginHttpClientError(t *testing.T) {
	orig := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = orig })

	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("dial error")
	})

	fp := common.Fingerprint{Value: "mock-fingerprint", New: false}
	resp, err := common.Login("mock-email", "mock-password", "", &fp)

	assert.Equal(t, (*common.LoginResponse)(nil), resp)
	assert.Equal(t, true, strings.HasPrefix(err.Error(), "HTTP request failed:"))
}

// Helper to avoid importing io for NopCloser in each test
type nopCloser struct{ *bytes.Buffer }

func (n nopCloser) Close() error { return nil }

func ioNopCloser(b *bytes.Buffer) nopCloser { return nopCloser{b} }

// func TestLoginHttpError(t *testing.T) {
// 	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusInternalServerError)
// 	}))
// 	defer mockServer.Close()

// 	fp := common.Fingerprint{
// 		Value: "mock-fingerprint",
// 		New:   false,
// 	}
// 	resp, err := common.Login("mock-email", "mock-password", "", &fp)

// 	assert.Equal(t, nil, resp)
// 	assert.Equal(t, "HTTP Status Code 500", err.Error())
// }

func TestHomescreenNominal(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"networks": [], "owls": [], "doorbells": []}`))
	}))
	defer mockServer.Close()

	resp, err := common.Homescreen(mockServer.URL, "xyz-auth-token")

	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(resp.Networks))
	assert.Equal(t, 0, len(resp.Owls))
	assert.Equal(t, 0, len(resp.Doorbells))
}

func TestHomescreenHttpError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	resp, err := common.Homescreen(mockServer.URL, "xyz-auth-token")

	assert.Equal(t, nil, resp)
	assert.Equal(t, "HTTP Status Code 500", err.Error())
}
