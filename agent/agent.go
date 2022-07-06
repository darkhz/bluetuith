package agent

import (
	"errors"
	"fmt"

	"github.com/darkhz/bluetuith/ui"
	"github.com/godbus/dbus/v5"
	bluezAgent "github.com/muka/go-bluetooth/bluez/profile/agent"
)

const (
	AgentBasePath        = "/org/bluez/agent/bluetuith"
	AgentPinCode         = "0000"
	AgentPassKey  uint32 = 1024
)

var (
	agent           *Agent
	alwaysAuthorize bool
)

// Agent implement interface Agent1Client
type Agent struct {
	path    dbus.ObjectPath
	pinCode string
	passKey uint32
}

// NewAgent return a Agent instance with default pincode and passcode
func NewAgent() *Agent {
	ag := &Agent{
		passKey: AgentPassKey,
		pinCode: AgentPinCode,
		path:    dbus.ObjectPath(AgentBasePath),
	}

	return ag
}

func SetupAgent(conn *dbus.Conn) error {
	agent = NewAgent()
	if err := bluezAgent.ExposeAgent(conn, agent, bluezAgent.CapKeyboardDisplay, true); err != nil {
		return err
	}

	return nil
}

func RemoveAgent() error {
	return bluezAgent.RemoveAgent(agent)
}

func (a *Agent) SetPassKey(passkey uint32) {
	a.passKey = passkey
}

func (a *Agent) SetPassCode(pinCode string) {
	a.pinCode = pinCode
}

func (a *Agent) PassKey() uint32 {
	return a.passKey
}

func (a *Agent) PassCode() string {
	return a.pinCode
}

func (a *Agent) Path() dbus.ObjectPath {
	return a.path
}

func (a *Agent) Interface() string {
	return "org.bluez.Agent1"
}

func (a *Agent) RequestPinCode(path dbus.ObjectPath) (string, *dbus.Error) {
	return a.pinCode, nil
}

func (a *Agent) RequestPasskey(path dbus.ObjectPath) (uint32, *dbus.Error) {
	return a.passKey, nil
}

func (a *Agent) DisplayPinCode(path dbus.ObjectPath, pincode string) *dbus.Error {
	device, err := ui.GetDeviceFromPath(string(path))
	if err != nil {
		return dbus.MakeFailedError(err)
	}

	msg := fmt.Sprintf("Pincode for %s is %s", device.Name, pincode)
	ui.InfoMessage(msg, false)

	return nil
}

func (a *Agent) DisplayPasskey(path dbus.ObjectPath, passkey uint32, entered uint16) *dbus.Error {
	device, err := ui.GetDeviceFromPath(string(path))
	if err != nil {
		return dbus.MakeFailedError(err)
	}

	msg := fmt.Sprintf("Passkey for %s is %d, entered %d", device.Name, passkey, entered)
	ui.InfoMessage(msg, false)

	return nil
}

func (a *Agent) RequestConfirmation(path dbus.ObjectPath, passkey uint32) *dbus.Error {
	msg := fmt.Sprintf("Confirm passkey %d (y/n)", passkey)

	reply := ui.SetInput(msg)
	if reply != "y" {
		return dbus.MakeFailedError(errors.New("Cancelled"))
	}

	err := ui.SetTrusted(string(path), true)
	if err != nil {
		return dbus.MakeFailedError(err)
	}

	return nil
}

func (a *Agent) RequestAuthorization(path dbus.ObjectPath) *dbus.Error {
	msg := "Confirm pairing (y/n)"

	reply := ui.SetInput(msg)
	if reply != "y" {
		return dbus.MakeFailedError(errors.New("Cancelled"))
	}

	err := ui.SetTrusted(string(path), true)
	if err != nil {
		return dbus.MakeFailedError(err)
	}
	return nil
}

func (a *Agent) AuthorizeService(device dbus.ObjectPath, uuid string) *dbus.Error {
	if alwaysAuthorize {
		return nil
	}

	msg := fmt.Sprintf("Authorize service %s (y/n/a)", uuid)

	reply := ui.SetInput(msg)
	switch reply {
	case "a":
		alwaysAuthorize = true
		fallthrough

	case "y":
		return nil
	}

	return dbus.MakeFailedError(errors.New("Cancelled"))
}

func (a *Agent) Cancel() *dbus.Error {
	return nil
}

func (a *Agent) Release() *dbus.Error {
	return dbus.MakeFailedError(bluezAgent.RemoveAgent(a))
}
