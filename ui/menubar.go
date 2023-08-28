package ui

import (
	"sync"

	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// Menu describes a region to display menu items.
type Menu struct {
	bar   *tview.TextView
	modal *Modal

	options map[string]map[cmd.Key]*MenuOption

	lock sync.Mutex
}

// MenuOption describes an option layout for a submenu.
type MenuOption struct {
	index int

	Key cmd.Key

	Display           string
	Enabled, Disabled string

	Toggled                    bool
	OnCreate, OnClick, Visible bool

	*cmd.KeyData
}

const menuBarRegions = `["adapter"][::b][Adapter[][""] ["device"][::b][Device[][""]`

var (
	menu Menu

	menuKeybindings = map[string][]*MenuOption{
		"adapter": {
			{
				Key:      cmd.KeyAdapterTogglePower,
				Enabled:  "On",
				Disabled: "Off",
				OnClick:  true,
				OnCreate: true,
			},
			{
				Key:      cmd.KeyAdapterToggleDiscoverable,
				Enabled:  "On",
				Disabled: "Off",
				OnClick:  true,
				OnCreate: true,
			},
			{
				Key:      cmd.KeyAdapterTogglePairable,
				Enabled:  "On",
				Disabled: "Off",
				OnClick:  true,
				OnCreate: true,
			},
			{
				Key:      cmd.KeyAdapterToggleScan,
				Disabled: "Stop Scan",
				OnClick:  true,
			},
			{
				Key:     cmd.KeyAdapterChange,
				OnClick: true,
			},
			{
				Key:     cmd.KeyProgressView,
				OnClick: true,
			},
			{
				Key:     cmd.KeyPlayerHide,
				OnClick: true,
			},
			{
				Key:     cmd.KeyQuit,
				OnClick: true,
			},
		},
		"device": {
			{
				Key:      cmd.KeyDeviceConnect,
				Disabled: "Disconnect",
				OnClick:  true,
				OnCreate: true,
			},
			{
				Key:     cmd.KeyDevicePair,
				OnClick: true,
			},
			{
				Key:      cmd.KeyDeviceTrust,
				Disabled: "Untrust",
				OnClick:  true,
				OnCreate: true,
			},
			{
				Key:     cmd.KeyDeviceSendFiles,
				OnClick: true,
				Visible: true,
			},
			{
				Key:     cmd.KeyDeviceNetwork,
				OnClick: true,
				Visible: true,
			},
			{
				Key:     cmd.KeyDeviceAudioProfiles,
				OnClick: true,
				Visible: true,
			},
			{
				Key:     cmd.KeyPlayerShow,
				OnClick: true,
				Visible: true,
			},
			{
				Key:     cmd.KeyDeviceInfo,
				OnClick: true,
			},
			{
				Key:     cmd.KeyDeviceRemove,
				OnClick: true,
			},
		},
	}
)

// menuBar sets up and returns the menu bar.
func menuBar() *tview.TextView {
	setupMenuOptions()

	menu.bar = tview.NewTextView()
	menu.bar.SetRegions(true)
	menu.bar.SetDynamicColors(true)
	menu.bar.SetBackgroundColor(theme.GetColor(theme.ThemeMenuBar))
	menu.bar.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		for _, region := range menu.bar.GetRegionInfos() {
			if region.ID == added[0] {
				if region.ID == "adapterchange" {
					adapterChange()
					menu.bar.Highlight("")
				} else {
					setMenu(region.FromX, 1, added[0])
				}

				break
			}
		}
	})

	menu.modal = NewMenuModal("menu", 0, 0)

	return menu.bar
}

// setupMenuOptions sets up the menu options with its attributes.
func setupMenuOptions() {
	menu.options = make(map[string]map[cmd.Key]*MenuOption)

	for menuName, keybindings := range menuKeybindings {
		if menu.options[menuName] == nil {
			menu.options[menuName] = make(map[cmd.Key]*MenuOption)

			for index, keybinding := range keybindings {
				keybinding.index = index
				menu.options[menuName][keybinding.Key] = keybinding
				keybinding.KeyData = cmd.OperationData(keybinding.Key)
			}
		}

	}
}

// setMenuItemToggle sets the toggled state of the specified menu item using
// the menu's name and the submenu's ID.
func setMenuItemToggle(menuName string, menuKey cmd.Key, toggle bool, nodraw ...struct{}) {
	menu.lock.Lock()
	defer menu.lock.Unlock()

	menuItem := menu.options[menuName][menuKey]
	if menuItem == nil || menuItem.KeyData == nil {
		return
	}

	title := menuItem.Title
	switch {
	case menuItem.Disabled == "":
		menuItem.Display = title
		return

	case menuItem.Enabled == "" && menuItem.Disabled != "":
		if toggle {
			menuItem.Display = menuItem.Disabled
		} else {
			menuItem.Display = title
		}

		menuItem.Toggled = toggle

		return
	}

	if toggle {
		menuItem.Display = title + " " + menuItem.Disabled
	} else {
		menuItem.Display = title + " " + menuItem.Enabled
	}

	menuItem.Toggled = toggle

	if nodraw == nil {
		UI.QueueUpdateDraw(func() {
			highlighted := menu.bar.GetHighlights()

			if menu.modal.Open && highlighted != nil && highlighted[0] == menuName {
				cell := menu.modal.Table.GetCell(menuItem.index, 0)
				if cell == nil {
					return
				}

				cell.Text = menuItem.Display
			}
		})
	}
}

// setMenu sets up a submenu for the specified menu.
func setMenu(x, y int, menuID string, device ...struct{}) {
	var width, skipped int

	modal := menu.modal
	modal.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch cmd.KeyOperation(event) {
		case cmd.KeyClose:
			exitMenu()
			return event

		case cmd.KeySwitch:
			switchMenu()
			return event
		}

		menuInputHandler(event)

		return event
	})
	modal.Table.SetSelectedFunc(func(row, col int) {
		cell := menu.modal.Table.GetCell(row, 0)
		if cell == nil {
			return
		}

		ref, ok := cell.GetReference().(*MenuOption)
		if !ok {
			return
		}

		KeyHandler(ref.Key, FunctionClick)()
	})

	modal.Table.Clear()
	for index, menuopt := range menuKeybindings[menuID] {
		if menuopt.Visible && !KeyHandler(menuopt.Key, FunctionVisible)() {
			skipped++
			continue
		}

		toggle := menuopt.Toggled
		if toggleHandler := KeyHandler(menuopt.Key, FunctionCreate); menuopt.OnCreate && toggleHandler != nil {
			toggle = toggleHandler()
		}

		setMenuItemToggle(menuID, menuopt.Key, toggle, struct{}{})

		display := menuopt.Display
		keybinding := cmd.KeyName(menuopt.KeyData.Kb)

		displayWidth := len(display) + len(keybinding) + 6
		if displayWidth > width {
			width = displayWidth
		}

		modal.Table.SetCell(index-skipped, 0, tview.NewTableCell(display).
			SetExpansion(1).
			SetReference(menuopt).
			SetAlign(tview.AlignLeft).
			SetTextColor(theme.GetColor(theme.ThemeMenuItem)).
			SetClickedFunc(KeyHandler(menuopt.Key, FunctionClick)).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor(theme.ThemeMenuItem)).
				Background(theme.BackgroundColor(theme.ThemeMenuItem)),
			),
		)
		modal.Table.SetCell(index-skipped, 1, tview.NewTableCell(keybinding).
			SetExpansion(1).
			SetAlign(tview.AlignRight).
			SetClickedFunc(KeyHandler(menuopt.Key, FunctionClick)).
			SetTextColor(theme.GetColor(theme.ThemeMenuItem)).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor(theme.ThemeMenuItem)).
				Background(theme.BackgroundColor(theme.ThemeMenuItem)),
			),
		)
	}

	modal.Table.Select(0, 0)

	drawMenuBox(x, y, width, device != nil)
}

// setContextMenu sets up a selector menu.
func setContextMenu(
	menuID string,
	selected func(table *tview.Table),
	changed func(table *tview.Table, row, col int),
	listContents func(table *tview.Table) (int, int),
) *tview.Table {
	var changeEnabled bool

	x, y := 0, 1

	modal := menu.modal
	modal.Table.Clear()
	modal.Table.SetSelectorWrap(false)
	modal.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch cmd.KeyOperation(event) {
		case cmd.KeySelect:
			if selected != nil {
				selected(modal.Table)
			}

			fallthrough

		case cmd.KeyClose:
			exitMenu()
		}

		return event
	})
	modal.Table.SetSelectionChangedFunc(func(row, col int) {
		if changed == nil {
			return
		}

		if !changeEnabled {
			changeEnabled = true
			return
		}

		changed(modal.Table, row, col)
	})
	modal.Table.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if action == tview.MouseLeftClick && modal.Table.InRect(event.Position()) {
			if selected != nil {
				modal.Table.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone), nil)
			}
		}

		return action, event
	})

	modal.Name = menuID
	width, index := listContents(modal.Table)

	exitMenu(struct{}{})
	modal.Table.Select(index, 0)
	drawMenuBox(x, y, width+20, menuID == "device")

	return modal.Table
}

// drawMenuBox draws the submenu.
func drawMenuBox(x, y, width int, device bool) {
	if menu.modal.Open {
		exitMenu(struct{}{})
	}

	if device {
		_, _, _, tableHeight := DeviceTable.GetInnerRect()
		deviceX, deviceY := getSelectionXY(DeviceTable)

		x = deviceX + 10
		if deviceY >= tableHeight-6 {
			y = deviceY - menu.modal.Table.GetRowCount()
		} else {
			y = deviceY + 1
		}
	}

	menu.modal.Height = menu.modal.Table.GetRowCount() + 2
	menu.modal.Width = width

	menu.modal.regionX = x
	menu.modal.regionY = y

	menu.modal.Show()
}

// setMenuBarHeader appends the header text with
// the menu bar's regions.
func setMenuBarHeader(header string) {
	menu.bar.SetText(header + "[-:-:-] " + theme.ColorWrap(theme.ThemeMenu, menuBarRegions))
}

// switchMenu switches between menus.
func switchMenu() {
	highlighted := menu.bar.GetHighlights()
	if highlighted == nil {
		return
	}

	for _, region := range menu.bar.GetRegionInfos() {
		if highlighted[0] != region.ID && highlighted[0] != "adapterchange" {
			menu.bar.Highlight(region.ID)
		}
	}
}

// exitMenu exits the menu.
func exitMenu(highlight ...struct{}) {
	menu.modal.Exit(false)

	if highlight == nil {
		menu.modal.Name = "menu"
		menu.bar.Highlight("")
	}

	UI.SetFocus(DeviceTable)
}

// menuInputHandler handles key events for a submenu.
func menuInputHandler(event *tcell.EventKey) {
	menu.lock.Lock()
	defer menu.lock.Unlock()

	if menu.options == nil {
		return
	}

	key := cmd.KeyOperation(event, cmd.KeyContextProgress)

	for _, options := range menu.options {
		for menuKey, option := range options {
			if menuKey == key {
				if option.Visible && !KeyHandler(menuKey, FunctionVisible)() {
					return
				}

				KeyHandler(menuKey, FunctionClick)()

				return
			}
		}
	}
}
