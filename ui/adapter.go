package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/godbus/dbus/v5"
)

// AdapterStatus describes the adapter status display.
type AdapterStatus struct {
	view *tview.TextView
}

var adapterStatus AdapterStatus

// adapterStatusView sets up and returns the adapter status display.
func adapterStatusView() *tview.TextView {
	adapterStatus.view = tview.NewTextView()
	adapterStatus.view.SetRegions(true)
	adapterStatus.view.SetDynamicColors(true)
	adapterStatus.view.SetTextAlign(tview.AlignRight)
	adapterStatus.view.SetBackgroundColor(theme.GetColor("MenuBar"))

	return adapterStatus.view
}

// adapterChange launches a popup with a list of adapters.
// Changing the selection will change the currently selected adapter.
func adapterChange() {
	setContextMenu(
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
				setMenuItemToggle("adapter", cmd.KeyAdapterToggleScan, false, struct{}{})
			}

			if strings.Contains(UI.Status.MessageBox.GetText(true), "Scanning for devices") {
				InfoMessage("Scanning stopped on "+UI.Bluez.GetCurrentAdapterID(), false)
			}

			UI.Bluez.SetCurrentAdapter(adapter)
			updateAdapterStatus(adapter)

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

// updateAdapterStatus updates the adapter status display.
func updateAdapterStatus(adapter bluez.Adapter) {
	var state string
	var regions []string

	properties := map[string]bool{
		"Powered":      false,
		"Discovering":  false,
		"Discoverable": false,
		"Pairable":     false,
	}

	props, err := UI.Bluez.GetAdapterProperties(adapter.Path)
	if err != nil {
		return
	}

	for name := range properties {
		enabled, ok := props[name].Value().(bool)
		if !ok {
			continue
		}

		properties[name] = enabled
	}

	for _, status := range []struct {
		Title   string
		Enabled bool
		Color   string
	}{
		{
			Title:   "Powered",
			Color:   "AdapterPowered",
			Enabled: properties["Powered"],
		},
		{
			Title:   "Scanning",
			Color:   "AdapterScanning",
			Enabled: properties["Discovering"],
		},
		{
			Title:   "Discoverable",
			Color:   "AdapterDiscoverable",
			Enabled: properties["Discoverable"],
		},
		{
			Title:   "Pairable",
			Color:   "AdapterPairable",
			Enabled: properties["Pairable"],
		},
	} {
		if !status.Enabled {
			if status.Title == "Powered" {
				status.Title = "Not " + status.Title
				status.Color = "AdapterNotPowered"
			} else {
				continue
			}
		}

		region := strings.ToLower(status.Title)

		textColor := "white"
		if IsColorBright(theme.GetColor(status.Color)) {
			textColor = "black"
		}

		state += theme.ColorWrap(
			status.Color,
			fmt.Sprintf("[\"%s\"] %s [\"\"]", region, status.Title), ":"+textColor+":b",
		)

		state += " "

		regions = append(regions, region)
	}

	adapterStatus.view.SetText(state)
	adapterStatus.view.Highlight(regions...)
}

// adapterEvent handles adapter-specific events.
func adapterEvent(signal *dbus.Signal, signalData interface{}) {
	switch signal.Name {
	case "org.freedesktop.DBus.Properties.PropertiesChanged":
		adapter, ok := signalData.(bluez.Adapter)
		if !ok {
			return
		}

		if adapter.Path != UI.Bluez.GetCurrentAdapter().Path {
			return
		}

		UI.QueueUpdateDraw(func() {
			updateAdapterStatus(adapter)
		})

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
			if modal, ok := ModalExists("adapter"); ok {
				modal.Exit(false)
				adapterChange()
			}
		})
	}
}
