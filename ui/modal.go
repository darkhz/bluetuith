package ui

import (
	"strings"

	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// Modal stores a layout to display a floating modal.
type Modal struct {
	Name          string
	Open          bool
	Height, Width int

	menu                                    bool
	regionX, regionY, pageHeight, pageWidth int

	Flex  *tview.Flex
	Table *tview.Table

	y      *tview.Flex
	x      *tview.Flex
	button *tview.TextView
}

var modals []*Modal

// NewModal returns a modal. If a primitive is not provided,
// a table is attach to it.
func NewModal(name, title string, item tview.Primitive, height, width int) *Modal {
	var modal *Modal
	var table *tview.Table

	box := tview.NewBox()
	box.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	modalTitle := tview.NewTextView()
	modalTitle.SetDynamicColors(true)
	modalTitle.SetText("[::bu]" + title)
	modalTitle.SetTextAlign(tview.AlignCenter)
	modalTitle.SetTextColor(theme.GetColor(theme.ThemeText))
	modalTitle.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	closeButton := tview.NewTextView()
	closeButton.SetRegions(true)
	closeButton.SetDynamicColors(true)
	closeButton.SetText(`["close"][::b][X[]`)
	closeButton.SetTextAlign(tview.AlignRight)
	closeButton.SetTextColor(theme.GetColor(theme.ThemeText))
	closeButton.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	closeButton.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		modal.Exit(false)
	})

	titleFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(box, 0, 1, false).
		AddItem(modalTitle, 0, 10, false).
		AddItem(closeButton, 0, 1, false)

	if item == nil {
		table = tview.NewTable()
		table.SetSelectorWrap(true)
		table.SetSelectable(true, false)
		table.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
		table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch cmd.KeyOperation(event) {
			case cmd.KeyClose:
				modal.Exit(false)
			}

			return ignoreDefaultEvent(event)
		})

		item = table
	}

	flex := tview.NewFlex()
	flex.SetBorder(true)
	flex.SetDirection(tview.FlexRow)

	flex.AddItem(titleFlex, 1, 0, false)
	flex.AddItem(horizontalLine(), 1, 0, false)
	flex.AddItem(item, 0, 1, true)
	flex.SetBorderColor(theme.GetColor(theme.ThemeBorder))
	flex.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	modal = &Modal{
		Name:  name,
		Flex:  flex,
		Table: table,

		Height: height,
		Width:  width,

		button: closeButton,
	}

	return modal
}

// NewMenuModal returns a menu modal.
func NewMenuModal(name string, regionX, regionY int) *Modal {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, false)
	table.SetBackgroundColor(tcell.ColorDefault)
	table.SetBorderColor(theme.GetColor(theme.ThemeBorder))
	table.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true)

	return &Modal{
		Name:  name,
		Table: table,
		Flex:  flex,

		menu:    true,
		regionX: regionX,
		regionY: regionY,
	}
}

// NewDisplayModal displays a modal with a message.
func NewDisplayModal(name, title, message string) {
	message += "\n\nPress any key or click the 'X' button to close this dialog."

	width := len(strings.Split(message, "\n")[0])
	if width > 100 {
		width = 100
	}

	height := len(tview.WordWrap(message, width)) * 4

	textview := tview.NewTextView()
	textview.SetText(message)
	textview.SetDynamicColors(true)
	textview.SetTextAlign(tview.AlignCenter)
	textview.SetTextColor(theme.GetColor(theme.ThemeText))
	textview.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	modal := NewModal(name, title, textview, height, width)
	textview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		modal.Exit(false)

		return event
	})

	go UI.QueueUpdateDraw(func() {
		modal.Show()
	})
}

// NewConfirmModal displays a modal, shows a message and asks for confirmation.
func NewConfirmModal(name, title, message string) string {
	var modal *Modal

	message += "\n\nPress y/n to Confirm/Cancel, click the required button or click the 'X' button to close this dialog."

	reply := make(chan string, 10)

	send := func(msg string) {
		modal.Exit(false)

		select {
		case reply <- msg:
		}
	}

	width := len(strings.Split(message, "\n")[0])
	if width > 100 {
		width = 100
	}

	height := len(tview.WordWrap(message, width)) * 4

	buttons := tview.NewTextView()
	buttons.SetRegions(true)
	buttons.SetDynamicColors(true)
	buttons.SetTextAlign(tview.AlignCenter)
	buttons.SetTextColor(theme.GetColor(theme.ThemeText))
	buttons.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	buttons.SetText(`["confirm"][::b][Confirm[] ["cancel"][::b][Cancel[]`)
	buttons.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		switch added[0] {
		case "confirm":
			send("y")

		case "cancel":
			send("n")
		}
	})

	textview := tview.NewTextView()
	textview.SetText(message)
	textview.SetDynamicColors(true)
	textview.SetTextAlign(tview.AlignCenter)
	textview.SetTextColor(theme.GetColor(theme.ThemeText))
	textview.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textview, 0, 1, false).
		AddItem(buttons, 1, 0, true)

	modal = NewModal(name, title, flex, height, width)
	buttons.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'y', 'n':
			send(string(event.Rune()))
		}

		switch cmd.KeyOperation(event) {
		case cmd.KeyClose:
			send("n")
		}

		return event
	})
	modal.button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		send("n")
		return event
	})

	go UI.QueueUpdateDraw(func() {
		modal.Show()
	})

	return <-reply
}

// Show shows the modal.
func (m *Modal) Show() {
	var x, y, xprop, xattach int

	if len(modals) > 0 && modals[len(modals)-1].Name == m.Name {
		return
	}

	switch {
	case m.menu:
		xprop = 1
		x, y = m.regionX, m.regionY

	default:
		xattach = 1
	}

	m.Open = true

	m.y = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, y, 0, false).
		AddItem(m.Flex, m.Height, 0, true).
		AddItem(nil, 1, 0, false)

	m.x = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, x, xattach, false).
		AddItem(m.y, m.Width, 0, true).
		AddItem(nil, xprop, xattach, false)

	UI.Pages.AddAndSwitchToPage(m.Name, m.x, true)
	for _, modal := range modals {
		UI.Pages.ShowPage(modal.Name)
	}
	UI.Pages.ShowPage(UI.page)

	UI.SetFocus(m.Flex)

	modals = append(modals, m)
	ResizeModal()
}

// Exit exits the modal.
func (m *Modal) Exit(focusInput bool) {
	if m == nil {
		return
	}

	m.Open = false
	m.pageWidth = 0
	m.pageHeight = 0

	UI.Pages.RemovePage(m.Name)

	for i, modal := range modals {
		if modal == m {
			modals[i] = modals[len(modals)-1]
			modals = modals[:len(modals)-1]

			break
		}
	}

	if focusInput {
		UI.SetFocus(UI.Status.InputField)
		return
	}

	SetPrimaryFocus()
}

// ResizeModal resizes the modal according to the current screen dimensions.
func ResizeModal() {
	var drawn bool

	for _, modal := range modals {
		_, _, pageWidth, pageHeight := UI.Layout.GetInnerRect()

		if modal == nil || !modal.Open ||
			(modal.pageHeight == pageHeight && modal.pageWidth == pageWidth) {
			continue
		}

		modal.pageHeight = pageHeight
		modal.pageWidth = pageWidth

		height := modal.Height
		width := modal.Width
		if height >= pageHeight {
			height = pageHeight
		}
		if width >= pageWidth {
			width = pageWidth
		}

		var x, y int

		if modal.menu {
			x, y = modal.regionX, modal.regionY
		} else {
			x = (pageWidth - modal.Width) / 2
			y = (pageHeight - modal.Height) / 2
		}

		modal.y.ResizeItem(modal.Flex, height, 0)
		modal.y.ResizeItem(nil, y, 0)

		modal.x.ResizeItem(modal.y, width, 0)
		modal.x.ResizeItem(nil, x, 0)

		drawn = true
	}

	if drawn {
		go UI.Draw()
	}
}

// SetPrimaryFocus sets the focus to the appropriate primitive.
func SetPrimaryFocus() {
	if pg, _ := UI.Status.GetFrontPage(); pg == "input" {
		UI.SetFocus(UI.Status.InputField)
		return
	}

	if len(modals) > 0 {
		UI.SetFocus(modals[len(modals)-1].Flex)
		return
	}

	UI.SetFocus(UI.Pages)
}

// ModalExists returns whether the modal with the given name is displayed.
func ModalExists(name string) (*Modal, bool) {
	var names []string
	for _, modal := range modals {
		names = append(names, modal.Name)
		if modal.Name == name {
			return modal, true
		}
	}

	return nil, false
}

// modalMouseHandler handles mouse events for a modal.
func modalMouseHandler(event *tcell.EventMouse, action tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
	for _, modal := range modals {
		if modal == nil || !modal.Open {
			continue
		}

		x, y := event.Position()

		switch action {
		case tview.MouseRightClick:
			if menu.bar.InRect(x, y) {
				return nil, action
			} else if DeviceTable.InRect(x, y) {
				menu.bar.Highlight("")
			}

		case tview.MouseLeftClick:
			if modal.Flex.InRect(x, y) {
				UI.SetFocus(modal.Flex)
			} else {
				if modal.menu {
					exitMenu()
					continue
				}

				modal.Exit(false)
			}
		}
	}

	return event, action
}
