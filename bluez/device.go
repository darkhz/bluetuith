// Copyright 2018 Jonathan Pentecost
//
// Copyright 2020 darkhz
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package bluez

import (
	"path/filepath"

	"github.com/godbus/dbus/v5"
)

const (
	dbusBluezDeviceIface  = "org.bluez.Device1"
	dbusBluezBatteryIface = "org.bluez.Battery1"
)

// Device holds bluetooth device information.
type Device struct {
	Path          string
	Name          string
	Type          string
	Alias         string
	Address       string
	AddressType   string
	Adapter       string
	Modalias      string
	UUIDs         []string
	Paired        bool
	Connected     bool
	Trusted       bool
	Blocked       bool
	Bonded        bool
	LegacyPairing bool
	RSSI          int16
	Class         uint32
	Percentage    int
}

// HaveService checks if the device has the specified service.
func (d Device) HaveService(service uint32) bool {
	return ServiceExists(d.UUIDs, service)
}

// CallDevice is used to interact with the bluez Device dbus interface.
// https://git.kernel.org/pub/scm/bluetooth/bluez.git/tree/doc/device-api.txt
func (b *Bluez) CallDevice(devicePath, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	path := dbus.ObjectPath(devicePath)
	return b.conn.Object(dbusBluezName, path).Call("org.bluez.Device1."+method, flags, args...)
}

// Pair will attempt to pair a bluetooth device that is in pairing mode.
func (b *Bluez) Pair(devicePath string) error {
	return b.CallDevice(devicePath, "Pair", 0).Store()
}

// CancelPairing will cancel a pairing attempt.
func (b *Bluez) CancelPairing(devicePath string) error {
	return b.CallDevice(devicePath, "CancelPairing", 0).Store()
}

// Connect will attempt to connect an already paired bluetooth device
// to an adapter.
func (b *Bluez) Connect(devicePath string) error {
	return b.CallDevice(devicePath, "Connect", 0).Store()
}

// Disconnect will remove the bluetooth device from the adapter.
func (b *Bluez) Disconnect(devicePath string) error {
	return b.CallDevice(devicePath, "Disconnect", 0).Store()
}

// RemoveDevice will permantently remove the bluetooth device from the adapter.
// Once a device is removed, it can only be added again by being paired.
func (b *Bluez) RemoveDevice(devicePath string) error {
	adapter := filepath.Dir(devicePath)

	return b.CallAdapter(adapter, "RemoveDevice", 0, dbus.ObjectPath(devicePath)).Store()
}

// GetDevice returns a Device with the provided device path.
func (b *Bluez) GetDevice(devicePath string) Device {
	return b.getDeviceFromStore(devicePath)
}

// GetDevices gets the stored devices.
func (b *Bluez) GetDevices() []Device {
	b.StoreLock.Lock()
	defer b.StoreLock.Unlock()

	var devices []Device

	store, ok := b.Store[b.GetCurrentAdapter().Path]
	if !ok {
		return nil
	}

	for _, device := range store.Devices {
		if device.Paired || device.Trusted || device.Blocked {
			devices = append([]Device{device}, devices...)
			continue
		}

		devices = append(devices, device)
	}

	return devices
}

// ConvertToDevices converts a map of dbus objects to a common Device structure.
func (b *Bluez) ConvertToDevice(path string, values map[string]dbus.Variant, devices *[]Device) error {
	/*
		org.bluez.Device1
			Icon => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"audio-card"}
			LegacyPairing => dbus.Variant{sig:dbus.Signature{str:"b"}, value:false}
			Address => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"2C:41:A1:49:37:CF"}
			Trusted => dbus.Variant{sig:dbus.Signature{str:"b"}, value:false}
			Connected => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
			Paired => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
			RSSI => dbus.Variant{sig:dbus.Signature{str:"n"}, value:-36}
			Modalias => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"bluetooth:v009Ep4020d0251"}
			Name => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"Bose QC35 II"}
			UUIDs => dbus.Variant{sig:dbus.Signature{str:"as"}, value:[]string{"00000000-deca-fade-deca-deafdecacaff", "00001101-0000-1000-8000-00805f9b34fb", "00001108-0000-1000-8000-00805f9b34fb", "0000110b-0000-1000-8000-00805f9b34fb", "0000110c-0000-1000-8000-00805f9b34fb", "0000110e-0000-1000-8000-00805f9b34fb", "0000111e-0000-1000-8000-00805f9b34fb", "00001200-0000-1000-8000-00805f9b34fb", "81c2e72a-0591-443e-a1ff-05f988593351", "f8d1fbe4-7966-4334-8024-ff96c9330e15"}}
			Adapter => dbus.Variant{sig:dbus.Signature{str:"o"}, value:"/org/bluez/hci0"}
			Blocked => dbus.Variant{sig:dbus.Signature{str:"b"}, value:false}
			Alias => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"Bose QC35 II"}
			Class => dbus.Variant{sig:dbus.Signature{str:"u"}, value:0x240418}

	*/
	var device Device

	if err := DecodeVariantMap(values, &device, "Name", "Address"); err != nil {
		return err
	}

	device.Path = path
	device.Type = GetDeviceType(device.Class)
	if p, err := b.GetBatteryPercentage(path); err == nil {
		device.Percentage = int(p)
	}

	if devices != nil {
		*devices = append(*devices, device)
	}

	return nil
}

// GetDeviceType parses the device class and returns its type.
//
//gocyclo:ignore
func GetDeviceType(class uint32) string {
	/*
		Adapted from:
		https://gitlab.freedesktop.org/upower/upower/-/blob/master/src/linux/up-device-bluez.c#L64
	*/
	switch (class & 0x1f00) >> 8 {
	case 0x01:
		return "Computer"

	case 0x02:
		switch (class & 0xfc) >> 2 {
		case 0x01, 0x02, 0x03, 0x05:
			return "Phone"

		case 0x04:
			return "Modem"
		}

	case 0x03:
		return "Network"

	case 0x04:
		switch (class & 0xfc) >> 2 {
		case 0x01, 0x02:
			return "Headset"

		case 0x05:
			return "Speakers"

		case 0x06:
			return "Headphones"

		case 0x0b, 0x0c, 0x0d:
			return "Video"

		default:
			return "Audio device"
		}

	case 0x05:
		switch (class & 0xc0) >> 6 {
		case 0x00:
			switch (class & 0x1e) >> 2 {
			case 0x01, 0x02:
				return "Gaming input"

			case 0x03:
				return "Remote control"
			}

		case 0x01:
			return "Keyboard"

		case 0x02:
			switch (class & 0x1e) >> 2 {
			case 0x05:
				return "Tablet"

			default:
				return "Mouse"
			}
		}

	case 0x06:
		if (class & 0x80) > 0 {
			return "Printer"
		}
		if (class & 0x40) > 0 {
			return "Scanner"
		}
		if (class & 0x20) > 0 {
			return "Camera"
		}
		if (class & 0x10) > 0 {
			return "Monitor"
		}

	case 0x07:
		return "Wearable"

	case 0x08:
		return "Toy"
	}

	return "Unknown"
}

// GetDeviceProperties gathers all the properties for a bluetooth device.
func (b *Bluez) GetDeviceProperties(devicePath string) (map[string]dbus.Variant, error) {
	result := make(map[string]dbus.Variant)
	path := dbus.ObjectPath(devicePath)
	if err := b.conn.Object(dbusBluezName, path).Call(dbusPropertiesGetAllPath, 0, dbusBluezDeviceIface).Store(&result); err != nil {
		return result, err
	}

	return result, nil
}

// GetBatteryPercentage gets the battery percentage of a device.
func (b *Bluez) GetBatteryPercentage(devicePath string) (byte, error) {
	var result byte
	path := dbus.ObjectPath(devicePath)
	if err := b.conn.Object(dbusBluezName, path).
		Call("org.freedesktop.DBus.Properties.Get", 0, dbusBluezBatteryIface, "Percentage").Store(&result); err != nil {
		return result, err
	}

	return result, nil
}

// SetDeviceProperty can be used to set certain properties for a bluetooth device.
func (b *Bluez) SetDeviceProperty(devicePath, key string, value interface{}) error {
	path := dbus.ObjectPath(devicePath)
	return b.conn.Object(dbusBluezName, path).Call("org.freedesktop.DBus.Properties.Set", 0, dbusBluezDeviceIface, key, dbus.MakeVariant(value)).Store()
}

// addDeviceToStore adds a device to the store.
func (b *Bluez) addDeviceToStore(device Device) {
	b.StoreLock.Lock()
	defer b.StoreLock.Unlock()

	var store StoreObject

	if _, ok := b.Store[device.Adapter]; ok {
		store = b.Store[device.Adapter]
	}

	if store.Devices == nil {
		store.Devices = make(map[string]Device)
	}

	store.Devices[device.Path] = device
	b.Store[device.Adapter] = store
}

// removeDeviceFromStore removes a device from the store,
// using its devicePath.
func (b *Bluez) removeDeviceFromStore(devicePath string) {
	b.StoreLock.Lock()
	defer b.StoreLock.Unlock()

	adapterPath := filepath.Dir(devicePath)

	store := b.Store[adapterPath]
	delete(store.Devices, devicePath)

	b.Store[adapterPath] = store
}

// getDeviceFromStore gets a device from the store,
// using its device path.
func (b *Bluez) getDeviceFromStore(devicePath string) Device {
	b.StoreLock.Lock()
	defer b.StoreLock.Unlock()

	adapterPath := filepath.Dir(devicePath)

	store, ok := b.Store[adapterPath]
	if !ok {
		return Device{}
	}

	device, ok := store.Devices[devicePath]
	if !ok {
		return Device{}
	}

	return device
}
