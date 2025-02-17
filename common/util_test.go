package common_test

import (
	"blink-liveview-websocket/common"
	"fmt"
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

func TestPrintDeviceOptionsSingleNetwork(t *testing.T) {
	resp := common.HomescreenResponse{
		Networks: []common.BaseNetwork{
			{
				Id:   1,
				Name: "Network 1",
			},
		},
		Owls: []common.BaseCameraDevice{
			{
				Id:        1,
				Name:      "Owl 1",
				Type:      "owl",
				NetworkId: 1,
			},
		},
		Doorbells: []common.BaseCameraDevice{
			{
				Id:        2,
				Name:      "Doorbell 1",
				Type:      "doorbell",
				NetworkId: 1,
			},
		},
	}

	output, options := common.PrintDeviceOptions(&resp)

	assert.Equal(t, output, "Network: Network 1\n  [01] Doorbell 1 (doorbell)\n  [02] Owl 1 (owl)\n")
	assert.Equal(t, len(options), 2)
	assert.Equal(t, options, []common.DeviceOption{
		{
			Option:        1,
			FormattedName: "Doorbell 1 (doorbell)",
			NetworkId:     1,
			DeviceId:      2,
			DeviceType:    "doorbell",
		},
		{
			Option:        2,
			FormattedName: "Owl 1 (owl)",
			NetworkId:     1,
			DeviceId:      1,
			DeviceType:    "owl",
		},
	})
}

func TestPrintDeviceOptionsMultiNetwork(t *testing.T) {
	resp := common.HomescreenResponse{
		Networks: []common.BaseNetwork{
			{
				Id:   1,
				Name: "CUSTOM-NETWORK-01",
			},
			{
				Id:   2,
				Name: "ANOTHER-NETWORK-02",
			},
		},
		Owls: []common.BaseCameraDevice{
			{
				Id:        1,
				Name:      "MY BLINK MINI",
				Type:      "owl",
				NetworkId: 1,
			},
		},
		Doorbells: []common.BaseCameraDevice{
			{
				Id:        2,
				Name:      "MY BLINK DOORBELL",
				Type:      "lotus",
				NetworkId: 2,
			},
		},
	}

	output, options := common.PrintDeviceOptions(&resp)

	assert.Equal(t, output, "Network: CUSTOM-NETWORK-01\n  [01] MY BLINK MINI (owl)\nNetwork: ANOTHER-NETWORK-02\n  [02] MY BLINK DOORBELL (lotus)\n")
	assert.Equal(t, len(options), 2)
	assert.Equal(t, options, []common.DeviceOption{
		{
			Option:        1,
			FormattedName: "MY BLINK MINI (owl)",
			NetworkId:     1,
			DeviceId:      1,
			DeviceType:    "owl",
		},
		{
			Option:        2,
			FormattedName: "MY BLINK DOORBELL (lotus)",
			NetworkId:     2,
			DeviceId:      2,
			DeviceType:    "lotus",
		},
	})
}

func TestPrintDeviceOptionsNoOwls(t *testing.T) {
	resp := common.HomescreenResponse{
		Networks: []common.BaseNetwork{
			{
				Id:   1,
				Name: "CUSTOM-NETWORK-01",
			},
		},
		Owls: nil,
		Doorbells: []common.BaseCameraDevice{
			{
				Id:        2,
				Name:      "Front Door",
				Type:      "lotus",
				NetworkId: 1,
			},
		},
	}

	output, options := common.PrintDeviceOptions(&resp)

	assert.Equal(t, output, "Network: CUSTOM-NETWORK-01\n  [01] Front Door (lotus)\n")
	assert.Equal(t, len(options), 1)
	assert.Equal(t, options, []common.DeviceOption{
		{
			Option:        1,
			FormattedName: "Front Door (lotus)",
			NetworkId:     1,
			DeviceId:      2,
			DeviceType:    "lotus",
		},
	})
}

func TestPrintDeviceOptionsNoDoorbells(t *testing.T) {
	resp := common.HomescreenResponse{
		Networks: []common.BaseNetwork{
			{
				Id:   950,
				Name: "A Very Custom Network Name",
			},
		},
		Owls: []common.BaseCameraDevice{
			{
				Id:        983713,
				Name:      "A Floodlight Cam",
				Type:      "superior",
				NetworkId: 950,
			},
		},
		Doorbells: nil,
	}

	output, options := common.PrintDeviceOptions(&resp)

	assert.Equal(t, output, "Network: A Very Custom Network Name\n  [01] A Floodlight Cam (superior)\n")
	assert.Equal(t, len(options), 1)
	assert.Equal(t, options, []common.DeviceOption{
		{
			Option:        1,
			FormattedName: "A Floodlight Cam (superior)",
			NetworkId:     950,
			DeviceId:      983713,
			DeviceType:    "superior",
		},
	})
}

func TestPrintDeviceOptionsEmptyNetwork(t *testing.T) {
	resp := common.HomescreenResponse{
		Networks: nil,
		Owls: []common.BaseCameraDevice{
			{
				Id:        983713,
				Name:      "A Floodlight Cam",
				Type:      "superior",
				NetworkId: 950,
			},
		},
		Doorbells: []common.BaseCameraDevice{
			{
				Id:        118292,
				Name:      "My Front Door",
				Type:      "doorbell",
				NetworkId: 950,
			},
		},
	}

	output, options := common.PrintDeviceOptions(&resp)

	assert.Equal(t, output, "")
	assert.Equal(t, len(options), 0)
	assert.Equal(t, options, nil)
}

func TestPrintDeviceOptionsEmptyDevices(t *testing.T) {
	resp := common.HomescreenResponse{
		Networks: []common.BaseNetwork{
			{
				Id:   950,
				Name: "A Very Custom Network Name",
			},
		},
		Owls:      nil,
		Doorbells: nil,
	}

	output, options := common.PrintDeviceOptions(&resp)

	assert.Equal(t, output, "")
	assert.Equal(t, len(options), 0)
	assert.Equal(t, options, nil)
}

func TestPrintDeviceOptionsMissingNetwork(t *testing.T) {
	resp := common.HomescreenResponse{
		Networks: []common.BaseNetwork{
			{
				Id:   1,
				Name: "Real Network",
			},
		},
		Owls: []common.BaseCameraDevice{
			{
				Id:        1,
				Name:      "Owl 1",
				Type:      "owl",
				NetworkId: 1,
			},
			{
				Id:        3,
				Name:      "Floodlight Cam",
				Type:      "superior",
				NetworkId: 999, // Does not exist
			},
		},
		Doorbells: []common.BaseCameraDevice{
			{
				Id:        2,
				Name:      "Doorbell 1",
				Type:      "doorbell",
				NetworkId: 999, // Does not exist
			},
		},
	}

	output, options := common.PrintDeviceOptions(&resp)

	assert.Equal(t, output, "Network: Real Network\n  [01] Owl 1 (owl)\n")
	assert.Equal(t, len(options), 1)
	assert.Equal(t, options, []common.DeviceOption{
		{
			Option:        1,
			FormattedName: "Owl 1 (owl)",
			NetworkId:     1,
			DeviceId:      1,
			DeviceType:    "owl",
		},
	})
}
