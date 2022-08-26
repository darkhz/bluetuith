package ui

import (
	"sort"
	"strings"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/godbus/dbus/v5"
)

// adapterChange launches a popup with a list of adapters.
// Changing the selection will change the currently selected adapter.
func adapterChange() {
	setSelectorMenu(
		"adapter", nil,
		func(adapterMenu *tview.Table, row, col int) {
			cell := adapterMenu.GetCell(row, 0)
			if cell == nil {
				return
			}

			adapter, ok := cell.GetReference().(bluez.Adapter)
			if !ok {
				return
			}

			if err := BluezConn.StopDiscovery(BluezConn.GetCurrentAdapter().Path); err == nil {
				setMenuItemToggle("adapter", "scan", false, struct{}{})
			}

			if strings.Contains(MessageBox.GetText(true), "Scanning for devices") {
				InfoMessage("Scanning stopped on "+BluezConn.GetCurrentAdapterID(), false)
			}

			BluezConn.SetCurrentAdapter(adapter)

			listDevices()
		},
		func(adapterMenu *tview.Table) (int, int) {
			var width, index int

			adapters := BluezConn.GetAdapters()
			sort.Slice(adapters, func(i, j int) bool {
				return adapters[i].Path < adapters[j].Path
			})

			for row, adapter := range adapters {
				if len(adapter.Name) > width {
					width = len(adapter.Name)
				}

				if adapter.Path == BluezConn.GetCurrentAdapter().Path {
					index = row
				}

				adapterMenu.SetCell(row, 0, tview.NewTableCell(adapter.Name).
					SetExpansion(1).
					SetReference(adapter).
					SetAlign(tview.AlignLeft).
					SetTextColor(theme.GetColor("Adapter")).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor("Adapter")).
						Background(theme.BackgroundColor("Adapter")),
					),
				)
				adapterMenu.SetCell(row, 1, tview.NewTableCell("("+bluez.GetAdapterID(adapter.Path)+")").
					SetAlign(tview.AlignRight).
					SetTextColor(theme.GetColor("Adapter")).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor("Adapter")).
						Background(theme.BackgroundColor("Adapter")),
					),
				)
			}

			return width, index
		})
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
