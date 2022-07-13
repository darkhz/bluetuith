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
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/pkg/errors"
)

const (
	dbusBluetoothPath        = "org.bluez"
	dbusPropertiesGetAllPath = "org.freedesktop.DBus.Properties.GetAll"
	dbusObjectManagerPath    = "org.freedesktop.DBus.ObjectManager.GetManagedObjects"

	bluezAdapterObject = "org.bluez.Adapter1"
	bluezDeviceObject  = "org.bluez.Device1"
)

// Adapter holds the bluetooth device adapter installed for a system.
type Adapter struct {
	Path         string
	Name         string
	Alias        string
	Address      string
	Discoverable bool
	Pairable     bool
	Powered      bool
	Discovering  bool
}

// Device holds bluetooth device information.
type Device struct {
	Path      string
	Name      string
	Type      string
	Alias     string
	Address   string
	Adapter   string
	Paired    bool
	Connected bool
	Trusted   bool
	Blocked   bool
	RSSI      int16
	Class     uint32
}

// StoreObject holds an Adapter and the Devices that belong to it.
// Each device is stored into Devices with the device adapter path
// (held by (Device).Adapter) as the identifier.
type StoreObject struct {
	Adapter Adapter
	Devices map[string]Device
}

// Bluez represents an overview of the bluetooth adapters and
// devices installed and configured on a system. A connection
// to `bluez` dbus is also used to interact with the bluetooth
// server on a system.
type Bluez struct {
	conn *dbus.Conn

	CurrentAdapter Adapter
	AdapterLock    sync.Mutex

	Store     map[string]StoreObject
	StoreLock sync.Mutex
}

// NewBluez returns a new Bluez.
func NewBluez() (*Bluez, error) {
	var b *Bluez

	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create dbus system bus:")
	}

	b = &Bluez{
		conn:  conn,
		Store: make(map[string]StoreObject),
	}
	if err := b.RefreshStore(); err != nil {
		return nil, errors.Wrapf(err, "unable to populate cache")
	}
	if err := b.SetCurrentAdapter(); err != nil {
		return nil, err
	}

	return b, nil
}

// Close closes the bluez connection.
func (b *Bluez) Close() {
	b.conn.Close()
}

// Conn returns the current SystemBus connection.
func (b *Bluez) Conn() *dbus.Conn {
	return b.conn
}

// RefreshStore will query system for known bluetooth adapters and devices
// and will store them on the Bluez structure.
func (b *Bluez) RefreshStore() error {
	b.StoreLock.Lock()
	defer b.StoreLock.Unlock()

	var store StoreObject

	results, err := b.ManagedObjects()
	if err != nil {
		return err
	}

	devices := []Device{}
	adapters := []Adapter{}
	for k, v := range results {
		adapters = append(adapters, b.ConvertToAdapters(string(k), v)...)
		devices = append(devices, b.ConvertToDevices(string(k), v)...)
	}

	for _, adapter := range adapters {
		if adapter == (Adapter{}) {
			continue
		}

		store.Adapter = adapter
		store.Devices = make(map[string]Device)

		for _, device := range devices {
			if device.Adapter != adapter.Path {
				continue
			}

			store.Devices[device.Path] = device
		}

		b.Store[adapter.Path] = store
	}

	return nil
}

// ManagedObjects gets all bluetooth devices and adapters that are currently managed by bluez.
func (b *Bluez) ManagedObjects() (map[dbus.ObjectPath]map[string]map[string]dbus.Variant, error) {
	result := make(map[dbus.ObjectPath]map[string]map[string]dbus.Variant)
	if err := b.conn.Object(dbusBluetoothPath, "/").Call(dbusObjectManagerPath, 0).Store(&result); err != nil {
		return result, err
	}
	return result, nil
}

// CallDevice is used to interact with the bluez Device dbus interface.
// https://git.kernel.org/pub/scm/bluetooth/bluez.git/tree/doc/device-api.txt
func (b *Bluez) CallDevice(devicePath, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	path := dbus.ObjectPath(devicePath)
	return b.conn.Object(dbusBluetoothPath, path).Call("org.bluez.Device1."+method, flags, args...)
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
		devices = append(devices, device)
	}

	return devices
}

// ConvertToDevices converts a map of dbus objects to a common Device structure.
func (b *Bluez) ConvertToDevices(path string, values map[string]map[string]dbus.Variant) []Device {
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
	devices := []Device{}
	for k, v := range values {
		var name string
		var rssi int16
		var class uint32

		if n, ok := v["Name"].Value().(string); ok {
			name = n
		}

		if i, ok := v["RSSI"].Value().(int16); ok {
			rssi = i
		}

		if c, ok := v["Class"].Value().(uint32); ok {
			class = c
		}

		switch k {
		case bluezDeviceObject:
			adapter, _ := v["Adapter"].Value().(dbus.ObjectPath)
			devices = append(devices, Device{
				Path:      path,
				Name:      name,
				Class:     class,
				RSSI:      rssi,
				Type:      GetDeviceType(class),
				Alias:     v["Alias"].Value().(string),
				Address:   v["Address"].Value().(string),
				Adapter:   string(adapter),
				Paired:    v["Paired"].Value().(bool),
				Connected: v["Connected"].Value().(bool),
				Trusted:   v["Trusted"].Value().(bool),
				Blocked:   v["Blocked"].Value().(bool),
			})
		}
	}

	return devices
}

//gocyclo: ignore
// GetDeviceType parses the device class and returns its type.
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
	if err := b.conn.Object(dbusBluetoothPath, path).Call(dbusPropertiesGetAllPath, 0, bluezDeviceObject).Store(&result); err != nil {
		return result, err
	}

	return result, nil
}

// SetDeviceProperty can be used to set certain properties for a bluetooth device.
func (b *Bluez) SetDeviceProperty(devicePath, key string, value interface{}) error {
	path := dbus.ObjectPath(devicePath)
	return b.conn.Object(dbusBluetoothPath, path).Call("org.freedesktop.DBus.Properties.Set", 0, bluezDeviceObject, key, dbus.MakeVariant(value)).Store()
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

// CallAdapter is used to interact with the bluez Adapter dbus interface.
// https://git.kernel.org/pub/scm/bluetooth/bluez.git/tree/doc/adapter-api.txt
func (b *Bluez) CallAdapter(adapter, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	return b.conn.Object(dbusBluetoothPath, dbus.ObjectPath(adapter)).Call("org.bluez.Adapter1."+method, flags, args...)
}

// StartDiscovery will put the adapter into "discovering" mode, which means
// the bluetooth device will be able to discover other bluetooth devices
// that are in pairing mode.
func (b *Bluez) StartDiscovery(adapter string) error {
	return b.CallAdapter(adapter, "StartDiscovery", 0).Store()
}

// StopDiscovery will stop the  "discovering" mode, which means the bluetooth device will
// no longer be able to discover other bluetooth devices that are in pairing mode.
func (b *Bluez) StopDiscovery(adapter string) error {
	return b.CallAdapter(adapter, "StopDiscovery", 0).Store()
}

// Power sets the powered state of the adapter.
func (b *Bluez) Power(adapterPath string, enable bool) error {
	currentAdapter := b.GetCurrentAdapter()
	currentAdapter.Powered = enable

	b.SetCurrentAdapter(currentAdapter)

	if err := b.SetAdapterProperty(adapterPath, "Powered", enable); err != nil {
		return err
	}

	return b.SetAdapterProperty(adapterPath, "Pairable", enable)
}

// GetCurrentAdapter gets the current adapter.
func (b *Bluez) GetCurrentAdapter() Adapter {
	b.AdapterLock.Lock()
	defer b.AdapterLock.Unlock()

	return b.CurrentAdapter
}

// SetCurrentAdapter sets the current adapter.
func (b *Bluez) SetCurrentAdapter(adapter ...Adapter) error {
	b.AdapterLock.Lock()
	defer b.AdapterLock.Unlock()

	if adapter == nil {
		adapters := b.GetAdapters()
		if len(adapters) == 0 {
			return errors.New("No adapters found")
		}
		for _, a := range adapters {
			adapter = append(adapter, a)
			break
		}
		if len(adapter) == 0 {
			adapter = append(adapter, adapters[0])
		}
	}

	b.CurrentAdapter = adapter[0]

	return nil
}

// GetAdapters gets the stored adapters.
func (b *Bluez) GetAdapters() []Adapter {
	b.StoreLock.Lock()
	defer b.StoreLock.Unlock()

	var adapters []Adapter

	for _, store := range b.Store {
		adapters = append(adapters, store.Adapter)
	}

	return adapters
}

// GetAdapterID gets the adapter ID from the adapter path.
func GetAdapterID(adapterPath string) string {
	currentAdapter := strings.Split(adapterPath, "/")

	return currentAdapter[len(currentAdapter)-1]
}

// GetCurrentAdapterID gets the adapter ID from the current
// adapter's path.
func (b *Bluez) GetCurrentAdapterID() string {
	return GetAdapterID(b.GetCurrentAdapter().Path)
}

// ConvertToAdapters converts a map of dbus objects to a common Adapter structure.
func (b *Bluez) ConvertToAdapters(path string, values map[string]map[string]dbus.Variant) []Adapter {
	/*
		/org/bluez/hci0
			org.bluez.Adapter1
					Discoverable => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
					UUIDs => dbus.Variant{sig:dbus.Signature{str:"as"}, value:[]string{"00001112-0000-1000-8000-00805f9b34fb", "00001801-0000-1000-8000-00805f9b34fb", "0000110e-0000-1000-8000-00805f9b34fb", "00001800-0000-1000-8000-00805f9b34fb", "00001200-0000-1000-8000-00805f9b34fb", "0000110c-0000-1000-8000-00805f9b34fb", "0000110b-0000-1000-8000-00805f9b34fb", "0000110a-0000-1000-8000-00805f9b34fb"}}
					Modalias => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"usb:v1D6Bp0246d0525"}
					Pairable => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
					DiscoverableTimeout => dbus.Variant{sig:dbus.Signature{str:"u"}, value:0x0}
					PairableTimeout => dbus.Variant{sig:dbus.Signature{str:"u"}, value:0x0}
					Powered => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
					Class => dbus.Variant{sig:dbus.Signature{str:"u"}, value:0xc010c}
					Discovering => dbus.Variant{sig:dbus.Signature{str:"b"}, value:true}
					Address => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"9C:B6:D0:1C:BB:B0"}
					Name => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"jonathan-Blade"}
					Alias => dbus.Variant{sig:dbus.Signature{str:"s"}, value:"jonathan-Blade"}

	*/
	adapters := []Adapter{}
	for k1, v1 := range values {
		switch k1 {
		case bluezAdapterObject:
			adapters = append(adapters, Adapter{
				Path:         path,
				Name:         v1["Name"].Value().(string),
				Alias:        v1["Alias"].Value().(string),
				Address:      v1["Address"].Value().(string),
				Discoverable: v1["Discoverable"].Value().(bool),
				Pairable:     v1["Pairable"].Value().(bool),
				Powered:      v1["Powered"].Value().(bool),
				Discovering:  v1["Discovering"].Value().(bool),
			})
		}
	}

	return adapters
}

// GetAdapterProperties gathers all the properties for a bluetooth adapter.
func (b *Bluez) GetAdapterProperties(adapterPath string) (map[string]dbus.Variant, error) {
	result := make(map[string]dbus.Variant)
	path := dbus.ObjectPath(adapterPath)
	if err := b.conn.Object(dbusBluetoothPath, path).Call(dbusPropertiesGetAllPath, 0, bluezAdapterObject).Store(&result); err != nil {
		return result, err
	}

	return result, nil
}

// SetAdapterProperty can be used to set certain properties for a bluetooth adapter.
func (b *Bluez) SetAdapterProperty(adapterPath, key string, value interface{}) error {
	path := dbus.ObjectPath(adapterPath)

	return b.conn.Object(dbusBluetoothPath, path).Call("org.freedesktop.DBus.Properties.Set", 0, bluezAdapterObject, key, dbus.MakeVariant(value)).Store()
}

// addAdapterToStore adds an adapter to the store.
func (b *Bluez) addAdapterToStore(adapter Adapter) {
	b.StoreLock.Lock()
	defer b.StoreLock.Unlock()

	var store StoreObject

	if _, ok := b.Store[adapter.Path]; ok {
		store = b.Store[adapter.Path]
	}

	store.Adapter = adapter
	b.Store[adapter.Path] = store
}

// removeAdapterFromStore removes an adapter from the store,
// using its adapter path.
func (b *Bluez) removeAdapterFromStore(adapterPath string) {
	b.StoreLock.Lock()
	defer b.StoreLock.Unlock()

	delete(b.Store, adapterPath)
}

// getAdapterFromStore gets an adapter from the store,
// using its adapter path.
func (b *Bluez) getAdapterFromStore(adapterPath string) Adapter {
	b.StoreLock.Lock()
	defer b.StoreLock.Unlock()

	store, ok := b.Store[adapterPath]
	if !ok {
		return Adapter{}
	}

	return store.Adapter
}

// WatchSignal will register to receive events form the bluez dbus interface. Any
// events received are passed along to the returned channel for the caller to use.
func (b *Bluez) WatchSignal() chan *dbus.Signal {
	signalMatch := "type='signal', sender='org.bluez'"
	b.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, signalMatch)
	ch := make(chan *dbus.Signal, 1)
	b.conn.Signal(ch)
	return ch
}

//gocyclo: ignore
// ParseSignalData parses signal data.
func (b *Bluez) ParseSignalData(signal *dbus.Signal) interface{} {
	switch signal.Name {
	case "org.freedesktop.DBus.Properties.PropertiesChanged":
		objInterface, ok := signal.Body[0].(string)
		if !ok {
			return nil
		}

		objMap, ok := signal.Body[1].(map[string]dbus.Variant)
		if !ok {
			return nil
		}

		switch objInterface {
		case bluezAdapterObject:
			adapter := b.getAdapterFromStore(string(signal.Path))
			if adapter == (Adapter{}) {
				return nil
			}

			for prop, value := range objMap {
				switch prop {
				case "Powered":
					adapter.Powered = value.Value().(bool)

				case "Discovering":
					adapter.Discovering = value.Value().(bool)
				}
			}

			b.addAdapterToStore(adapter)

			return adapter

		case bluezDeviceObject:
			device := b.getDeviceFromStore(string(signal.Path))
			if device == (Device{}) {
				return nil
			}

			for prop, value := range objMap {
				switch prop {
				case "Connected":
					device.Connected = value.Value().(bool)

				case "Paired":
					device.Paired = value.Value().(bool)

				case "Trusted":
					device.Trusted = value.Value().(bool)

				case "RSSI":
					device.RSSI = value.Value().(int16)
				}
			}

			b.addDeviceToStore(device)

			return device
		}

	case "org.freedesktop.DBus.ObjectManager.InterfacesAdded":
		if len(signal.Body) != 2 {
			return nil
		}

		objPath, ok := signal.Body[0].(dbus.ObjectPath)
		if !ok {
			return nil
		}
		objMap, ok := signal.Body[1].(map[string]map[string]dbus.Variant)
		if !ok {
			return nil
		}

		for iftype := range objMap {
			switch iftype {
			case bluezAdapterObject:
				adapterPath := string(objPath)

				adapters := b.ConvertToAdapters(adapterPath, objMap)
				for _, adapter := range adapters {
					b.addAdapterToStore(adapter)
				}

				return adapters

			case bluezDeviceObject:
				devicePath := string(objPath)

				devices := b.ConvertToDevices(devicePath, objMap)
				for _, device := range devices {
					b.addDeviceToStore(device)
				}

				devResultMap := make(map[string][]Device)
				devResultMap[devicePath] = devices

				return devResultMap
			}
		}

	case "org.freedesktop.DBus.ObjectManager.InterfacesRemoved":
		objPath, ok := signal.Body[0].(dbus.ObjectPath)
		if !ok {
			return nil
		}

		objArray, ok := signal.Body[1].([]string)
		if !ok {
			return nil
		}

		for _, obj := range objArray {
			switch obj {
			case bluezAdapterObject:
				adapterPath := string(objPath)
				b.removeAdapterFromStore(adapterPath)

				return adapterPath

			case bluezDeviceObject:
				devicePath := string(objPath)
				b.removeDeviceFromStore(devicePath)

				return devicePath
			}
		}
	}

	return nil
}
