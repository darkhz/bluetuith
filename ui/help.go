package ui

import (
	"github.com/darkhz/bluetuith/theme"
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
		"Send files":                       "f",
		"Progress view":                    "v",
		"Show device information":          "i",
		"Connect to selected device":       "c",
		"Pair with selected device":        "p",
		"Trust selected device":            "t",
		"Remove device from adapter":       "d",
		"Cancel operation":                 "Ctrl+x",
		"Quit":                             "q",
	}

	filePickerKeyBindings := map[string]string{
		"Navigate between directory entries": "Up/Down",
		"Enter a directory":                  "Right",
		"Go back one directory":              "Left",
		"Select one file":                    "Space",
		"Invert file selection":              "a",
		"Select all files":                   "A",
		"Refresh current directory":          "Ctrl + r",
		"Toggle hidden files":                "Ctrl+h",
		"Confirm file(s) selection":          "Ctrl+s",
		"Exit":                               "Escape",
	}

	progressViewKeyBindings := map[string]string{
		"Navigate between transfers": "Up/Down",
		"Suspend transfer":           "z",
		"Resume transfer":            "g",
		"Cancel transfer":            "x",
		"Exit":                       "Escape",
	}

	helpTable := tview.NewTable()
	helpTable.SetBorder(true)
	helpTable.SetSelectable(true, false)
	helpTable.SetBorderColor(theme.GetColor("Border"))
	helpTable.SetTitle(theme.ColorWrap("Text", "[ HELP ]"))
	helpTable.SetBackgroundColor(theme.GetColor("Background"))
	helpTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			Pages.RemovePage("helpmodal")
		}

		return event
	})
	helpTable.SetSelectionChangedFunc(func(row, col int) {
		if row == 1 {
			helpTable.ScrollToBeginning()
		}
	})

	for title, helpMap := range map[string]map[string]string{
		"Device Screen": deviceKeyBindings,
		"File Picker":   filePickerKeyBindings,
		"Progress View": progressViewKeyBindings,
	} {
		helpTable.SetCell(row, 0, tview.NewTableCell("[::bu]"+title).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor("Text")),
		)

		row++

		for op, key := range helpMap {
			helpTable.SetCell(row, 0, tview.NewTableCell(theme.ColorWrap("Text", op)).
				SetExpansion(1).
				SetAlign(tview.AlignLeft).
				SetTextColor(theme.GetColor("Text")).
				SetSelectedStyle(tcell.Style{}.
					Foreground(theme.GetColor("Text")).
					Background(theme.BackgroundColor("Text")),
				),
			)

			helpTable.SetCell(row, 1, tview.NewTableCell(theme.ColorWrap("Text", key)).
				SetExpansion(0).
				SetAlign(tview.AlignLeft).
				SetTextColor(theme.GetColor("Text")).
				SetSelectedStyle(tcell.Style{}.
					Foreground(theme.GetColor("Text")).
					Background(theme.BackgroundColor("Text")),
				),
			)

			row++
		}

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
