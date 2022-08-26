package ui

import (
	"errors"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
)

//gocyclo:ignore
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

	case "progress":
		clickFunc = progress

	case "quit":
		clickFunc = quit

	case "connect":
		clickFunc = connect

	case "pair":
		clickFunc = pair

	case "trust":
		clickFunc = trust

	case "send":
		clickFunc = send

	case "network":
		clickFunc = networkAP

	case "profiles":
		clickFunc = profiles

	case "showplayer":
		clickFunc = showplayer

	case "hideplayer":
		clickFunc = hideplayer

	case "info":
		clickFunc = info

	case "remove":
		clickFunc = remove
	}

	return func() bool {
		go clickFunc()
		exitMenu("menulist")

		return false
	}
}

// power checks and toggles the adapter's powered state.
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
		ErrorMessage(errors.New("Cannot get powered state"))
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

// progress displays the progress view.
func progress() {
	App.QueueUpdateDraw(func() {
		progressView(true)
	})
}

// quit stops discovery mode for all existing adapters, closes the bluez connection
// and exits the application.
func quit() {
	if !confirmQuit() {
		return
	}

	for _, adapter := range BluezConn.GetAdapters() {
		BluezConn.StopDiscovery(adapter.Path)
	}

	BluezConn.Close()

	StopUI()
}

// createPower sets the oncreate handler for the power submenu option.
func createPower() bool {
	adapterPath := BluezConn.GetCurrentAdapter().Path

	props, err := BluezConn.GetAdapterProperties(adapterPath)
	if err != nil {
		return false
	}

	powered, ok := props["Powered"].Value().(bool)
	if !ok {
		return false
	}

	return powered
}

// createConnect sets the oncreate handler for the connect submenu option.
func createConnect() bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return device.Connected
}

// createTrust sets the oncreate handler for the trust submenu option.
func createTrust() bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return device.Trusted
}

// visibleSend sets the visible handler for the send submenu option.
func visibleSend() bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return cmd.GetConfigProperty("obex") == "true" &&
		device.HaveService(bluez.OBEX_OBJPUSH_SVCLASS_ID)
}

// visibleNetwork sets the visible handler for the network submenu option.
func visibleNetwork() bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return cmd.GetConfigProperty("network") == "true" &&
		device.HaveService(bluez.NAP_SVCLASS_ID) &&
		(device.HaveService(bluez.PANU_SVCLASS_ID) ||
			device.HaveService(bluez.DIALUP_NET_SVCLASS_ID))
}

// visibleProfile sets the visible handler for the audio profiles submenu option.
func visibleProfile() bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return device.HaveService(bluez.AUDIO_SOURCE_SVCLASS_ID) ||
		device.HaveService(bluez.AUDIO_SINK_SVCLASS_ID)
}

// visiblePlayer sets the visible handler for the media player submenu option.
func visiblePlayer() bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return device.HaveService(bluez.AUDIO_SOURCE_SVCLASS_ID) &&
		device.HaveService(bluez.AV_REMOTE_SVCLASS_ID) &&
		device.HaveService(bluez.AV_REMOTE_TARGET_SVCLASS_ID)
}

// connect retrieves the selected device, and toggles its connection state.
func connect() {
	device := getDeviceFromSelection(true)
	if device.Path == "" {
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
	if device.Path == "" {
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
	if device.Path == "" {
		return
	}

	if err := BluezConn.SetDeviceProperty(device.Path, "Trusted", !device.Trusted); err != nil {
		ErrorMessage(errors.New("Cannot set trusted property for " + device.Name))
		return
	}

	setMenuItemToggle("device", "trust", !device.Trusted)
}

// send gets a file list from the file picker and sends all selected files
// to the target device.
func send() {
	adapter := BluezConn.GetCurrentAdapter()
	if !adapter.Lock.TryAcquire(1) {
		return
	}
	defer adapter.Lock.Release(1)

	device := getDeviceFromSelection(true)
	if !device.Paired || !device.Connected {
		ErrorMessage(errors.New(device.Name + " is not paired and/or connected"))
		return
	}

	InfoMessage("Initializing OBEX session..", true)

	sessionPath, err := ObexConn.CreateSession(device.Address)
	if err != nil {
		ErrorMessage(err)
		return
	}

	InfoMessage("Created OBEX session", false)

	for _, file := range filePicker() {
		transferPath, transferProps, err := ObexConn.SendFile(sessionPath, file)
		if err != nil {
			ErrorMessage(err)
			continue
		}

		if !StartProgress(transferPath, transferProps) {
			break
		}
	}

	ObexConn.RemoveSession(sessionPath)
}

// networkAP launches a popup with the available networks.
func networkAP() {
	App.QueueUpdateDraw(func() {
		networkSelect()
	})
}

// profiles launches a popup with the available audio profiles.
func profiles() {
	App.QueueUpdateDraw(func() {
		audioProfiles()
	})
}

// showplayer starts the media player.
func showplayer() {
	StartMediaPlayer()
}

// hideplayer hides the media player.
func hideplayer() {
	StopMediaPlayer()
}

// info retreives the selected device, and shows the device information.
func info() {
	App.QueueUpdateDraw(func() {
		getDeviceInfo()
	})
}

// remove retrieves the selected device, and removes it from the adapter.
func remove() {
	device := getDeviceFromSelection(true)
	if device.Path == "" {
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
