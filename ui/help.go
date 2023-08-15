package ui

import (
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

func showHelp() {
	var row int

	deviceKeyBindings := map[string]string{
		"Open the menu":                    "Alt+m",
		"Navigate between menus":           "Tab",
		"Navigate between devices/options": "Up/Down",
		"Toggle adapter power state":       "o",
		"Toggle discoverable state":        "S",
		"Toggle pairable state":            "P",
		"Toggle scan (discovery state)":    "s",
		"Change adapter":                   "a",
		"Send files":                       "f",
		"Connect to network":               "n",
		"Progress view":                    "v",
		"Show player":                      "m",
		"Hide player":                      "M",
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

	mediaPlayerKeyBindings := map[string]string{
		"Toggle play/pause": "Space",
		"Next":              ">",
		"Previous":          "<",
		"Rewind":            "Left",
		"Fast forward":      "Right",
		"Stop":              "]",
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
			UI.Pages.RemovePage("helpmodal")
		}

		return event
	})
	helpTable.SetSelectionChangedFunc(func(row, col int) {
		if row == 1 {
			helpTable.ScrollToBeginning()
		}
	})
	helpModal.Table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseScrollUp {
			helpModal.Table.InputHandler()(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone), nil)
		}

		return action, event
	})

	for title, helpMap := range map[string]map[string]string{
		"Device Screen": deviceKeyBindings,
		"File Picker":   filePickerKeyBindings,
		"Progress View": progressViewKeyBindings,
		"Media Player":  mediaPlayerKeyBindings,
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

	ShowModal("helpmodal", helpTable, 20, 60)
}
