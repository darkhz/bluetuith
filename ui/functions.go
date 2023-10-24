package ui

import (
	"context"
	"errors"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
)

// FunctionContext describes the context in which the
// function is supoosed to be executed in.
type FunctionContext string

// The different context types for functions.
const (
	FunctionClick   FunctionContext = "Click"
	FunctionCreate  FunctionContext = "Create"
	FunctionVisible FunctionContext = "Visible"
)

var functions = map[FunctionContext]map[cmd.Key]func(set ...string) bool{
	FunctionClick: {
		cmd.KeyAdapterTogglePower:        power,
		cmd.KeyAdapterToggleDiscoverable: discoverable,
		cmd.KeyAdapterTogglePairable:     pairable,
		cmd.KeyAdapterToggleScan:         scan,
		cmd.KeyAdapterChange:             change,
		cmd.KeyDeviceConnect:             connect,
		cmd.KeyDevicePair:                pair,
		cmd.KeyDeviceTrust:               trust,
		cmd.KeyDeviceSendFiles:           send,
		cmd.KeyDeviceNetwork:             networkAP,
		cmd.KeyDeviceAudioProfiles:       profiles,
		cmd.KeyPlayerShow:                showplayer,
		cmd.KeyDeviceInfo:                info,
		cmd.KeyDeviceRemove:              remove,
		cmd.KeyProgressView:              progress,
		cmd.KeyPlayerHide:                hideplayer,
		cmd.KeyQuit:                      quit,
	},
	FunctionCreate: {
		cmd.KeyAdapterTogglePower:        createPower,
		cmd.KeyAdapterToggleDiscoverable: createDiscoverable,
		cmd.KeyAdapterTogglePairable:     createPairable,
		cmd.KeyDeviceConnect:             createConnect,
		cmd.KeyDeviceTrust:               createTrust,
	},
	FunctionVisible: {
		cmd.KeyDeviceSendFiles:     visibleSend,
		cmd.KeyDeviceNetwork:       visibleNetwork,
		cmd.KeyDeviceAudioProfiles: visibleProfile,
		cmd.KeyPlayerShow:          visiblePlayer,
	},
}

// KeyHandler executes the handler assigned to the key type based on
// the function context.
func KeyHandler(key cmd.Key, context FunctionContext) func() bool {
	handler := functions[context][key]

	if context == FunctionClick {
		return func() bool {
			go handler()
			exitMenu()

			return false
		}
	}

	return func() bool {
		return handler()
	}
}

// power checks and toggles the adapter's powered state.
func power(set ...string) bool {
	var poweredText string

	adapterPath := UI.Bluez.GetCurrentAdapter().Path
	adapterID := bluez.GetAdapterID(adapterPath)

	props, err := UI.Bluez.GetAdapterProperties(adapterPath)
	if err != nil {
		ErrorMessage(err)
		return false
	}

	powered, ok := props["Powered"].Value().(bool)
	if !ok {
		ErrorMessage(errors.New("Cannot get powered state"))
		return false
	}

	if set != nil {
		state := set[0] == "yes"
		if state == powered {
			return false
		}

		powered = !state
	}

	if err := UI.Bluez.Power(adapterPath, !powered); err != nil {
		ErrorMessage(errors.New("Cannot set adapter power state"))
		return false
	}

	if powered {
		poweredText = "off"
	} else {
		poweredText = "on"
	}

	InfoMessage(adapterID+" is powered "+poweredText, false)

	setMenuItemToggle("adapter", cmd.KeyAdapterTogglePower, !powered)

	return true
}

// discoverable checks and toggles the adapter's discoverable state.
func discoverable(set ...string) bool {
	var discoverableText string

	adapterPath := UI.Bluez.GetCurrentAdapter().Path
	adapterID := bluez.GetAdapterID(adapterPath)

	props, err := UI.Bluez.GetAdapterProperties(adapterPath)
	if err != nil {
		ErrorMessage(err)
		return false
	}

	discoverable, ok := props["Discoverable"].Value().(bool)
	if !ok {
		ErrorMessage(errors.New("Cannot get discoverable state"))
		return false
	}

	if set != nil {
		state := set[0] == "yes"
		if state == discoverable {
			return false
		}

		discoverable = !state
	}

	if err := UI.Bluez.SetAdapterProperty(adapterPath, "Discoverable", !discoverable); err != nil {
		ErrorMessage(err)
		return false
	}

	if !discoverable {
		discoverableText = "discoverable"
	} else {
		discoverableText = "not discoverable"
	}

	InfoMessage(adapterID+" is "+discoverableText, false)

	setMenuItemToggle("adapter", cmd.KeyAdapterToggleDiscoverable, !discoverable)

	return true
}

// pairable checks and toggles the adapter's pairable state.
func pairable(set ...string) bool {
	var pairableText string

	adapterPath := UI.Bluez.GetCurrentAdapter().Path
	adapterID := bluez.GetAdapterID(adapterPath)

	props, err := UI.Bluez.GetAdapterProperties(adapterPath)
	if err != nil {
		ErrorMessage(err)
		return false
	}

	pairable, ok := props["Pairable"].Value().(bool)
	if !ok {
		ErrorMessage(errors.New("Cannot get pairable state"))
		return false
	}

	if set != nil {
		state := set[0] == "yes"
		if state == pairable {
			return false
		}

		pairable = !state
	}

	if err := UI.Bluez.SetAdapterProperty(adapterPath, "Pairable", !pairable); err != nil {
		ErrorMessage(err)
		return false
	}

	if !pairable {
		pairableText = "pairable"
	} else {
		pairableText = "not pairable"
	}

	InfoMessage(adapterID+" is "+pairableText, false)

	setMenuItemToggle("adapter", cmd.KeyAdapterTogglePairable, !pairable)

	return true
}

// scan checks the current adapter's state and starts/stops discovery.
func scan(set ...string) bool {
	adapterPath := UI.Bluez.GetCurrentAdapter().Path

	props, err := UI.Bluez.GetAdapterProperties(adapterPath)
	if err != nil {
		ErrorMessage(err)
		return false
	}

	discover, ok := props["Discovering"].Value().(bool)
	if !ok {
		ErrorMessage(errors.New("Cannot get discovery state"))
		return false
	}

	if set != nil {
		state := set[0] == "yes"
		if state == discover {
			return false
		}

		discover = !state
	}

	if !discover {
		if err := UI.Bluez.StartDiscovery(adapterPath); err != nil {
			ErrorMessage(err)
			return false
		}
		InfoMessage("Scanning for devices...", true)
	} else {
		if err := UI.Bluez.StopDiscovery(adapterPath); err != nil {
			ErrorMessage(err)
			return false
		}
		InfoMessage("Scanning stopped", false)
	}

	setMenuItemToggle("adapter", cmd.KeyAdapterToggleScan, !discover)

	return true
}

// change launches a popup with the adapters list.
func change(set ...string) bool {
	UI.QueueUpdateDraw(func() {
		adapterChange()
	})

	return true
}

// progress displays the progress view.
func progress(set ...string) bool {
	UI.QueueUpdateDraw(func() {
		progressView(true)
	})

	return true
}

// quit stops discovery mode for all existing adapters, closes the bluez connection
// and exits the application.
func quit(set ...string) bool {
	if cmd.IsPropertyEnabled("confirm-on-quit") && !confirmQuit() {
		return false
	}

	for _, adapter := range UI.Bluez.GetAdapters() {
		UI.Bluez.StopDiscovery(adapter.Path)
	}

	UI.Bluez.Close()

	StopUI()

	return true
}

// createPower sets the oncreate handler for the power submenu option.
func createPower(set ...string) bool {
	adapterPath := UI.Bluez.GetCurrentAdapter().Path

	props, err := UI.Bluez.GetAdapterProperties(adapterPath)
	if err != nil {
		return false
	}

	powered, ok := props["Powered"].Value().(bool)
	if !ok {
		return false
	}

	return powered
}

// createDiscoverable sets the oncreate handler for the discoverable submenu option
func createDiscoverable(set ...string) bool {
	adapterPath := UI.Bluez.GetCurrentAdapter().Path

	props, err := UI.Bluez.GetAdapterProperties(adapterPath)
	if err != nil {
		return false
	}

	discoverable, ok := props["Discoverable"].Value().(bool)
	if !ok {
		return false
	}

	return discoverable
}

// createPairable sets the oncreate handler for the pairable submenu option.
func createPairable(set ...string) bool {
	adapterPath := UI.Bluez.GetCurrentAdapter().Path

	props, err := UI.Bluez.GetAdapterProperties(adapterPath)
	if err != nil {
		return false
	}

	pairable, ok := props["Pairable"].Value().(bool)
	if !ok {
		return false
	}

	return pairable
}

// createConnect sets the oncreate handler for the connect submenu option.
func createConnect(set ...string) bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return device.Connected
}

// createTrust sets the oncreate handler for the trust submenu option.
func createTrust(set ...string) bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return device.Trusted
}

// visibleSend sets the visible handler for the send submenu option.
func visibleSend(set ...string) bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return cmd.IsPropertyEnabled("obex") &&
		device.HaveService(bluez.OBEX_OBJPUSH_SVCLASS_ID)
}

// visibleNetwork sets the visible handler for the network submenu option.
func visibleNetwork(set ...string) bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return cmd.IsPropertyEnabled("network") &&
		device.HaveService(bluez.NAP_SVCLASS_ID) &&
		(device.HaveService(bluez.PANU_SVCLASS_ID) ||
			device.HaveService(bluez.DIALUP_NET_SVCLASS_ID))
}

// visibleProfile sets the visible handler for the audio profiles submenu option.
func visibleProfile(set ...string) bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return device.HaveService(bluez.AUDIO_SOURCE_SVCLASS_ID) ||
		device.HaveService(bluez.AUDIO_SINK_SVCLASS_ID)
}

// visiblePlayer sets the visible handler for the media player submenu option.
func visiblePlayer(set ...string) bool {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return false
	}

	return device.HaveService(bluez.AUDIO_SOURCE_SVCLASS_ID) &&
		device.HaveService(bluez.AV_REMOTE_SVCLASS_ID) &&
		device.HaveService(bluez.AV_REMOTE_TARGET_SVCLASS_ID)
}

// connect retrieves the selected device, and toggles its connection state.
func connect(set ...string) bool {
	var device bluez.Device

	if set != nil {
		for _, d := range UI.Bluez.GetDevices() {
			if d.Address == set[0] {
				device = d
				break
			}
		}
	} else {
		device = getDeviceFromSelection(true)
		if device.Path == "" {
			return false
		}
	}

	disconnectFunc := func() {
		if err := UI.Bluez.Disconnect(device.Path); err != nil {
			ErrorMessage(err)
			return
		}
	}

	connectFunc := func() {
		InfoMessage("Connecting to "+device.Name, true)
		if err := UI.Bluez.Connect(device.Path); err != nil {
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

	setMenuItemToggle("device", cmd.KeyDeviceConnect, !device.Connected)

	return true
}

// pair retrieves the selected device, and attempts to pair with it.
func pair(set ...string) bool {
	device := getDeviceFromSelection(true)
	if device.Path == "" {
		return false
	}
	if device.Paired {
		InfoMessage(device.Name+" is already paired", false)
		return false
	}

	startOperation(
		func() {
			InfoMessage("Pairing with "+device.Name, true)
			if err := UI.Bluez.Pair(device.Path); err != nil {
				ErrorMessage(err)
				return
			}
			InfoMessage("Paired with "+device.Name, false)
		},
		func() {
			if err := UI.Bluez.CancelPairing(device.Path); err != nil {
				ErrorMessage(err)
				return
			}
			InfoMessage("Cancelled pairing with "+device.Name, false)
		},
	)

	return true
}

// trust retrieves the selected device, and toggles its trust property.
func trust(set ...string) bool {
	device := getDeviceFromSelection(true)
	if device.Path == "" {
		return false
	}

	if err := UI.Bluez.SetDeviceProperty(device.Path, "Trusted", !device.Trusted); err != nil {
		ErrorMessage(errors.New("Cannot set trusted property for " + device.Name))
		return false
	}

	setMenuItemToggle("device", cmd.KeyDeviceTrust, !device.Trusted)

	return true
}

// send gets a file list from the file picker and sends all selected files
// to the target device.
func send(set ...string) bool {
	adapter := UI.Bluez.GetCurrentAdapter()
	if !adapter.Lock.TryAcquire(1) {
		return false
	}
	defer adapter.Lock.Release(1)

	device := getDeviceFromSelection(true)
	if !device.Paired || !device.Connected {
		ErrorMessage(errors.New(device.Name + " is not paired and/or connected"))
		return false
	}

	ctx, cancel := context.WithCancel(context.Background())

	startOperation(
		func() {
			InfoMessage("Initializing OBEX session..", true)

			sessionPath, err := UI.Obex.CreateSession(ctx, device.Address)
			if err != nil {
				ErrorMessage(err)
				return
			}

			cancelOperation(false)

			InfoMessage("Created OBEX session", false)

			for _, file := range filePicker() {
				transferPath, transferProps, err := UI.Obex.SendFile(sessionPath, file)
				if err != nil {
					ErrorMessage(err)
					continue
				}

				if !StartProgress(transferPath, transferProps) {
					break
				}
			}

			UI.Obex.RemoveSession(sessionPath)
		},
		func() {
			cancel()
			InfoMessage("Cancelled OBEX session creation", false)
		},
	)

	return true
}

// networkAP launches a popup with the available networks.
func networkAP(set ...string) bool {
	UI.QueueUpdateDraw(func() {
		networkSelect()
	})

	return true
}

// profiles launches a popup with the available audio profiles.
func profiles(set ...string) bool {
	UI.QueueUpdateDraw(func() {
		audioProfiles()
	})

	return true
}

// showplayer starts the media player.
func showplayer(set ...string) bool {
	StartMediaPlayer()

	return true
}

// hideplayer hides the media player.
func hideplayer(set ...string) bool {
	StopMediaPlayer()

	return true
}

// info retreives the selected device, and shows the device information.
func info(set ...string) bool {
	UI.QueueUpdateDraw(func() {
		getDeviceInfo()
	})

	return true
}

// remove retrieves the selected device, and removes it from the adapter.
func remove(set ...string) bool {
	device := getDeviceFromSelection(true)
	if device.Path == "" {
		return false
	}

	if txt := SetInput("Remove " + device.Name + " (y/n)?"); txt != "y" {
		return false
	}

	if err := UI.Bluez.RemoveDevice(device.Path); err != nil {
		ErrorMessage(err)
		return false
	}

	InfoMessage("Removed "+device.Name, false)

	return true
}
