package ui

import (
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

var prevPage string

func showHelp() {
	var row int

	prevPage, _ = Pages.GetFrontPage()

	deviceKeyBindings := map[string]string{
		"Open the menu":                    "Ctrl+m",
		"Navigate between menus":           "Tab",
		"Navigate between devices/options": "Up/Down",
		"Toggle adapter power state":       "o",
		"Toggle scan (discovery state)":    "s",
		"Change adapter":                   "a",
		"Connect to selected device":       "c",
		"Pair with selected device":        "p",
		"Trust selected device":            "t",
		"Remove device from adapter":       "r",
		"Cancel operation":                 "Ctrl+x",
		"Quit":                             "q",
	}

	helpTable := tview.NewTable()
	helpTable.SetBorder(true)
	helpTable.SetTitle("[ HELP ]")
	helpTable.SetSelectable(true, false)
	helpTable.SetBackgroundColor(tcell.ColorDefault)
	helpTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			Pages.RemovePage("helpmodal")
		}

		return event
	})

	for op, key := range deviceKeyBindings {
		helpTable.SetCell(row, 0, tview.NewTableCell(op).
			SetExpansion(1).
			SetAlign(tview.AlignLeft),
		)

		helpTable.SetCell(row, 1, tview.NewTableCell(key).
			SetExpansion(0).
			SetAlign(tview.AlignLeft),
		)

		row++
	}

	helpWrap := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 20, false).
		AddItem(helpTable, 0, 20, true).
		AddItem(nil, 0, 20, false)

	helpFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, 60, false).
		AddItem(helpWrap, 0, 60, true).
		AddItem(nil, 0, 60, false)

	Pages.AddAndSwitchToPage("helpmodal", helpFlex, true).ShowPage(prevPage)
}
