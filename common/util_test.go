package common_test

import (
	"blink-liveview-websocket/common"
	"fmt"
	"os"
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

func TestParseConnectionStringNominal1(t *testing.T) {
	server := "immis://3.932.71.73:443/Az3eugol5Zsg6J5x__IMDS_A1B2C3D4E5F6G7H8?client_id=918202"
	conn, err := common.ParseConnectionString(server)

	assert.Equal(t, err, nil)
	assert.Equal(t, conn.Host, "3.932.71.73")
	assert.Equal(t, conn.Port, "443")
	assert.Equal(t, conn.ClientId, 918202)
	assert.Equal(t, conn.ConnectionId, "Az3eugol5Zsg6J5x")
}

func TestParseConnectionStringNominal2(t *testing.T) {
	server := "rtsp://3.233.10.25:443/Cy5gwipn7Bui8L7z__IMDS_B9C8D7E6F5G4H3I2?client_id=255"
	conn, err := common.ParseConnectionString(server)

	assert.Equal(t, err, nil)
	assert.Equal(t, conn.Host, "3.233.10.25")
	assert.Equal(t, conn.Port, "443")
	assert.Equal(t, conn.ClientId, 255)
	assert.Equal(t, conn.ConnectionId, "Cy5gwipn7Bui8L7z")
}

func TestParseConnectionStringNominal3(t *testing.T) {
	server := "rtsp://3.233.10.25:443/Kz3oepxv5Jcq6T5h_SERIAL?client_id=75555"
	conn, err := common.ParseConnectionString(server)

	assert.Equal(t, err, nil)
	assert.Equal(t, conn.Host, "3.233.10.25")
	assert.Equal(t, conn.Port, "443")
	assert.Equal(t, conn.ClientId, 75555)
	assert.Equal(t, conn.ConnectionId, "Kz3oepxv5Jcq6T5h")
}

func TestParseConnectionStringParseError(t *testing.T) {
	conn, err := common.ParseConnectionString("not a url")

	assert.Equal(t, conn, nil)
	assert.NotEqual(t, err, nil)
}

func TestParseConnectionStringInvalidHostname(t *testing.T) {
	server := "rtsp:///Cy5gwipn7Bui8L7z__IMDS_B9C8D7E6F5G4H3I2?client_id=255"
	conn, err := common.ParseConnectionString(server)

	assert.Equal(t, conn, nil)
	assert.Equal(t, err, fmt.Errorf("invalid host"))
}

func TestParseConnectionStringInvalidPort(t *testing.T) {
	server := "immis://3.233.10.25:80/Cy5gwipn7Bui8L7z__IMDS_B9C8D7E6F5G4H3I2?client_id=255"
	conn, err := common.ParseConnectionString(server)

	assert.Equal(t, conn, nil)
	assert.Equal(t, err, fmt.Errorf("unexpected port 80. Expecting 443"))
}

func TestParseConnectionStringInvalidConnectionID(t *testing.T) {
	server1 := "immis://3.233.10.25:443?client_id=255"
	conn1, err1 := common.ParseConnectionString(server1)

	assert.Equal(t, conn1, nil)
	assert.Equal(t, err1, fmt.Errorf("invalid connection ID"))

	server2 := "immis://3.233.10.25:443/XXXXXXXXX/?client_id=255"
	conn2, err2 := common.ParseConnectionString(server2)

	assert.Equal(t, conn2, nil)
	assert.Equal(t, err2, fmt.Errorf("invalid connection ID"))
}

func TestParseConnectionStringInvalidClientID(t *testing.T) {
	server1 := "immis://3.233.10.25:443/Cy5gwipn7Bui8L7z__IMDS_B9C8D7E6F5G4H3I2?client_id="
	conn1, err1 := common.ParseConnectionString(server1)

	assert.Equal(t, conn1, nil)
	assert.Equal(t, err1, fmt.Errorf("invalid client ID"))
}

func TestGetTCPAuthFramesNominal(t *testing.T) {
	frames := common.GetTCPAuthFrames("connection-id", 123)

	assert.Equal(t, len(frames), 5)
}

func TestGetTCPAuthFrame1(t *testing.T) {
	frames := common.GetTCPAuthFrames("", 0)

	assert.Equal(t, frames[0], []byte{
		0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})
}

func TestGetTCPAuthFrame2(t *testing.T) {
	frames := common.GetTCPAuthFrames("", 123)
	assert.Equal(t, frames[1], []byte{
		0x00, 0x00, 0x00, 0x7b,
	})

	frames2 := common.GetTCPAuthFrames("", 456)
	assert.Equal(t, frames2[1], []byte{
		0x00, 0x00, 0x01, 0xc8,
	})

	frames3 := common.GetTCPAuthFrames("", 75890)
	assert.Equal(t, frames3[1], []byte{
		0x00, 0x01, 0x28, 0x72,
	})
}

func TestGetTCPAuthFrame3(t *testing.T) {
	frames := common.GetTCPAuthFrames("", 0)

	assert.Equal(t, frames[2], []byte{
		0x01, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x10,
	})
}

func TestGetTCPAuthFrame4(t *testing.T) {
	frames := common.GetTCPAuthFrames("Cy5gwipn7Bui8L7z", 0)
	assert.Equal(t, frames[3], []byte{
		0x43, 0x79, 0x35, 0x67, 0x77, 0x69, 0x70, 0x6e,
		0x37, 0x42, 0x75, 0x69, 0x38, 0x4c, 0x37, 0x7a,
	})

	frames2 := common.GetTCPAuthFrames("Fz8jzlsq0Exl1O0c", 0)
	assert.Equal(t, frames2[3], []byte{
		0x46, 0x7a, 0x38, 0x6a, 0x7a, 0x6c, 0x73, 0x71,
		0x30, 0x45, 0x78, 0x6c, 0x31, 0x4f, 0x30, 0x63,
	})

	frames3 := common.GetTCPAuthFrames("ABC", 0)
	assert.Equal(t, frames3[3], []byte{
		0x41, 0x42, 0x43,
	})
}

func TestGetTCPAuthFrame5(t *testing.T) {
	frames := common.GetTCPAuthFrames("", 0)

	assert.Equal(t, frames[4], []byte{
		0x00, 0x00, 0x00, 0x01, 0x0a, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
	})
}

func TestGetFingerprintNominal(t *testing.T) {
	mockdir := t.TempDir()
	mockFile := mockdir + "/fingerprint.txt"
	os.WriteFile(mockFile, []byte("this-is-a-fake-uuid"), 0644)

	generated, fingerprint, err := common.GetFingerprint(mockFile)

	assert.Equal(t, generated, false)
	assert.Equal(t, fingerprint, "this-is-a-fake-uuid")
	assert.Equal(t, err, nil)
}

func TestGetFingerprintFileDoesNotExist(t *testing.T) {
	mockdir := t.TempDir()
	mockFile := mockdir + "/fingerprint.txt"
	generated, fingerprint, err := common.GetFingerprint(mockFile)

	assert.Equal(t, generated, true)
	assert.NotEqual(t, fingerprint, "") // should be a UUID
	assert.Equal(t, err, nil)
}

func TestGetFingerprintEmptyFile(t *testing.T) {
	mockdir := t.TempDir()
	mockFile := mockdir + "/fingerprint.txt"
	os.WriteFile(mockFile, []byte(""), 0644)

	generated, fingerprint, err := common.GetFingerprint(mockFile)

	assert.Equal(t, generated, true)
	assert.NotEqual(t, fingerprint, "") // should be a UUID
	assert.Equal(t, err, nil)
}

func TestGenerateFingerprintNominal(t *testing.T) {
	mockdir := t.TempDir()
	mockFile := mockdir + "/fingerprint-normal.txt"
	fingerprint, err := common.GenerateFingerprint(mockFile)

	file, _ := os.ReadFile(mockFile)

	assert.NotEqual(t, fingerprint, "")
	assert.Equal(t, fingerprint, string(file)) // Ensure the file contains the fingerprint returned
	assert.Equal(t, err, nil)
}
