package common_test

import (
	"blink-liveview-websocket/common"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestGetURLWithRegion(t *testing.T) {
	r := common.GetApiUrl("u014")

	assert.Equal(t, "https://rest-u014.immedia-semi.com", r)
}

func TestGetURLWithoutRegion(t *testing.T) {
	r := common.GetApiUrl("")

	assert.Equal(t, "https://rest-prod.immedia-semi.com", r)
}

func TestGetLiveviewPathCamera(t *testing.T) {
	path, err := common.GetLiveviewPath("camera")

	assert.Equal(t, "%s/api/v5/accounts/%d/networks/%d/cameras/%d/liveview", path)
	assert.Equal(t, err, nil)
}

func TestGetLiveviewPathOwl(t *testing.T) {
	path, err := common.GetLiveviewPath("owl")

	assert.Equal(t, "%s/api/v2/accounts/%d/networks/%d/owls/%d/liveview", path)
	assert.Equal(t, err, nil)
}

func TestGetLiveviewPathDoorbell(t *testing.T) {
	path, err := common.GetLiveviewPath("doorbell")

	assert.Equal(t, "", path)
	assert.NotEqual(t, err, nil)
}

func TestGetLiveviewPathLotus(t *testing.T) {
	path, err := common.GetLiveviewPath("lotus")

	assert.Equal(t, "%s/api/v2/accounts/%d/networks/%d/doorbells/%d/liveview", path)
	assert.Equal(t, err, nil)
}

func TestGetLiveviewPathUnknown(t *testing.T) {
	path, err := common.GetLiveviewPath("unknown")

	assert.Equal(t, "", path)
	assert.NotEqual(t, err, nil)
}

func TestParseConnectionStrin(t *testing.T) {
	t.Skip("Not implemented")
}

func TestGetTCPAuthFrame(t *testing.T) {
	t.Skip("Not implemented")
}
