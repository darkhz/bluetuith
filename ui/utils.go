package ui

import (
	"errors"

	"github.com/darkhz/bluetuith/bluez"
)

// SetBluezConn sets up the bluez connection.
func SetBluezConn(b *bluez.Bluez) {
	BluezConn = b
}

func SetTrusted(devicePath string, enable bool) error {
	BluezConn.SetDeviceProperty(devicePath, "Trusted", true)

	return nil
}

func GetDeviceFromPath(devicePath string) (bluez.Device, error) {
	device := BluezConn.GetDevice(devicePath)
	if device == (bluez.Device{}) {
		return bluez.Device{}, errors.New("Device not found")
	}

	return device, nil
}
