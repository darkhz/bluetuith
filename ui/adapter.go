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

			if err := UI.Bluez.StopDiscovery(UI.Bluez.GetCurrentAdapter().Path); err == nil {
				setMenuItemToggle("adapter", "scan", false, struct{}{})
			}

			if strings.Contains(UI.Status.MessageBox.GetText(true), "Scanning for devices") {
				InfoMessage("Scanning stopped on "+UI.Bluez.GetCurrentAdapterID(), false)
			}

			UI.Bluez.SetCurrentAdapter(adapter)

			listDevices()
		},
		func(adapterMenu *tview.Table) (int, int) {
			var width, index int

			adapters := UI.Bluez.GetAdapters()
			sort.Slice(adapters, func(i, j int) bool {
				return adapters[i].Path < adapters[j].Path
			})

			for row, adapter := range adapters {
				if len(adapter.Name) > width {
					width = len(adapter.Name)
				}

				if adapter.Path == UI.Bluez.GetCurrentAdapter().Path {
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

		if adapterPath == UI.Bluez.GetCurrentAdapter().Path {
			UI.Bluez.SetCurrentAdapter()
			listDevices()
		}

		fallthrough

	case "org.freedesktop.DBus.ObjectManager.InterfacesAdded":
		UI.QueueUpdateDraw(func() {
			if UI.Pages.HasPage("adaptermenu") {
				UI.Pages.RemovePage("adaptermenu")
				adapterChange()
			}
		})
	}
}
