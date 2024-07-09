package ui

import (
	"strings"

	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// Help describes the help item.
type Help struct {
	Title, Description string
	Keys               []cmd.Key
	ShowInStatus       bool
}

type HelpArea struct {
	page string
	area *tview.Flex
}

var (
	help HelpArea

	// HelpTopics store the various help topics.
	HelpTopics = map[string][]Help{
		"Device Screen": {
			{"Menu", "Open the menu", []cmd.Key{cmd.KeyMenu}, true},
			{"Switch", "Navigate between menus", []cmd.Key{cmd.KeySwitch}, true},
			{"Navigation", "Navigate between devices/options", []cmd.Key{cmd.KeyNavigateUp, cmd.KeyNavigateDown}, true},
			{"Power", "Toggle adapter power state", []cmd.Key{cmd.KeyAdapterTogglePower}, true},
			{"Discoverable", "Toggle discoverable state", []cmd.Key{cmd.KeyAdapterToggleDiscoverable}, false},
			{"Pairable", "Toggle pairable state", []cmd.Key{cmd.KeyAdapterTogglePairable}, false},
			{"Scan", "Toggle scan (discovery state)", []cmd.Key{cmd.KeyAdapterToggleScan}, true},
			{"Adapter", "Change adapter", []cmd.Key{cmd.KeyAdapterChange}, true},
			{"Send", "Send files", []cmd.Key{cmd.KeyDeviceSendFiles}, true},
			{"Network", "Connect to network", []cmd.Key{cmd.KeyDeviceNetwork}, false},
			{"Progress", "Progress view", []cmd.Key{cmd.KeyProgressView}, false},
			{"Player", "Show/Hide player", []cmd.Key{cmd.KeyPlayerShow, cmd.KeyPlayerHide}, false},
			{"Device Info", "Show device information", []cmd.Key{cmd.KeyDeviceInfo}, false},
			{"Connect", "Toggle connection with selected device", []cmd.Key{cmd.KeyDeviceConnect}, true},
			{"Pair", "Toggle pair with selected device", []cmd.Key{cmd.KeyDevicePair}, true},
			{"Trust", "Toggle trust with selected device", []cmd.Key{cmd.KeyDeviceTrust}, false},
			{"Remove", "Remove device from adapter", []cmd.Key{cmd.KeyDeviceRemove}, false},
			{"Cancel", "Cancel operation", []cmd.Key{cmd.KeyCancel}, false},
			{"Help", "Show help", []cmd.Key{cmd.KeyHelp}, true},
			{"Quit", "Quit", []cmd.Key{cmd.KeyQuit}, false},
		},
		"File Picker": {
			{"Navigation", "Navigate between directory entries", []cmd.Key{cmd.KeyNavigateUp, cmd.KeyNavigateDown}, true},
			{"ChgDir Fwd/Back", "Enter/Go back a directory", []cmd.Key{cmd.KeyNavigateRight, cmd.KeyNavigateLeft}, true},
			{"One", "Select one file", []cmd.Key{cmd.KeyFilebrowserSelect}, true},
			{"Invert", "Invert file selection", []cmd.Key{cmd.KeyFilebrowserInvertSelection}, true},
			{"All", "Select all files", []cmd.Key{cmd.KeyFilebrowserSelectAll}, true},
			{"Refresh", "Refresh current directory", []cmd.Key{cmd.KeyFilebrowserRefresh}, false},
			{"Hidden", "Toggle hidden files", []cmd.Key{cmd.KeyFilebrowserToggleHidden}, false},
			{"Confirm", "Confirm file(s) selection", []cmd.Key{cmd.KeyFilebrowserConfirmSelection}, true},
			{"Exit", "Exit", []cmd.Key{cmd.KeyClose}, false},
		},
		"Progress View": {
			{"Navigation", "Navigate between transfers", []cmd.Key{cmd.KeyNavigateUp, cmd.KeyNavigateDown}, true},
			{"Suspend", "Suspend transfer", []cmd.Key{cmd.KeyProgressTransferSuspend}, true},
			{"Resume", "Resume transfer", []cmd.Key{cmd.KeyProgressTransferResume}, true},
			{"Cancel", "Cancel transfer", []cmd.Key{cmd.KeyProgressTransferCancel}, true},
			{"Exit", "Exit", []cmd.Key{cmd.KeyClose}, true},
		},
		"Media Player": {
			{"Play/Pause", "Toggle play/pause", []cmd.Key{cmd.KeyNavigateUp, cmd.KeyNavigateDown}, false},
			{"Next", "Next", []cmd.Key{cmd.KeyPlayerNext}, false},
			{"Previous", "Previous", []cmd.Key{cmd.KeyPlayerPrevious}, false},
			{"Rewind", "Rewind", []cmd.Key{cmd.KeyPlayerSeekBackward}, false},
			{"Forward", "Fast forward", []cmd.Key{cmd.KeyPlayerSeekForward}, false},
			{"Stop", "Stop", []cmd.Key{cmd.KeyPlayerStop}, false},
		},
	}
)

func statusHelpArea(add bool) {
	if cmd.IsPropertyEnabled("no-help-display") {
		return
	}

	if !add && help.area != nil {
		UI.Layout.RemoveItem(help.area)
		return
	}

	if help.area == nil {
		help.area = tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(horizontalLine(), 1, 0, false).
			AddItem(UI.Status.Help, 1, 0, false)
	}

	UI.Layout.AddItem(help.area, 2, 0, false)
}

func showStatusHelp(page string) {
	if cmd.IsPropertyEnabled("no-help-display") || help.page == UI.page {
		return
	}

	help.page = UI.page
	pages := map[string]string{
		"main":         "Device Screen",
		"filepicker":   "File Picker",
		"progressview": "Progress View",
	}

	items, ok := HelpTopics[pages[page]]
	if !ok {
		UI.Status.Help.Clear()
		return
	}

	groups := map[string][]Help{}

	for _, item := range items {
		if !item.ShowInStatus {
			continue
		}

		var group string

		for _, key := range item.Keys {
			switch key {
			case cmd.KeyMenu, cmd.KeySwitch:
				group = "Open"

			case cmd.KeyFilebrowserSelect, cmd.KeyFilebrowserInvertSelection, cmd.KeyFilebrowserSelectAll:
				group = "Select"

			case cmd.KeyProgressTransferSuspend, cmd.KeyProgressTransferResume, cmd.KeyProgressTransferCancel:
				group = "Transfer"

			case cmd.KeyDeviceConnect, cmd.KeyDevicePair, cmd.KeyAdapterToggleScan, cmd.KeyAdapterTogglePower:
				group = "Toggle"
			}
		}
		if group == "" {
			group = item.Title
		}

		helpItem := groups[group]
		if helpItem == nil {
			helpItem = []Help{}
		}

		helpItem = append(helpItem, item)
		groups[group] = helpItem
	}

	text := ""
	count := 0
	for group, items := range groups {
		var names, keys []string

		for _, item := range items {
			if item.Title != group {
				names = append(names, item.Title)
			}
			for _, k := range item.Keys {
				keys = append(keys, cmd.KeyName(cmd.OperationData(k).Kb))
			}
		}
		if names != nil {
			group += " " + strings.Join(names, "/")
		}

		helpKeys := strings.Join(keys, "/")
		if count < len(groups)-1 {
			helpKeys += ", "
		}

		title := theme.ColorWrap(theme.ThemeText, group, "::bu")
		helpKeys = theme.ColorWrap(theme.ThemeText, ": "+helpKeys)

		text += title + helpKeys
		count++
	}

	UI.Status.Help.SetText(text)
}

func showHelp() {
	var row int

	helpModal := NewModal("help", "Help", nil, 40, 60)
	helpModal.Table.SetSelectionChangedFunc(func(row, col int) {
		if row == 1 {
			helpModal.Table.ScrollToBeginning()
		}
	})
	helpModal.Table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseScrollUp {
			helpModal.Table.InputHandler()(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone), nil)
		}

		return action, event
	})

	for title, helpItems := range HelpTopics {
		helpModal.Table.SetCell(row, 0, tview.NewTableCell("[::bu]"+title).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)

		row++

		for _, item := range helpItems {
			var names []string

			for _, k := range item.Keys {
				names = append(names, cmd.KeyName(cmd.OperationData(k).Kb))
			}

			keybinding := strings.Join(names, "/")

			helpModal.Table.SetCell(row, 0, tview.NewTableCell(theme.ColorWrap(theme.ThemeText, item.Description)).
				SetExpansion(1).
				SetAlign(tview.AlignLeft).
				SetTextColor(theme.GetColor(theme.ThemeText)).
				SetSelectedStyle(tcell.Style{}.
					Foreground(theme.GetColor(theme.ThemeText)).
					Background(theme.BackgroundColor(theme.ThemeText)),
				),
			)

			helpModal.Table.SetCell(row, 1, tview.NewTableCell(theme.ColorWrap(theme.ThemeText, keybinding)).
				SetExpansion(0).
				SetAlign(tview.AlignLeft).
				SetTextColor(theme.GetColor(theme.ThemeText)).
				SetSelectedStyle(tcell.Style{}.
					Foreground(theme.GetColor(theme.ThemeText)).
					Background(theme.BackgroundColor(theme.ThemeText)),
				),
			)

			row++
		}

		row++

	}

	helpModal.Show()
}
