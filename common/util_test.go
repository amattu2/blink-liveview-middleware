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

func TestGetLiveviewPathHawk(t *testing.T) {
	path, err := common.GetLiveviewPath("hawk")

	assert.Equal(t, "%s/api/v2/accounts/%d/networks/%d/owls/%d/liveview", path)
	assert.Equal(t, err, nil)
}

func TestGetLiveviewPathDoorbell(t *testing.T) {
	path, err := common.GetLiveviewPath("doorbell")

	assert.Equal(t, "%s/api/v2/accounts/%d/networks/%d/doorbells/%d/liveview", path)
	assert.Equal(t, err, nil)
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
