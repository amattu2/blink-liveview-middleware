package common_test

import (
	"blink-liveview-websocket/common"
	"context"
	"net/http"
	"net/http/httptest"
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
	assert.Equal(t, "ANDROID_28799573", req.Header.Get("app-build"))
	assert.Equal(t, "37.0ANDROID_28799573", req.Header.Get("user-agent"))
	assert.Equal(t, "en_US", req.Header.Get("locale"))
	assert.Equal(t, "America/New_York", req.Header.Get("x-blink-time-zone"))
	assert.Equal(t, "xyz-auth-token", req.Header.Get("token-auth"))
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
