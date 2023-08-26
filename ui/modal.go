package ui

import (
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

	y *tview.Flex
	x *tview.Flex
}

var modals []*Modal

// NewModal returns a modal. If a primitive is not provided,
// a table is attach to it.
func NewModal(name, title string, item tview.Primitive, height, width int) *Modal {
	var modal *Modal
	var table *tview.Table

	box := tview.NewBox()
	box.SetBackgroundColor(theme.GetColor("Background"))

	modalTitle := tview.NewTextView()
	modalTitle.SetDynamicColors(true)
	modalTitle.SetText("[::bu]" + title)
	modalTitle.SetTextAlign(tview.AlignCenter)
	modalTitle.SetTextColor(theme.GetColor("Text"))
	modalTitle.SetBackgroundColor(theme.GetColor("Background"))

	closeButton := tview.NewTextView()
	closeButton.SetRegions(true)
	closeButton.SetDynamicColors(true)
	closeButton.SetText(`["close"][::b][X[]`)
	closeButton.SetTextAlign(tview.AlignRight)
	closeButton.SetTextColor(theme.GetColor("Text"))
	closeButton.SetBackgroundColor(theme.GetColor("Background"))
	closeButton.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		modal.Exit(false)
	})

	titleFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(box, 0, 1, false).
		AddItem(modalTitle, 0, 1, false).
		AddItem(closeButton, 0, 1, false)

	if item == nil {
		table = tview.NewTable()
		table.SetSelectorWrap(true)
		table.SetSelectable(true, false)
		table.SetBackgroundColor(theme.GetColor("Background"))
		table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch cmd.KeyOperation(event) {
			case cmd.KeyClose:
				modal.Exit(false)
			}

			return event
		})

		item = table
	}

	flex := tview.NewFlex()
	flex.SetBorder(true)
	flex.SetDirection(tview.FlexRow)

	flex.AddItem(titleFlex, 1, 0, false)
	flex.AddItem(horizontalLine(), 1, 0, false)
	flex.AddItem(item, 0, 1, true)
	flex.SetBorderColor(theme.GetColor("Border"))
	flex.SetBackgroundColor(theme.GetColor("Background"))

	modal = &Modal{
		Name:  name,
		Flex:  flex,
		Table: table,

		Height: height,
		Width:  width,
	}

	return modal
}

// NewMenuModal returns a menu modal.
func NewMenuModal(name string, regionX, regionY int) *Modal {
	table := tview.NewTable()
	table.SetBorder(true)
	table.SetSelectable(true, false)
	table.SetBackgroundColor(tcell.ColorDefault)
	table.SetBorderColor(theme.GetColor("Border"))
	table.SetBackgroundColor(theme.GetColor("Background"))

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
