package ui

import (
	"strings"

	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

func showHelp() {
	var row int

	deviceKeyBindings := map[string][]cmd.Key{
		"Open the menu":                    {cmd.KeyMenu},
		"Navigate between menus":           {cmd.KeySwitch},
		"Navigate between devices/options": {cmd.KeyNavigateUp, cmd.KeyNavigateDown},
		"Toggle adapter power state":       {cmd.KeyAdapterTogglePower},
		"Toggle discoverable state":        {cmd.KeyAdapterToggleDiscoverable},
		"Toggle pairable state":            {cmd.KeyAdapterTogglePairable},
		"Toggle scan (discovery state)":    {cmd.KeyAdapterToggleScan},
		"Change adapter":                   {cmd.KeyAdapterChange},
		"Send files":                       {cmd.KeyDeviceSendFiles},
		"Connect to network":               {cmd.KeyDeviceNetwork},
		"Progress view":                    {cmd.KeyProgressView},
		"Show/Hide player":                 {cmd.KeyPlayerShow, cmd.KeyPlayerHide},
		"Show device information":          {cmd.KeyDeviceInfo},
		"Connect to selected device":       {cmd.KeyDeviceConnect},
		"Pair with selected device":        {cmd.KeyDevicePair},
		"Trust selected device":            {cmd.KeyDeviceTrust},
		"Remove device from adapter":       {cmd.KeyDeviceRemove},
		"Cancel operation":                 {cmd.KeyCancel},
		"Quit":                             {cmd.KeyQuit},
	}

	filePickerKeyBindings := map[string][]cmd.Key{
		"Navigate between directory entries": {cmd.KeyNavigateUp, cmd.KeyNavigateDown},
		"Enter/Go back a directory":          {cmd.KeyNavigateRight, cmd.KeyNavigateLeft},
		"Select one file":                    {cmd.KeyFilebrowserSelect},
		"Invert file selection":              {cmd.KeyFilebrowserInvertSelection},
		"Select all files":                   {cmd.KeyFilebrowserSelectAll},
		"Refresh current directory":          {cmd.KeyFilebrowserRefresh},
		"Toggle hidden files":                {cmd.KeyFilebrowserToggleHidden},
		"Confirm file(s) selection":          {cmd.KeyFilebrowserConfirmSelection},
		"Exit":                               {cmd.KeyClose},
	}

	progressViewKeyBindings := map[string][]cmd.Key{
		"Navigate between transfers": {cmd.KeyNavigateUp, cmd.KeyNavigateDown},
		"Suspend transfer":           {cmd.KeyProgressTransferSuspend},
		"Resume transfer":            {cmd.KeyProgressTransferResume},
		"Cancel transfer":            {cmd.KeyProgressTransferCancel},
		"Exit":                       {cmd.KeyClose},
	}

	mediaPlayerKeyBindings := map[string][]cmd.Key{
		"Toggle play/pause": {cmd.KeyNavigateUp, cmd.KeyNavigateDown},
		"Next":              {cmd.KeyPlayerNext},
		"Previous":          {cmd.KeyPlayerPrevious},
		"Rewind":            {cmd.KeyPlayerSeekBackward},
		"Fast forward":      {cmd.KeyPlayerSeekForward},
		"Stop":              {cmd.KeyPlayerStop},
	}

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

	for title, helpMap := range map[string]map[string][]cmd.Key{
		"Device Screen": deviceKeyBindings,
		"File Picker":   filePickerKeyBindings,
		"Progress View": progressViewKeyBindings,
		"Media Player":  mediaPlayerKeyBindings,
	} {
		helpModal.Table.SetCell(row, 0, tview.NewTableCell("[::bu]"+title).
			SetSelectable(false).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeText)),
		)

		row++

		for op, key := range helpMap {
			var names []string

			for _, k := range key {
				names = append(names, cmd.KeyName(cmd.OperationData(k).Kb))
			}

			keybinding := strings.Join(names, "/")

			helpModal.Table.SetCell(row, 0, tview.NewTableCell(theme.ColorWrap(theme.ThemeText, op)).
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
