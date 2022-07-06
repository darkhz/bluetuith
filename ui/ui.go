package ui

import (
	"fmt"

	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

var (
	// App contains the application.
	App *tview.Application

	// Pages holds the DeviceTable, along with
	// any menu popups that will be added.
	Pages *tview.Pages
)

// StartUI starts the UI.
func StartUI() {
	App = tview.NewApplication()
	Pages = tview.NewPages()

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(menuBar(), 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(deviceTable(), 0, 10, true).
		AddItem(statusBar(), 1, 0, false)
	flex.SetBackgroundColor(tcell.ColorDefault)

	Pages.AddPage("main", flex, true, true)
	Pages.SetBackgroundColor(tcell.ColorDefault)

	App.SetFocus(flex)
	App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlX:
			cancelOperation(true)
		}

		return event
	})

	if err := setupDevices(); err != nil {
		fmt.Println(err.Error())
		return
	}

	InfoMessage("bluetuith is ready.", false)

	if err := App.SetRoot(Pages, true).SetFocus(DeviceTable).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

// StopUI stops the UI.
func StopUI() {
	stopStatus()

	App.Stop()
}

func confirmQuit() bool {
	return SetInput("Quit (y/n)?") == "y"
}
