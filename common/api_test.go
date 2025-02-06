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
