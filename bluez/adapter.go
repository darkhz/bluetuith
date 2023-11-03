package bluez

import (
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
)

const dbusBluezAdapterIface = "org.bluez.Adapter1"

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

	Lock *semaphore.Weighted
}

// CallAdapter is used to interact with the bluez Adapter dbus interface.
// https://git.kernel.org/pub/scm/bluetooth/bluez.git/tree/doc/adapter-api.txt
func (b *Bluez) CallAdapter(adapter, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	return b.conn.Object(dbusBluezName, dbus.ObjectPath(adapter)).Call("org.bluez.Adapter1."+method, flags, args...)
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
func (b *Bluez) ConvertToAdapter(path string, values map[string]dbus.Variant, adapters *[]Adapter) error {
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

	var adapter Adapter

	if err := DecodeVariantMap(values, &adapter, "Address"); err != nil {
		return err
	}

	adapter.Path = path
	adapter.Lock = semaphore.NewWeighted(1)

	if adapters != nil {
		*adapters = append(*adapters, adapter)
	}

	return nil
}

// GetAdapterProperties gathers all the properties for a bluetooth adapter.
func (b *Bluez) GetAdapterProperties(adapterPath string) (map[string]dbus.Variant, error) {
	result := make(map[string]dbus.Variant)
	path := dbus.ObjectPath(adapterPath)
	if err := b.conn.Object(dbusBluezName, path).Call(dbusPropertiesGetAllPath, 0, dbusBluezAdapterIface).Store(&result); err != nil {
		return result, err
	}

	return result, nil
}

// SetAdapterProperty can be used to set certain properties for a bluetooth adapter.
func (b *Bluez) SetAdapterProperty(adapterPath, key string, value interface{}) error {
	path := dbus.ObjectPath(adapterPath)

	return b.conn.Object(dbusBluezName, path).Call("org.freedesktop.DBus.Properties.Set", 0, dbusBluezAdapterIface, key, dbus.MakeVariant(value)).Store()
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
