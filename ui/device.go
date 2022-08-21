package ui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/network"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/godbus/dbus/v5"
)

var (
	// DeviceTable holds the device list.
	DeviceTable *tview.Table

	// BluezConn holds the current bluez DBus connection.
	BluezConn *bluez.Bluez

	// ObexConn holds the current bluez obex DBus connection.
	ObexConn *bluez.Obex

	// NetworkConn holds the current network connection.
	NetworkConn *network.Network
)

// deviceTable sets up and returns the DeviceTable.
func deviceTable() *tview.Table {
	DeviceTable = tview.NewTable()
	DeviceTable.SetSelectorWrap(true)
	DeviceTable.SetSelectable(true, false)
	DeviceTable.SetBackgroundColor(theme.GetColor("Background"))
	DeviceTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlM:
			MenuBar.Highlight("adapter")
			return event
		}

		switch event.Rune() {
		case '?':
			showHelp()
			return event
		}

		menuListInputHandler(event)

		return event
	})
	DeviceTable.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if menuListMouseHandler(action, event) == nil {
			return action, nil
		}

		return action, event
	})

	return DeviceTable
}

// setupDevices initializes the bluez DBus interface, sets up
// a bluez event listener via watchEvent, and lists the devices.
func setupDevices() {
	listDevices()
	go watchEvent()
}

// listDevices lists the devices belonging to the selected adapter.
func listDevices() {
	if BluezConn == nil {
		return
	}

	headerText := fmt.Sprintf("%s (%s)",
		BluezConn.GetCurrentAdapter().Name,
		BluezConn.GetCurrentAdapterID(),
	)
	setMenuBarHeader(theme.ColorWrap("Adapter", headerText, "bu"))

	DeviceTable.Clear()
	for i, device := range BluezConn.GetDevices() {
		setDeviceTableInfo(i, device)
	}
	DeviceTable.Select(0, 0)
}

// checkDeviceTable iterates through the DeviceTable and checks
// if a device whose path matches the path parameter exists.
func checkDeviceTable(path string) (int, bool) {
	for row := 0; row < DeviceTable.GetRowCount(); row++ {
		cell := DeviceTable.GetCell(row, 0)
		if cell == nil {
			continue
		}

		ref, ok := cell.GetReference().(bluez.Device)
		if !ok {
			continue
		}

		if ref.Path == path {
			return row, true
		}
	}

	return -1, false
}

// getDeviceInfo shows information about a device.
func getDeviceInfo() {
	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return
	}

	yesno := func(val bool) string {
		if !val {
			return "no"
		}

		return "yes"
	}

	props := [][]string{
		{"Name", device.Name},
		{"Address", device.Address},
		{"Class", strconv.FormatUint(uint64(device.Class), 10)},
		{"Adapter", filepath.Base(device.Adapter)},
		{"Connected", yesno(device.Connected)},
		{"Paired", yesno(device.Paired)},
		{"Trusted", yesno(device.Trusted)},
		{"Blocked", yesno(device.Blocked)},
		{"LegacyPairing", yesno(device.LegacyPairing)},
	}
	if device.Modalias != "" {
		props = append(props, []string{"Modalias", device.Modalias})
	}
	props = append(props, []string{"UUIDs", ""})

	deviceInfoTable := tview.NewTable()
	deviceInfoTable.SetBorder(true)
	deviceInfoTable.SetSelectorWrap(true)
	deviceInfoTable.SetSelectable(true, false)
	deviceInfoTable.SetBorderColor(theme.GetColor("Border"))
	deviceInfoTable.SetBackgroundColor(theme.GetColor("Background"))
	deviceInfoTable.SetTitle(theme.ColorWrap("Text", "[ DEVICE INFO ]"))
	deviceInfoTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			Pages.RemovePage("infomodal")
		}

		return event
	})
	deviceInfoTable.SetSelectionChangedFunc(func(row, col int) {
		_, _, _, height := deviceInfoTable.GetRect()
		deviceInfoTable.SetOffset(row-((height-1)/2), 0)
	})

	for i, prop := range props {
		propName := prop[0]
		propValue := prop[1]

		switch propName {
		case "Address":
			propValue += " (" + device.AddressType + ")"

		case "Class":
			propValue += " (" + device.Type + ")"
		}

		deviceInfoTable.SetCell(i, 0, tview.NewTableCell("[::b]"+propName+":").
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor("Text")).
			SetSelectedStyle(tcell.Style{}.
				Bold(true).
				Underline(true),
			),
		)

		deviceInfoTable.SetCell(i, 1, tview.NewTableCell(propValue).
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor("Text")).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor("Text")).
				Background(theme.BackgroundColor("Text")),
			),
		)
	}

	rows := deviceInfoTable.GetRowCount() - 1
	for i, serviceUUID := range device.UUIDs {
		serviceType := bluez.ServiceType(serviceUUID)
		serviceUUID = "(" + serviceUUID + ")"

		deviceInfoTable.SetCell(rows+i, 1, tview.NewTableCell(serviceType).
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor("Text")).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor("Text")).
				Background(theme.BackgroundColor("Text")),
			),
		)

		deviceInfoTable.SetCell(rows+i, 2, tview.NewTableCell(serviceUUID).
			SetExpansion(0).
			SetTextColor(theme.GetColor("Text")).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor("Text")).
				Background(theme.BackgroundColor("Text")),
			),
		)
	}

	showModal("infomodal", deviceInfoTable, 60, 1)
}

// getDeviceFromSelection retrieves device information from
// the current selection in the DeviceTable.
func getDeviceFromSelection(lock bool) bluez.Device {
	var device bluez.Device

	getdevice := func() {
		row, _ := DeviceTable.GetSelection()

		cell := DeviceTable.GetCell(row, 0)
		if cell == nil {
			device = bluez.Device{}
		}

		d, ok := cell.GetReference().(bluez.Device)
		if !ok {
			device = bluez.Device{}
		}

		device = d
	}

	if lock {
		App.QueueUpdateDraw(func() {
			getdevice()
		})

		return device
	}

	getdevice()

	return device
}

// setDeviceTableInfo writes device information into the
// specified row of the DeviceTable.
func setDeviceTableInfo(row int, device bluez.Device) {
	var props string

	name := device.Name
	name += theme.ColorWrap("DeviceType", " ("+device.Type+")")

	nameColor := "Device"
	propColor := "DeviceProperty"

	if device.Connected {
		props += "Connected"

		nameColor = "DeviceConnected"
		propColor = "DevicePropertyConnected"

		if device.RSSI < 0 {
			rssi := strconv.FormatInt(int64(device.RSSI), 10)
			props += "[" + rssi + "[]"
		}

		props += ", "
	}
	if device.Trusted {
		props += "Trusted, "
	}
	if device.Paired {
		props += "Paired, "
	}
	if device.Blocked {
		props += "Blocked, "
	}

	if props != "" {
		props = "(" + strings.TrimRight(props, ", ") + ")"
	} else {
		props = "[New Device[]"
		nameColor = "DeviceDiscovered"
		propColor = "DevicePropertyDiscovered"
	}

	DeviceTable.SetCell(
		row, 0, tview.NewTableCell(name).
			SetExpansion(1).
			SetReference(device).
			SetAlign(tview.AlignLeft).
			SetAttributes(tcell.AttrBold).
			SetTextColor(theme.GetColor(nameColor)).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor(nameColor)).
				Background(theme.BackgroundColor(nameColor)),
			),
	)
	DeviceTable.SetCell(
		row, 1, tview.NewTableCell(props).
			SetExpansion(1).
			SetAlign(tview.AlignRight).
			SetTextColor(theme.GetColor(propColor)).
			SetSelectedStyle(tcell.Style{}.
				Bold(true),
			),
	)
}

// deviceEvent handles device-specific events.
func deviceEvent(signal *dbus.Signal, signalData interface{}) {
	switch signal.Name {
	case "org.freedesktop.DBus.Properties.PropertiesChanged":
		device, ok := signalData.(bluez.Device)
		if !ok {
			return
		}

		App.QueueUpdateDraw(func() {
			row, ok := checkDeviceTable(device.Path)
			if ok {
				setDeviceTableInfo(row, device)
			}
		})

	case "org.freedesktop.DBus.ObjectManager.InterfacesAdded":
		deviceMap, ok := signalData.(map[string][]bluez.Device)
		if !ok {
			return
		}

		for devicePath, devices := range deviceMap {
			for _, device := range devices {
				if device.Adapter != BluezConn.GetCurrentAdapter().Path {
					continue
				}

				App.QueueUpdateDraw(func() {
					_, ok := checkDeviceTable(devicePath)
					if !ok {
						setDeviceTableInfo(DeviceTable.GetRowCount(), device)
					}
				})
			}
		}

	case "org.freedesktop.DBus.ObjectManager.InterfacesRemoved":
		devicePath, ok := signalData.(string)
		if !ok {
			return
		}

		App.QueueUpdateDraw(func() {
			row, ok := checkDeviceTable(devicePath)
			if ok {
				DeviceTable.RemoveRow(row)
			}
		})
	}
}
