package ui

import (
	"syscall"

	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

var (
	// App contains the application.
	App *tview.Application

	// Pages holds the DeviceTable, along with
	// any menu popups that will be added.
	Pages *tview.Pages

	appSuspend bool
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
		case tcell.KeyCtrlC:
			return nil

		case tcell.KeyCtrlZ:
			appSuspend = true

		case tcell.KeyCtrlX:
			cancelOperation(true)
		}

		return event
	})
	App.SetBeforeDrawFunc(func(t tcell.Screen) bool {
		suspendUI(t)

		return false
	})

	setupDevices()
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

// suspendUI suspends the application.
func suspendUI(t tcell.Screen) {
	if !appSuspend {
		return
	}

	appSuspend = false

	if err := t.Suspend(); err != nil {
		return
	}
	if err := syscall.Kill(syscall.Getpid(), syscall.SIGSTOP); err != nil {
		return
	}
	if err := t.Resume(); err != nil {
		return
	}
}

func confirmQuit() bool {
	return SetInput("Quit (y/n)?") == "y"
}
