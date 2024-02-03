package ui

import (
	"fmt"
	"sort"
	"strings"
	"sync"

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
	adapterStatus.view.SetBackgroundColor(theme.GetColor(theme.ThemeMenuBar))

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

			cancelOperation(true)
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
					SetTextColor(theme.GetColor(theme.ThemeAdapter)).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor(theme.ThemeAdapter)).
						Background(theme.BackgroundColor(theme.ThemeAdapter)),
					),
				)
				adapterMenu.SetCell(row, 1, tview.NewTableCell("("+bluez.GetAdapterID(adapter.Path)+")").
					SetAlign(tview.AlignRight).
					SetTextColor(theme.GetColor(theme.ThemeAdapter)).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor(theme.ThemeAdapter)).
						Background(theme.BackgroundColor(theme.ThemeAdapter)),
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
		Color   theme.ThemeContext
	}{
		{
			Title:   "Powered",
			Enabled: properties["Powered"],
			Color:   theme.ThemeAdapterPowered,
		},
		{
			Title:   "Scanning",
			Enabled: properties["Discovering"],
			Color:   theme.ThemeAdapterScanning,
		},
		{
			Title:   "Discoverable",
			Enabled: properties["Discoverable"],
			Color:   theme.ThemeAdapterDiscoverable,
		},
		{
			Title:   "Pairable",
			Enabled: properties["Pairable"],
			Color:   theme.ThemeAdapterPairable,
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

		textColor := theme.ColorName(theme.BackgroundColor(status.Color))
		bgColor := theme.ThemeConfig[status.Color]

		region := strings.ToLower(status.Title)
		state += fmt.Sprintf("[\"%s\"][%s:%s:b] %s [-:-:-][\"\"] ", region, textColor, bgColor, status.Title)

		regions = append(regions, region)
	}

	adapterStatus.view.SetText(state)
}

// setAdapterStates sets the adapter states which were parsed from
// the "adapter-states" command-line option.
func setAdapterStates() {
	var lock sync.Mutex

	properties := cmd.GetPropertyMap("adapter-states")
	if len(properties) == 0 {
		return
	}

	seq, ok := properties["sequence"]
	if !ok {
		InfoMessage("Cannot get adapter states", false)
		return
	}

	sequence := strings.Split(seq, ",")
	for _, property := range sequence {
		var handler func(set ...string) bool

		state, ok := properties[property]
		if !ok {
			InfoMessage("Cannot set adapter "+state+" state", false)
			return
		}

		switch property {
		case "powered":
			handler = power

		case "scan":
			handler = scan

		case "discoverable":
			handler = discoverable

		case "pairable":
			handler = pairable

		default:
			continue
		}

		go func() {
			lock.Lock()
			defer lock.Unlock()

			handler(state)
		}()
	}
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
