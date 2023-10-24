package ui

import (
	"strings"
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

	suspend    bool
	warn, page string
	focus      tview.Primitive

	*tview.Application
}

var UI Application

// StartUI starts the UI.
func StartUI() {
	UI.Application = tview.NewApplication()
	UI.Pages = tview.NewPages()

	box := tview.NewBox().
		SetBackgroundColor(theme.GetColor(theme.ThemeMenuBar))

	menuArea := tview.NewFlex().
		AddItem(menuBar(), 0, 1, false).
		AddItem(box, 1, 0, false).
		AddItem(adapterStatusView(), 0, 1, false)
	menuArea.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(menuArea, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(deviceTable(), 0, 10, true)
	flex.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	UI.focus = flex
	UI.page = "main"
	UI.Pages.AddPage("main", flex, true, true)
	UI.Pages.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
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
	UI.Layout.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	UI.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		operation := cmd.KeyOperation(event)
		if operation != cmd.KeyQuit && event.Key() == tcell.KeyCtrlC {
			return nil
		}

		switch operation {
		case cmd.KeySuspend:
			UI.suspend = true

		case cmd.KeyCancel:
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
	displayWarning()
	updateAdapterStatus(UI.Bluez.GetCurrentAdapter())
	setAdapterStates()

	InfoMessage("bluetuith is ready.", false)

	if err := UI.SetRoot(UI.Layout, true).SetFocus(UI.focus).EnableMouse(true).Run(); err != nil {
		cmd.PrintError("Cannot initialize application", err)
	}
}

// StopUI stops the UI.
func StopUI() {
	stopStatus()

	UI.Stop()
}

// SetConnections sets the connections to bluez and networkmanager.
func SetConnections(b *bluez.Bluez, o *bluez.Obex, n *network.Network, warn string) {
	UI.Bluez = b
	UI.Obex = o
	UI.Network = n
	UI.warn = warn
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

func displayWarning() {
	if UI.warn == "" || cmd.IsPropertyEnabled("no-warning") {
		return
	}

	height := 10
	if strings.Count(UI.warn, "\n") > 2 {
		height += 4
	}

	UI.warn += "\n\n-- Press any key to close this dialog --\n\n"

	warningTextView := tview.NewTextView()
	warningTextView.SetText(UI.warn)
	warningTextView.SetDynamicColors(true)
	warningTextView.SetTextAlign(tview.AlignCenter)
	warningTextView.SetTextColor(theme.GetColor(theme.ThemeText))
	warningTextView.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	modal := NewModal("warning", "Warning", warningTextView, height, 60)
	warningTextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		modal.Exit(false)

		return event
	})

	UI.focus = modal.Flex
	modal.Show()
}

func confirmQuit() bool {
	return SetInput("Quit (y/n)?") == "y"
}
