package ui

import (
	"errors"

	"github.com/darkhz/bluetuith/bluez"
)

// onClickFunc is a handler for the submenu options in a menu.
// Clicking an option in the submenu will trigger the specified function.
func onClickFunc(id string) func() bool {
	var clickFunc func()

	switch id {
	case "power":
		clickFunc = power

	case "scan":
		clickFunc = scan

	case "change":
		clickFunc = change

	case "quit":
		clickFunc = quit

	case "connect":
		clickFunc = connect

	case "pair":
		clickFunc = pair

	case "trust":
		clickFunc = trust

	case "remove":
		clickFunc = remove
	}

	return func() bool {
		go clickFunc()
		exitMenuList()

		return false
	}
}

func power() {
	var poweredText string

	adapterPath := BluezConn.GetCurrentAdapter().Path
	adapterID := bluez.GetAdapterID(adapterPath)

	props, err := BluezConn.GetAdapterProperties(adapterPath)
	if err != nil {
		ErrorMessage(err)
		return
	}

	powered, ok := props["Powered"].Value().(bool)
	if !ok {
		ErrorMessage(errors.New("Cannot get discovery state"))
		return
	}

	if err := BluezConn.Power(adapterPath, !powered); err != nil {
		ErrorMessage(errors.New("Cannot set adapter power state"))
		return
	}

	if powered {
		poweredText = "off"
	} else {
		poweredText = "on"
	}

	InfoMessage(adapterID+" is powered "+poweredText, false)

	setMenuItemToggle("adapter", "power", !powered)
}

// scan checks the current adapter's state and starts/stops discovery.
func scan() {
	adapterPath := BluezConn.GetCurrentAdapter().Path

	props, err := BluezConn.GetAdapterProperties(adapterPath)
	if err != nil {
		ErrorMessage(err)
		return
	}

	discover, ok := props["Discovering"].Value().(bool)
	if !ok {
		ErrorMessage(errors.New("Cannot get discovery state"))
		return
	}

	if !discover {
		if err := BluezConn.StartDiscovery(adapterPath); err != nil {
			ErrorMessage(err)
			return
		}
		InfoMessage("Scanning for devices...", true)
	} else {
		if err := BluezConn.StopDiscovery(adapterPath); err != nil {
			ErrorMessage(err)
			return
		}
		InfoMessage("Scanning stopped", false)
	}

	setMenuItemToggle("adapter", "scan", !discover)
}

// change launches a popup with the adapters list.
func change() {
	App.QueueUpdateDraw(func() {
		adapterChange()
	})
}

// quit stops the adapter's discovery mode, closes the bluez connection
// and exits the application.
func quit() {
	if !confirmQuit() {
		return
	}

	BluezConn.StopDiscovery(BluezConn.GetCurrentAdapterID())
	BluezConn.Close()

	StopUI()
}

// createConnect sets the oncreate handler for the connect submenu option.
func createConnect() bool {
	device := getDeviceFromSelection(false)
	if device == (bluez.Device{}) {
		return false
	}

	return device.Connected
}

// createTrust sets the oncreate handler for the trust submenu option.
func createTrust() bool {
	device := getDeviceFromSelection(false)
	if device == (bluez.Device{}) {
		return false
	}

	return device.Trusted
}

// connect retrieves the selected device, and toggles its connection state.
func connect() {
	device := getDeviceFromSelection(true)
	if device == (bluez.Device{}) {
		return
	}

	disconnectFunc := func() {
		if err := BluezConn.Disconnect(device.Path); err != nil {
			ErrorMessage(err)
			return
		}
	}

	connectFunc := func() {
		InfoMessage("Connecting to "+device.Name, true)
		if err := BluezConn.Connect(device.Path); err != nil {
			ErrorMessage(err)
			return
		}
		InfoMessage("Connected to "+device.Name, false)
	}

	if !device.Connected {
		startOperation(
			connectFunc,
			func() {
				disconnectFunc()
				InfoMessage("Cancelled connection to "+device.Name, false)
			},
		)
	} else {
		InfoMessage("Disconnecting from "+device.Name, true)
		disconnectFunc()
		InfoMessage("Disconnected from "+device.Name, false)
	}

	setMenuItemToggle("device", "connect", !device.Connected)
}

// pair retrieves the selected device, and attempts to pair with it.
func pair() {
	device := getDeviceFromSelection(true)
	if device == (bluez.Device{}) {
		return
	}
	if device.Paired {
		InfoMessage(device.Name+" is already paired", false)
		return
	}

	startOperation(
		func() {
			InfoMessage("Pairing with "+device.Name, true)
			if err := BluezConn.Pair(device.Path); err != nil {
				ErrorMessage(err)
				return
			}
			InfoMessage("Paired with "+device.Name, false)
		},
		func() {
			if err := BluezConn.CancelPairing(device.Path); err != nil {
				ErrorMessage(err)
				return
			}
			InfoMessage("Cancelled pairing with "+device.Name, false)
		},
	)
}

// trust retrieves the selected device, and toggles its trust property.
func trust() {
	device := getDeviceFromSelection(true)
	if device == (bluez.Device{}) {
		return
	}

	if err := BluezConn.SetDeviceProperty(device.Path, "Trusted", !device.Trusted); err != nil {
		ErrorMessage(errors.New("Cannot set trusted property for " + device.Name))
		return
	}

	setMenuItemToggle("device", "trust", !device.Trusted)
}

// remove retrieves the selected device, and removes it from the adapter.
func remove() {
	device := getDeviceFromSelection(true)
	if device == (bluez.Device{}) {
		return
	}

	if txt := SetInput("Remove " + device.Name + " (y/n)?"); txt != "y" {
		return
	}

	if err := BluezConn.RemoveDevice(device.Path); err != nil {
		ErrorMessage(err)
		return
	}

	InfoMessage("Removed "+device.Name, false)
}
