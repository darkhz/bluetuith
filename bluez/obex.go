package bluez

import (
	"path/filepath"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/pkg/errors"
)

const (
	dbusObexName            = "org.bluez.obex"
	dbusObexClientIface     = "org.bluez.obex.Client1"
	dbusObexSessionIface    = "org.bluez.obex.Session1"
	dbusObexTransferIface   = "org.bluez.obex.Transfer1"
	dbusObexObjectPushIface = "org.bluez.obex.ObjectPush1"

	dbusObexPath = dbus.ObjectPath("/org/bluez/obex")
)

// ObexSessionProperties describes the session properties
// of an OBEX transfer.
type ObexSessionProperties struct {
	Root        string
	Target      string
	Source      string
	Destination string
}

// ObexTransferProperties describes the transfer properties
// of an OBEX transfer.
type ObexTransferProperties struct {
	Name     string
	Type     string
	Status   string
	Filename string

	Size        uint64
	Transferred uint64

	Session dbus.ObjectPath
}

// ObexProperties stores the session and transfer paths of
// an OBEX transfer, along with their properties.
type ObexProperties struct {
	SessionPath       dbus.ObjectPath
	SessionProperties ObexSessionProperties

	TransferPath       dbus.ObjectPath
	TransferProperties ObexTransferProperties
}

// Obex represents an OBEX connection. It also stores
// information about the currently running transfers.
type Obex struct {
	conn *dbus.Conn

	Store     map[dbus.ObjectPath]ObexProperties
	StoreLock sync.Mutex
}

// NewObex returns a new Obex.
func NewObex() (*Obex, error) {
	var o *Obex

	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create dbus session bus:")
	}

	o = &Obex{
		conn:  conn,
		Store: make(map[dbus.ObjectPath]ObexProperties),
	}

	return o, nil
}

// Close closes the OBEX connection.
func (o *Obex) Close() {
	o.conn.Close()
}

// Conn returns the current SessionBus connection.
func (o *Obex) Conn() *dbus.Conn {
	return o.conn
}

// CreateSession creates a new OBEX transfer session.
func (o *Obex) CreateSession(address string) (dbus.ObjectPath, error) {
	var sessionPath dbus.ObjectPath

	args := make(map[string]interface{})
	args["Target"] = "opp"

	if err := o.CallClient("CreateSession", address, args).Store(&sessionPath); err != nil {
		return "", err
	}

	sessionProperties, err := o.GetSessionProperties(sessionPath)
	if err != nil {
		return "", err
	}

	o.addSessionPropertiesToStore(sessionPath, sessionProperties)

	return sessionPath, nil
}

// SendFile sends a file to the target device.
func (o *Obex) SendFile(sessionPath dbus.ObjectPath, path string) (dbus.ObjectPath, ObexTransferProperties, error) {
	var transferPath dbus.ObjectPath

	transferPropertyMap := make(map[string]dbus.Variant)
	if err := o.CallObjectPush(sessionPath, "SendFile", path).Store(&transferPath, &transferPropertyMap); err != nil {
		return "", ObexTransferProperties{}, err
	}

	transferProperties := o.GetTransferProperties(transferPropertyMap)
	o.addTransferPropertiesToStore(transferPath, transferProperties)

	return transferPath, transferProperties, nil
}

// ReceiveFile returns a path where the OBEX daemon (obexd) will receive the file, along with
// the transfer properties.
func (o *Obex) ReceiveFile(sessionPath, transferPath dbus.ObjectPath) (string, string, ObexTransferProperties, error) {
	var sessionProperty ObexSessionProperties
	var transferProperty ObexTransferProperties

	objMap, err := o.ManagedObjects()
	if err != nil {
		return "", "", ObexTransferProperties{}, err
	}

	for path, valueMap := range objMap {
		switch path {
		case sessionPath:
			for iface, value := range valueMap {
				if iface != dbusObexSessionIface {
					continue
				}

				sessionProperty, _ = o.GetSessionProperties(sessionPath, value)
				o.addSessionPropertiesToStore(sessionPath, sessionProperty)
			}

		case transferPath:
			for iface, value := range valueMap {
				if iface != dbusObexTransferIface {
					continue
				}

				transferProperty = o.GetTransferProperties(value)
				if transferProperty.Status == "error" {
					return "", "", ObexTransferProperties{}, errors.New("Transfer error")
				}

				o.addTransferPropertiesToStore(transferPath, transferProperty)
			}
		}
	}

	if sessionProperty == (ObexSessionProperties{}) || transferProperty == (ObexTransferProperties{}) {
		return "", "", ObexTransferProperties{}, errors.New("Cannot get obex properties")
	}

	return filepath.Join(sessionProperty.Root, transferProperty.Name), sessionProperty.Destination, transferProperty, nil
}

// CancelTransfer cancels the transfer.
func (o *Obex) CancelTransfer(transferPath dbus.ObjectPath) error {
	return o.CallTransfer(transferPath, "Cancel").Store()
}

// SuspendTransfer suspends the transfer.
func (o *Obex) SuspendTransfer(transferPath dbus.ObjectPath) error {
	return o.CallTransfer(transferPath, "Suspend").Store()
}

// ResumeTransfer resumes the transfer.
func (o *Obex) ResumeTransfer(transferPath dbus.ObjectPath) error {
	return o.CallTransfer(transferPath, "Resume").Store()
}

// RemoveSession removes the OBEX transfer session and cancels any pending transfers.
func (o *Obex) RemoveSession(sessionPath dbus.ObjectPath) error {
	o.removePropertiesFromStore(sessionPath)
	return o.CallClient("RemoveSession", sessionPath).Store()
}

// GetSessionProperties converts a map of OBEX session properties to ObexSessionProperties.
func (o *Obex) GetSessionProperties(sessionPath dbus.ObjectPath, sprop ...map[string]dbus.Variant) (ObexSessionProperties, error) {
	var root, source, target string
	var sessionProperties ObexSessionProperties

	props := make(map[string]dbus.Variant)
	if sprop == nil {
		if err := o.conn.Object(dbusObexName, sessionPath).Call(dbusPropertiesGetAllPath, 0, dbusObexSessionIface).Store(&props); err != nil {
			return ObexSessionProperties{}, err
		}
	} else {
		props = sprop[0]
	}

	if t, ok := props["Root"].Value().(string); ok {
		root = t
	}

	if s, ok := props["Source"].Value().(string); ok {
		source = s
	}

	if t, ok := props["Target"].Value().(string); ok {
		target = t
	}

	sessionProperties = ObexSessionProperties{
		Root:        root,
		Target:      target,
		Source:      source,
		Destination: props["Destination"].Value().(string),
	}

	return sessionProperties, nil
}

// GetTransferProperties converts a map of transfer properties to ObexTransferProperties.
func (o *Obex) GetTransferProperties(props map[string]dbus.Variant) ObexTransferProperties {
	var size, transferred uint64
	var name, ftype, fname string

	if n, ok := props["Name"].Value().(string); ok {
		name = n
	}

	if t, ok := props["Type"].Value().(string); ok {
		ftype = t
	}

	if f, ok := props["Filename"].Value().(string); ok {
		fname = f
	}

	if s, ok := props["Size"].Value().(uint64); ok {
		size = s
	}

	if t, ok := props["Transferred"].Value().(uint64); ok {
		transferred = t
	}

	return ObexTransferProperties{
		Name:     name,
		Type:     ftype,
		Status:   props["Status"].Value().(string),
		Filename: fname,

		Size:        size,
		Transferred: transferred,

		Session: props["Session"].Value().(dbus.ObjectPath),
	}
}

// ManagedObjects gets the currently managed objects from the OBEX DBus.
func (o *Obex) ManagedObjects() (map[dbus.ObjectPath]map[string]map[string]dbus.Variant, error) {
	result := make(map[dbus.ObjectPath]map[string]map[string]dbus.Variant)
	if err := o.conn.Object(dbusObexName, "/").Call(dbusObjectManagerPath, 0).Store(&result); err != nil {
		return result, err
	}

	return result, nil
}

// WatchSignal will register a signal and watch for events from the OBEX DBus interface.
func (o *Obex) WatchSignal() chan *dbus.Signal {
	signalMatch := "type='signal', sender='org.bluez.obex'"
	o.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, signalMatch)
	ch := make(chan *dbus.Signal, 1)
	o.conn.Signal(ch)
	return ch
}

// ParseSignalData parses OBEX DBus signal data.
func (o *Obex) ParseSignalData(signal *dbus.Signal) interface{} {
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
		case dbusObexTransferIface:
			sessionPath := dbus.ObjectPath(filepath.Dir(string(signal.Path)))

			props, ok := o.getPropertiesFromStore(sessionPath)
			if !ok {
				break
			}

			for prop, value := range objMap {
				switch prop {
				case "Status":
					props.TransferProperties.Status = value.Value().(string)

				case "Transferred":
					props.TransferProperties.Transferred = value.Value().(uint64)
				}
			}

			o.addTransferPropertiesToStore(signal.Path, props.TransferProperties)

			return props
		}
	}

	return nil
}

// CallClient calls the Client1 interface with the provided method.
func (o *Obex) CallClient(method string, args ...interface{}) *dbus.Call {
	return o.conn.Object(dbusObexName, dbusObexPath).Call(dbusObexClientIface+"."+method, 0, args...)
}

// CallObjectPush calls the ObjectPush1 interface with the provided method.
func (o *Obex) CallObjectPush(sessionPath dbus.ObjectPath, method string, args ...interface{}) *dbus.Call {
	return o.conn.Object(dbusObexName, sessionPath).Call(dbusObexObjectPushIface+"."+method, 0, args...)
}

// CallTransfer calls the Transfer1 interface with the provided method.
func (o *Obex) CallTransfer(transferPath dbus.ObjectPath, method string, args ...interface{}) *dbus.Call {
	return o.conn.Object(dbusObexName, transferPath).Call(dbusObexTransferIface+"."+method, 0, args...)
}

// addSessionPropertiesToStore adds a session path and property to the store.
func (o *Obex) addSessionPropertiesToStore(sessionPath dbus.ObjectPath, sessionProperties ObexSessionProperties) {
	o.StoreLock.Lock()
	defer o.StoreLock.Unlock()

	obexProperties := o.Store[sessionPath]
	obexProperties.SessionPath = sessionPath
	obexProperties.SessionProperties = sessionProperties

	o.Store[sessionPath] = obexProperties
}

// addTransferPropertiesToStore adds a transfer path and property to the store.
func (o *Obex) addTransferPropertiesToStore(transferPath dbus.ObjectPath, transferProperties ObexTransferProperties) {
	o.StoreLock.Lock()
	defer o.StoreLock.Unlock()

	obexProperties := o.Store[transferProperties.Session]
	obexProperties.TransferPath = transferPath
	obexProperties.TransferProperties = transferProperties

	o.Store[transferProperties.Session] = obexProperties
}

// getPropertiesFromStore gets the ObexProperties from the store.
func (o *Obex) getPropertiesFromStore(sessionPath dbus.ObjectPath) (ObexProperties, bool) {
	o.StoreLock.Lock()
	defer o.StoreLock.Unlock()

	props, ok := o.Store[sessionPath]

	return props, ok
}

// removePropertiesFromStore removes the session path and ObexProperties from the store.
func (o *Obex) removePropertiesFromStore(sessionPath dbus.ObjectPath) {
	o.StoreLock.Lock()
	defer o.StoreLock.Unlock()

	delete(o.Store, sessionPath)
}
