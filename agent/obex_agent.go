package agent

import (
	"errors"
	"path/filepath"

	"github.com/darkhz/bluetuith/ui"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

const (
	ObexAgentBluezName = "org.bluez.obex"

	ObexAgentIface        = "org.bluez.obex.Agent1"
	ObexAgentManagerIface = "org.bluez.obex.AgentManager1"

	ObexAgentManagerPath = dbus.ObjectPath("/org/bluez/obex")
	ObexAgentPath        = dbus.ObjectPath("/org/bluez/obex/agent/bluetuith")
)

var (
	obexAgent    *ObexAgent
	knownDevices []string
)

// ObexAgent describes an OBEX agent connection.
type ObexAgent struct {
	conn *dbus.Conn
}

// NewObexAgent returns a new ObexAgent.
func NewObexAgent() (*ObexAgent, error) {
	var agobex *ObexAgent

	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	agobex = &ObexAgent{
		conn: conn,
	}

	return agobex, nil
}

// SetupObexAgent sets up an OBEX agent.
func SetupObexAgent() error {
	var err error

	obexAgent, err = NewObexAgent()
	if err != nil {
		return err
	}

	if err := ExportObexAgent(); err != nil {
		return err
	}

	return RegisterObexAgent()
}

// RemoveObexAgent removes the OBEX agent.
func RemoveObexAgent() error {
	return UnregisterObexAgent()
}

// RegisterObexAgent registers the OBEX agent.
func RegisterObexAgent() error {
	return CallObexAgentManager("RegisterAgent", ObexAgentPath).Store()
}

// ExportObexAgent exports all ObexAgent methods to the OBEX DBus interface.
func ExportObexAgent() error {
	err := obexAgent.conn.Export(obexAgent, ObexAgentPath, ObexAgentIface)
	if err != nil {
		return err
	}

	node := &introspect.Node{
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			{
				Name:    ObexAgentIface,
				Methods: introspect.Methods(obexAgent),
			},
		},
	}

	return obexAgent.conn.Export(introspect.NewIntrospectable(node), ObexAgentPath, dbusIntrospectable)
}

// UnregisterObexAgent unregisters the OBEX agent.
func UnregisterObexAgent() error {
	return CallObexAgentManager("UnregisterAgent", ObexAgentPath).Store()
}

// CallObexAgentManager calls the OBEX AgentManager1 interface with the provided arguments.
func CallObexAgentManager(method string, args ...interface{}) *dbus.Call {
	return obexAgent.conn.Object(ObexAgentBluezName, ObexAgentManagerPath).Call(ObexAgentManagerIface+"."+method, 0, args...)
}

// AuthorizePush asks for confirmation before receiving a transfer from the host device.
// If the "Accept all" reply is given in response to the confirmation query, the device
// will be added to a list of known devices and all transfers will be automatically accepted.
func (o *ObexAgent) AuthorizePush(transferPath dbus.ObjectPath) (string, *dbus.Error) {
	var msg, reply string

	adapter := ui.BluezConn.GetCurrentAdapter()
	if !adapter.Lock.TryAcquire(1) {
		return "", dbus.MakeFailedError(errors.New("Operation in progress"))
	}

	sessionPath := dbus.ObjectPath(filepath.Dir(string(transferPath)))

	path, device, transferProps, err := ui.ObexConn.ReceiveFile(sessionPath, transferPath)
	if err != nil {
		return "", dbus.MakeFailedError(err)
	}

	for _, knownDevice := range knownDevices {
		if device == knownDevice {
			goto SkipAuthentication
		}
	}

	msg = "Accept file " + filepath.Base(path) + " (y/n/a)?"
	reply = ui.SetInput(msg)
	switch reply {
	case "a":
		knownDevices = append(knownDevices, device)

	case "y":
		break

	default:
		return "", dbus.MakeFailedError(errors.New("Cancelled"))
	}

SkipAuthentication:
	go func() {
		defer adapter.Lock.Release(1)

		ui.StartProgress(transferPath, transferProps, path)
		ui.ObexConn.RemoveSession(sessionPath)
	}()

	return path, nil
}

// Cancel is called when the OBEX agent request was cancelled.
func (o *ObexAgent) Cancel() *dbus.Error {
	return nil
}

// Release is called when the OBEX agent is unregistered.
func (o *ObexAgent) Release() *dbus.Error {
	return nil
}
