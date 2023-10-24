package ui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/godbus/dbus/v5"
)

var DeviceTable *tview.Table

// deviceTable sets up and returns the DeviceTable.
func deviceTable() *tview.Table {
	DeviceTable = tview.NewTable()
	DeviceTable.SetSelectorWrap(true)
	DeviceTable.SetSelectable(true, false)
	DeviceTable.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	DeviceTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch cmd.KeyOperation(event) {
		case cmd.KeyMenu:
			menu.bar.Highlight("adapter")
			return event

		case cmd.KeyHelp:
			showHelp()
			return event
		}

		playerEvents(event, false)

		menuInputHandler(event)

		return ignoreDefaultEvent(event)
	})
	DeviceTable.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseRightClick {
			device := getDeviceFromSelection(false)
			if device.Path == "" {
				return action, event
			}

			setMenu(0, 0, "device", struct{}{})
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
	if UI.Bluez == nil {
		return
	}

	headerText := fmt.Sprintf("[\"adapterchange\"]%s (%s)[\"\"]",
		UI.Bluez.GetCurrentAdapter().Name,
		UI.Bluez.GetCurrentAdapterID(),
	)
	setMenuBarHeader(theme.ColorWrap(theme.ThemeAdapter, headerText, "::bu"))

	DeviceTable.Clear()
	for i, device := range UI.Bluez.GetDevices() {
		setDeviceTableInfo(i, device)
	}
	DeviceTable.Select(0, 0)
}

// connectDeviceByAddress connects to a device based on the provided address
// which was parsed from the "connect-bdaddr" command-line option.
func connectDeviceByAddress() {
	address := cmd.GetProperty("connect-bdaddr")
	if address == "" || UI.Bluez == nil {
		return
	}

	go connect(address)
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
		{"Bonded", yesno(device.Bonded)},
		{"Trusted", yesno(device.Trusted)},
		{"Blocked", yesno(device.Blocked)},
		{"LegacyPairing", yesno(device.LegacyPairing)},
	}
	if device.Modalias != "" {
		props = append(props, []string{"Modalias", device.Modalias})
	}
	props = append(props, []string{"UUIDs", ""})

	infoModal := NewModal("info", "Device Information", nil, 40, 100)
	infoModal.Table.SetSelectionChangedFunc(func(row, col int) {
		_, _, _, height := infoModal.Table.GetRect()
		infoModal.Table.SetOffset(row-((height-1)/2), 0)
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

		infoModal.Table.SetCell(i, 0, tview.NewTableCell("[::b]"+propName+":").
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeText)).
			SetSelectedStyle(tcell.Style{}.
				Bold(true).
				Underline(true),
			),
		)

		infoModal.Table.SetCell(i, 1, tview.NewTableCell(propValue).
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)
	}

	rows := infoModal.Table.GetRowCount() - 1
	for i, serviceUUID := range device.UUIDs {
		serviceType := bluez.ServiceType(serviceUUID)
		serviceUUID = "(" + serviceUUID + ")"

		infoModal.Table.SetCell(rows+i, 1, tview.NewTableCell(serviceType).
			SetExpansion(1).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)

		infoModal.Table.SetCell(rows+i, 2, tview.NewTableCell(serviceUUID).
			SetExpansion(0).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)
	}

	infoModal.Height = infoModal.Table.GetRowCount() + 4
	if infoModal.Height > 60 {
		infoModal.Height = 60
	}

	infoModal.Show()
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
		UI.QueueUpdateDraw(func() {
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
	if name == "" {
		name = device.Address
	}
	name += theme.ColorWrap(theme.ThemeDeviceType, " ("+device.Type+")")

	nameColor := theme.ThemeDevice
	propColor := theme.ThemeDeviceProperty

	if device.Connected {
		props += "Connected"

		nameColor = theme.ThemeDeviceConnected
		propColor = theme.ThemeDevicePropertyConnected

		if device.RSSI < 0 {
			rssi := strconv.FormatInt(int64(device.RSSI), 10)
			props += "[" + rssi + "[]"
		}

		if device.Percentage > 0 {
			props += ", Battery " + strconv.Itoa(device.Percentage) + "%"
		}

		props += ", "
	}

	if device.Trusted {
		props += "Trusted, "
	}
	if device.Blocked {
		props += "Blocked, "
	}
	if device.Bonded && device.Paired {
		props += "Bonded, "
	} else if !device.Bonded && device.Paired {
		props += "Paired, "
	}

	if props != "" {
		props = "(" + strings.TrimRight(props, ", ") + ")"
	} else {
		props = "[New Device[]"
		nameColor = theme.ThemeDeviceDiscovered
		propColor = theme.ThemeDevicePropertyDiscovered
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

		UI.QueueUpdateDraw(func() {
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
				if device.Adapter != UI.Bluez.GetCurrentAdapter().Path {
					continue
				}

				UI.QueueUpdateDraw(func() {
					deviceRow := DeviceTable.GetRowCount()

					row, ok := checkDeviceTable(devicePath)
					if ok {
						deviceRow = row
					}

					setDeviceTableInfo(deviceRow, device)
				})
			}
		}

	case "org.freedesktop.DBus.ObjectManager.InterfacesRemoved":
		devicePath, ok := signalData.(string)
		if !ok {
			return
		}

		UI.QueueUpdateDraw(func() {
			row, ok := checkDeviceTable(devicePath)
			if ok {
				DeviceTable.RemoveRow(row)
			}
		})
	}
}
