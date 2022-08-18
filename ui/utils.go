package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
)

// SetBluezConn sets up the bluez connection.
func SetBluezConn(b *bluez.Bluez) {
	BluezConn = b
}

// SetObexConn sets up the bluez obex connection.
func SetObexConn(o *bluez.Obex) {
	ObexConn = o
}

// SetTrusted sets the trusted state of a device.
func SetTrusted(devicePath string, enable bool) error {
	BluezConn.SetDeviceProperty(devicePath, "Trusted", true)

	return nil
}

// GetDeviceFromPath gets a device from the device path.
func GetDeviceFromPath(devicePath string) (bluez.Device, error) {
	device := BluezConn.GetDevice(devicePath)
	if device.Path == "" {
		return bluez.Device{}, errors.New("Device not found")
	}

	return device, nil
}

// formatSize returns the human readable form of a size value in bytes.
// Adapted from: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func formatSize(size int64) string {
	const unit = 1000
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "kMGTPE"[exp])
}

// savefile moves a file from the obex cache to a specified user-accessible directory.
// If the directory is not specified, it automatically creates a directory in the
// user's home path and moves the file there.
func savefile(path string) error {
	userpath := cmd.GetConfigProperty("receive-dir")
	if userpath == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		userpath = filepath.Join(homedir, "bluetuith")

		if _, err := os.Stat(userpath); err != nil {
			err = os.Mkdir(userpath, 0700)
			if err != nil {
				return err
			}
		}
	}

	return os.Rename(path, filepath.Join(userpath, filepath.Base(path)))
}
