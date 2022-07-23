package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/godbus/dbus/v5"
)

var (
	// DeviceTable holds the device list.
	DeviceTable *tview.Table

	// BluezConn holds the current bluez DBus connection.
	BluezConn *bluez.Bluez
)

// deviceTable sets up and returns the DeviceTable.
func deviceTable() *tview.Table {
	DeviceTable = tview.NewTable()
	DeviceTable.SetSelectorWrap(true)
	DeviceTable.SetSelectable(true, false)
	DeviceTable.SetBackgroundColor(tcell.ColorDefault)
	DeviceTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlM:
			MenuBar.Highlight("adapter")
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

	headerText := fmt.Sprintf("[::bu]%s (%s)",
		BluezConn.GetCurrentAdapter().Name,
		BluezConn.GetCurrentAdapterID(),
	)
	setMenuBarHeader(headerText)

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

	name := device.Name + " (" + device.Type + ")"
	propColor := "[grey::b]"

	if device.Connected {
		propColor = "[green::b]"
		props += "Connected"

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
		props = strings.TrimRight(props, ", ")
		props = propColor + "(" + props + ")"
	} else {
		props = "[orange][+]"
	}

	DeviceTable.SetCell(
		row, 0, tview.NewTableCell(name).
			SetExpansion(1).
			SetReference(device).
			SetAlign(tview.AlignLeft),
	)
	DeviceTable.SetCell(
		row, 1, tview.NewTableCell(props).
			SetExpansion(1).
			SetAlign(tview.AlignRight).
			SetSelectedStyle(tcell.Style{}.
				Attributes(tcell.AttrBold),
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
