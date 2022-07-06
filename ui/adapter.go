package ui

import (
	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/godbus/dbus/v5"
)

// adapterChange launches a popup with a list of adapters.
// Changing the selection will change the currently selected adapter.
func adapterChange() {
	var adapterMenu *tview.Table
	var x, y, maxWidth, currentIndex, selCount int

	for _, region := range MenuBar.GetRegionInfos() {
		if region.ID == "adapter" {
			x, y = region.FromX, region.FromY+2
			break
		}
	}

	exitMenuList()

	adapterMenu = tview.NewTable()
	adapterMenu.SetBorder(true)
	adapterMenu.SetSelectable(true, false)
	adapterMenu.SetBackgroundColor(tcell.ColorDefault)
	adapterMenu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter, tcell.KeyEscape:
			Pages.RemovePage("adaptermenu")
			App.SetFocus(DeviceTable)
		}

		return event
	})
	adapterMenu.SetSelectionChangedFunc(func(row, col int) {
		if selCount == 0 {
			selCount++
			return
		}

		cell := adapterMenu.GetCell(row, 0)
		if cell == nil {
			return
		}

		adapter, ok := cell.GetReference().(bluez.Adapter)
		if !ok {
			return
		}

		BluezConn.StopDiscovery(BluezConn.GetCurrentAdapterID())
		InfoMessage("", false)

		BluezConn.SetCurrentAdapter(adapter)

		listDevices()
	})

	for row, adapter := range BluezConn.GetAdapters() {
		if len(adapter.Name) > maxWidth {
			maxWidth = len(adapter.Name)
		}

		if adapter == BluezConn.GetCurrentAdapter() {
			currentIndex = row
		}

		adapterMenu.SetCell(row, 0, tview.NewTableCell(adapter.Name).
			SetExpansion(1).
			SetReference(adapter).
			SetAlign(tview.AlignLeft),
		)
		adapterMenu.SetCell(row, 1, tview.NewTableCell("("+bluez.GetAdapterID(adapter.Path)+")").
			SetAlign(tview.AlignRight),
		)
	}

	adapterMenu.Select(currentIndex, 0)

	Pages.AddAndSwitchToPage(
		"adaptermenu",
		drawMenuBox(adapterMenu, adapterMenu.GetRowCount()+2, maxWidth+20, x, y),
		true,
	).ShowPage("main")
	App.SetFocus(adapterMenu)
}

// adapterEvent handles adapter-specific events.
func adapterEvent(signal *dbus.Signal, signalData interface{}) {
	switch signal.Name {
	case "org.freedesktop.DBus.ObjectManager.InterfacesRemoved":
		adapterPath, ok := signalData.(string)
		if !ok {
			return
		}

		if adapterPath == BluezConn.GetCurrentAdapter().Path {
			BluezConn.SetCurrentAdapter()
			listDevices()
		}

		fallthrough

	case "org.freedesktop.DBus.ObjectManager.InterfacesAdded":
		App.QueueUpdateDraw(func() {
			if Pages.HasPage("adaptermenu") {
				Pages.RemovePage("adaptermenu")
				adapterChange()
			}
		})
	}
}
