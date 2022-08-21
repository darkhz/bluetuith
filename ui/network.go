package ui

import (
	"fmt"
	"strings"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// networkSelect shows a popup to select the network type.
func networkSelect() {
	var connTypes [][]string

	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return
	}

	if device.HaveService(bluez.PANU_SVCLASS_ID) {
		connTypes = append(connTypes, []string{
			"panu",
			"Personal Area Network",
		})
	}
	if device.HaveService(bluez.DIALUP_NET_SVCLASS_ID) {
		connTypes = append(connTypes, []string{
			"dun",
			"Dialup Network",
		})
	}

	if connTypes == nil {
		InfoMessage("No network options exist for "+device.Name, false)
		return
	}

	setSelectorMenu(
		"device",
		func(networkMenu *tview.Table) {
			row, _ := networkMenu.GetSelection()

			cell := networkMenu.GetCell(row, 0)
			if cell == nil {
				return
			}

			connType, ok := cell.GetReference().(string)
			if !ok {
				return
			}

			go networkConnect(device, connType)

		}, nil,
		func(networkMenu *tview.Table) (int, int) {
			var width int

			for row, connType := range connTypes {
				ctype := connType[0]
				description := connType[1]

				if len(description) > width {
					width = len(description)
				}

				networkMenu.SetCell(row, 0, tview.NewTableCell(description).
					SetExpansion(1).
					SetReference(ctype).
					SetAlign(tview.AlignLeft).
					SetTextColor(theme.GetColor("Text")).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor("Text")).
						Background(theme.BackgroundColor("Text")),
					),
				)
				networkMenu.SetCell(row, 1, tview.NewTableCell("("+strings.ToUpper(ctype)+")").
					SetAlign(tview.AlignRight).
					SetTextColor(theme.GetColor("Text")).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor("Text")).
						Background(theme.BackgroundColor("Text")),
					),
				)
			}

			return width, 0
		},
	)
}

// networkConnect connects to the network with the selected network type.
func networkConnect(device bluez.Device, connType string) {
	info := fmt.Sprintf("%s (%s)",
		device.Name, strings.ToUpper(connType),
	)

	startOperation(
		func() {
			InfoMessage("Connecting to "+info, true)
			err := NetworkConn.Connect(device.Name, connType, device.Address)
			if err != nil {
				ErrorMessage(err)
				return
			}
			InfoMessage("Connected to "+info, false)
		},
		func() {
			err := NetworkConn.DeactivateConnection(device.Address)
			if err != nil {
				ErrorMessage(err)
				return
			}
			InfoMessage("Cancelled connection to "+info, false)
		},
	)
}
