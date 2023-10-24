package bluez

import (
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/pkg/errors"
)

const (
	dbusBluezName            = "org.bluez"
	dbusPropertiesGetPath    = "org.freedesktop.DBus.Properties.Get"
	dbusPropertiesGetAllPath = "org.freedesktop.DBus.Properties.GetAll"
	dbusObjectManagerPath    = "org.freedesktop.DBus.ObjectManager.GetManagedObjects"
)

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

	CurrentPlayer dbus.ObjectPath
	PlayerLock    sync.Mutex
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
	if err := b.conn.Object(dbusBluezName, "/").Call(dbusObjectManagerPath, 0).Store(&result); err != nil {
		return result, err
	}
	return result, nil
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

// ParseSignalData parses bluez DBus signal data.
//
//gocyclo:ignore
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
		case dbusBluezAdapterIface:
			adapter := b.getAdapterFromStore(string(signal.Path))
			if adapter == (Adapter{}) {
				return nil
			}

			for prop, value := range objMap {
				switch prop {
				case "Powered":
					adapter.Powered = value.Value().(bool)

				case "Discoverable":
					adapter.Discoverable = value.Value().(bool)

				case "Pairable":
					adapter.Pairable = value.Value().(bool)

				case "Discovering":
					adapter.Discovering = value.Value().(bool)
				}
			}

			b.addAdapterToStore(adapter)

			return adapter

		case dbusBluezDeviceIface:
			device := b.getDeviceFromStore(string(signal.Path))
			if device.Path == "" {
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

				case "Bonded":
					device.Bonded = value.Value().(bool)

				case "RSSI":
					device.RSSI = value.Value().(int16)
				}
			}

			b.addDeviceToStore(device)

			return device

		case dbusBluezMediaPlayerIface:
			var media MediaProperties

			for prop, value := range objMap {
				switch prop {
				case "Status":
					media.Status = value.Value().(string)

				case "Position":
					media.Position = value.Value().(uint32)

				case "Track":
					media.Track = getTrackProperties(value.Value().(map[string]dbus.Variant))
				}
			}

			return media

		case dbusBluezBatteryIface:
			device := b.getDeviceFromStore(string(signal.Path))
			if device.Path == "" {
				return nil
			}

			device.Percentage = int(objMap["Percentage"].Value().(byte))
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
			case dbusBluezAdapterIface:
				adapterPath := string(objPath)

				adapters := b.ConvertToAdapters(adapterPath, objMap)
				for _, adapter := range adapters {
					b.addAdapterToStore(adapter)
				}

				return adapters

			case dbusBluezDeviceIface:
				devicePath := string(objPath)

				devices := b.ConvertToDevices(devicePath, objMap)
				for _, device := range devices {
					b.addDeviceToStore(device)
				}

				devResultMap := make(map[string][]Device)
				devResultMap[devicePath] = devices

				return devResultMap

			case dbusBluezBatteryIface:
				devicePath := string(objPath)

				device := b.getDeviceFromStore(devicePath)
				if device.Path == "" {
					return nil
				}

				device.Percentage = int(objMap[iftype]["Percentage"].Value().(byte))
				b.addDeviceToStore(device)

				return map[string][]Device{devicePath: {device}}
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
			case dbusBluezAdapterIface:
				adapterPath := string(objPath)
				b.removeAdapterFromStore(adapterPath)

				return adapterPath

			case dbusBluezDeviceIface:
				devicePath := string(objPath)
				b.removeDeviceFromStore(devicePath)

				return devicePath

			case dbusBluezBatteryIface:
				devicePath := string(objPath)

				device := b.getDeviceFromStore(devicePath)
				if device.Path == "" {
					return nil
				}

				device.Percentage = 0
				b.addDeviceToStore(device)

				return nil
			}
		}
	}

	return nil
}
