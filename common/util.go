package common

import (
	"fmt"
	"strings"
)

// GetApiUrl builds the Blink API URL based on the region if provided
// region: region to build the URL for
//
// Example: GetApiUrl("u011") = "https://rest-u011.immedia-semi.com"
//
// Example: GetApiUrl("") = "https://rest-prod.immedia-semi.com"
func GetApiUrl(region string) string {
	if region == "" {
		region = "prod"
	}

	return fmt.Sprintf("https://rest-%s.immedia-semi.com", region)
}

// GetLiveviewPath returns the liveview path based on the device type
//
// deviceType: the type of device to get the liveview path for
//
// Example: GetLiveviewPath("camera") = "%s/api/v5/accounts/%d/networks/%d/cameras/%d/liveview"
func GetLiveviewPath(deviceType string) (string, error) {
	switch deviceType {
	case "camera":
		return "%s/api/v5/accounts/%d/networks/%d/cameras/%d/liveview", nil
	case "owl", "hawk":
		return "%s/api/v2/accounts/%d/networks/%d/owls/%d/liveview", nil
	case "doorbell", "lotus":
		return "%s/api/v2/accounts/%d/networks/%d/doorbells/%d/liveview", nil
	}

	return "", fmt.Errorf("cannot build path for unknown device type: %s", deviceType)
}

type NetworkGroup struct {
	Name    string
	Devices []BaseCameraDevice
}

type DeviceOption struct {
	Option        int
	FormattedName string
	NetworkId     int
	DeviceId      int
	DeviceType    string
}

// PrintDeviceOptions prints the device options and returns a list of DeviceOption structs
//
// resp: the HomescreenResponse containing the device information
//
// Example: PrintDeviceOptions(&HomescreenResponse{Networks: []Network{...}, Owls: []Owl{...}, Doorbells: []Doorbell{...}})
func PrintDeviceOptions(resp *HomescreenResponse) (string, []DeviceOption) {
	if len(resp.Networks) == 0 {
		return "", nil
	}
	if len(resp.Owls) == 0 && len(resp.Doorbells) == 0 {
		return "", nil
	}

	var networkGroups []NetworkGroup
	for _, network := range resp.Networks {
		var devices []BaseCameraDevice
		for _, device := range resp.Doorbells {
			if device.NetworkId == network.Id {
				devices = append(devices, device)
			}
		}
		for _, device := range resp.Owls {
			if device.NetworkId == network.Id {
				devices = append(devices, device)
			}
		}

		networkGroups = append(networkGroups, NetworkGroup{
			Name:    network.Name,
			Devices: devices,
		})
	}

	var sb strings.Builder
	var options []DeviceOption
	var idx int = 1
	for _, group := range networkGroups {
		sb.WriteString(fmt.Sprintf("Network: %s\n", group.Name))
		for _, device := range group.Devices {
			formattedName := fmt.Sprintf("%s (%s)", device.Name, device.Type)
			sb.WriteString(fmt.Sprintf("  [%02d] %s\n", idx, formattedName))
			options = append(options, DeviceOption{
				Option:        idx,
				FormattedName: formattedName,
				NetworkId:     device.NetworkId,
				DeviceId:      device.Id,
				DeviceType:    device.Type,
			})
			idx++
		}
	}

	return sb.String(), options
}
