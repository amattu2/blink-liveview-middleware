package common_test

import (
	"blink-liveview-websocket/common"
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestTCPStream(t *testing.T) {
	t.Skip("Not implemented")
}

func TestSendPing(t *testing.T) {
	t.Skip("Not implemented")
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

func TestGenerateAuthFramesNominal(t *testing.T) {
	frames := common.TCPConnectionDetails{ConnectionId: "connection-id", ClientId: 123}.GenerateAuthFrames()

	assert.Equal(t, len(frames), 5)
}

func TestGenerateAuthFrame1(t *testing.T) {
	frames := common.TCPConnectionDetails{ConnectionId: "", ClientId: 0}.GenerateAuthFrames()

	assert.Equal(t, frames[0], []byte{
		0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})
}

func TestGenerateAuthFrame2(t *testing.T) {
	frames := common.TCPConnectionDetails{ConnectionId: "", ClientId: 123}.GenerateAuthFrames()
	assert.Equal(t, frames[1], []byte{
		0x00, 0x00, 0x00, 0x7b,
	})

	frames2 := common.TCPConnectionDetails{ConnectionId: "", ClientId: 456}.GenerateAuthFrames()
	assert.Equal(t, frames2[1], []byte{
		0x00, 0x00, 0x01, 0xc8,
	})

	frames3 := common.TCPConnectionDetails{ConnectionId: "", ClientId: 75890}.GenerateAuthFrames()
	assert.Equal(t, frames3[1], []byte{
		0x00, 0x01, 0x28, 0x72,
	})
}

func TestGenerateAuthFrame3(t *testing.T) {
	frames := common.TCPConnectionDetails{ConnectionId: "", ClientId: 0}.GenerateAuthFrames()

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

func TestGenerateAuthFrame4(t *testing.T) {
	frames := common.TCPConnectionDetails{ConnectionId: "Cy5gwipn7Bui8L7z", ClientId: 0}.GenerateAuthFrames()
	assert.Equal(t, frames[3], []byte{
		0x43, 0x79, 0x35, 0x67, 0x77, 0x69, 0x70, 0x6e,
		0x37, 0x42, 0x75, 0x69, 0x38, 0x4c, 0x37, 0x7a,
	})

	frames2 := common.TCPConnectionDetails{ConnectionId: "Fz8jzlsq0Exl1O0c", ClientId: 0}.GenerateAuthFrames()
	assert.Equal(t, frames2[3], []byte{
		0x46, 0x7a, 0x38, 0x6a, 0x7a, 0x6c, 0x73, 0x71,
		0x30, 0x45, 0x78, 0x6c, 0x31, 0x4f, 0x30, 0x63,
	})

	frames3 := common.TCPConnectionDetails{ConnectionId: "ABC", ClientId: 0}.GenerateAuthFrames()
	assert.Equal(t, frames3[3], []byte{
		0x41, 0x42, 0x43,
	})
}

func TestGenerateAuthFrame5(t *testing.T) {
	frames := common.TCPConnectionDetails{ConnectionId: "", ClientId: 0}.GenerateAuthFrames()

	assert.Equal(t, frames[4], []byte{
		0x00, 0x00, 0x00, 0x01, 0x0a, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00,
	})
}
