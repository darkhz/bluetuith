package ui

import (
	"sort"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

var device bluez.Device

// audioProfiles shows a popup to select the audio profile.
func audioProfiles() {
	device = getDeviceFromSelection(false)
	if device.Path == "" {
		return
	}

	profiles, err := bluez.ListAudioProfiles(device.Address)
	if err != nil {
		ErrorMessage(err)
		return
	}
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name == "off"
	})

	setContextMenu(
		"device",
		func(profileMenu *tview.Table) {
			row, _ := profileMenu.GetSelection()

			setProfile(profileMenu, row, 0)

		}, nil,
		func(profileMenu *tview.Table) (int, int) {
			var width, index int

			profileMenu.SetSelectorWrap(true)

			for row, profile := range profiles {
				if profile.Active {
					index = row
				}

				if len(profile.Description) > width {
					width = len(profile.Description)
				}

				profileMenu.SetCellSimple(row, 0, "")

				profileMenu.SetCell(row, 1, tview.NewTableCell(profile.Description).
					SetExpansion(1).
					SetReference(profile).
					SetAlign(tview.AlignLeft).
					SetOnClickedFunc(setProfile).
					SetTextColor(theme.GetColor(theme.ThemeText)).
					SetSelectedStyle(tcell.Style{}.
						Foreground(theme.GetColor(theme.ThemeText)).
						Background(theme.BackgroundColor(theme.ThemeText)),
					),
				)

			}

			markActiveProfile(profileMenu, device, index)

			return width - 16, index
		},
	)
}

// setProfile sets the selected audio profile.
func setProfile(profileMenu *tview.Table, row, column int) {
	cell := profileMenu.GetCell(row, 1)
	if cell == nil {
		return
	}

	profile, ok := cell.GetReference().(bluez.AudioProfile)
	if !ok {
		return
	}

	if err := profile.SetAudioProfile(); err != nil {
		ErrorMessage(err)
		return
	}

	markActiveProfile(profileMenu, device, row)
}

// markActiveProfile marks the active profile in the profiles list.
func markActiveProfile(profileMenu *tview.Table, device bluez.Device, index ...int) {
	for i := 0; i < profileMenu.GetRowCount(); i++ {
		var activeIndicator string

		if i == index[0] {
			activeIndicator = string('\u2022')
		} else {
			activeIndicator = ""
		}

		profileMenu.SetCell(i, 0, tview.NewTableCell(activeIndicator).
			SetSelectable(false).
			SetTextColor(theme.GetColor(theme.ThemeText)).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor(theme.ThemeText)).
				Background(theme.BackgroundColor(theme.ThemeText)),
			),
		)
	}
}
