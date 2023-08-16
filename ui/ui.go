package ui

import (
	"syscall"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/network"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

type Application struct {
	// Pages holds the DeviceTable, along with
	// any menu popups that will be added.
	Pages *tview.Pages

	// Layout holds the layout of the application.
	Layout *tview.Flex

	// Status holds the status bar.
	Status Status

	// Bluez holds the current bluez DBus connection.
	Bluez *bluez.Bluez

	// Obex holds the current bluez obex DBus connection.
	Obex *bluez.Obex

	// network holds the current network connection.
	Network *network.Network

	suspend bool
	page string

	*tview.Application
}

var UI Application

// StartUI starts the UI.
func StartUI() {
	UI.Application = tview.NewApplication()
	UI.Pages = tview.NewPages()

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(menuBar(), 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(deviceTable(), 0, 10, true)
	flex.SetBackgroundColor(theme.GetColor("Background"))

	UI.page = "main"
	UI.Pages.AddPage("main", flex, true, true)
	UI.Pages.SetBackgroundColor(theme.GetColor("Background"))
	UI.Pages.SetChangedFunc(func() {
		page, _ := UI.Pages.GetFrontPage()

		switch page {
		case "main", "filepicker", "progressview":
			UI.page = page
		}
	})

	UI.Layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(UI.Pages, 0, 10, true).
		AddItem(statusBar(), 1, 0, false)
	UI.Layout.SetBackgroundColor(theme.GetColor("Background"))

	UI.SetFocus(flex)
	UI.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			return nil

		case tcell.KeyCtrlZ:
			UI.suspend = true

		case tcell.KeyCtrlX:
			cancelOperation(true)
		}

		return event
	})
	UI.SetMouseCapture(func(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
		return modalMouseHandler(event, action)
	})
	UI.SetBeforeDrawFunc(func(t tcell.Screen) bool {
		ResizeModal()
		suspendUI(t)

		return false
	})

	setupDevices()
	InfoMessage("bluetuith is ready.", false)

	if err := UI.SetRoot(UI.Layout, true).SetFocus(DeviceTable).EnableMouse(true).Run(); err != nil {
		cmd.PrintError("Cannot initialize application", err)
	}
}

// StopUI stops the UI.
func StopUI() {
	stopStatus()

	UI.Stop()
}

// SetConnections sets the connections to bluez and networkmanager.
func SetConnections(b *bluez.Bluez, o *bluez.Obex, n *network.Network) {
	UI.Bluez = b
	UI.Obex = o
	UI.Network = n
}

// suspendUI suspends the application.
func suspendUI(t tcell.Screen) {
	if !UI.suspend {
		return
	}

	UI.suspend = false

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
