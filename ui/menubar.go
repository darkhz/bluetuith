package ui

import (
	"sort"
	"sync"

	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// Menu describes a region to display menu items.
type Menu struct {
	bar   *tview.TextView
	modal *Modal

	options map[string]map[string]*MenuOption

	lock sync.Mutex
}

// MenuOption describes an option layout for a submenu.
type MenuOption struct {
	index      int
	title      string
	menuid     string
	togglestr  string
	displaystr string
	keybinding rune
	toggled    bool
	oncreate   func() bool
	onclick    func() bool
	visible    func() bool
}

const menuBarRegions = `["adapter"][::b][Adapter[][""] ["device"][::b][Device[][""]`

var menu Menu

// menuBar sets up and returns the menu bar.
func menuBar() *tview.TextView {
	setupMenuOptions()

	menu.bar = tview.NewTextView()
	menu.bar.SetRegions(true)
	menu.bar.SetDynamicColors(true)
	menu.bar.SetBackgroundColor(theme.GetColor("MenuBar"))
	menu.bar.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		for _, region := range menu.bar.GetRegionInfos() {
			if region.ID == added[0] {
				setMenu(region.FromX, 1, added[0], menu.options[added[0]])
				break
			}
		}
	})

	menu.modal = NewMenuModal("menu", 0, 0)

	return menu.bar
}

// setupMenuOptions sets up the menu options with its attributes.
func setupMenuOptions() {
	menu.options = make(map[string]map[string]*MenuOption)

	adapterOptions := []*MenuOption{
		{
			index:      0,
			title:      "Power on",
			menuid:     "power",
			togglestr:  "Power off",
			keybinding: 'o',
			onclick:    onClickFunc("power"),
			oncreate:   createPower,
		},
		{
			index:      1,
			title:      "Discovery on",
			menuid:     "discoverable",
			togglestr:  "Discovery off",
			keybinding: 'S',
			onclick:    onClickFunc("discoverable"),
			oncreate:   createDiscoverable,
		},
		{
			index:      2,
			title:      "Pairable on",
			menuid:     "pairable",
			togglestr:  "Pairable off",
			keybinding: 'P',
			onclick:    onClickFunc("pairable"),
			oncreate:   createPairable,
		},
		{
			index:      3,
			title:      "Scan",
			menuid:     "scan",
			togglestr:  "Stop scan",
			keybinding: 's',
			onclick:    onClickFunc("scan"),
		},

		{
			index:      4,
			title:      "Change",
			menuid:     "change",
			keybinding: 'a',
			onclick:    onClickFunc("change"),
		},
		{
			index:      5,
			title:      "Progress View",
			menuid:     "progress",
			keybinding: 'v',
			onclick:    onClickFunc("progress"),
		},
		{
			index:      6,
			title:      "Hide player",
			menuid:     "playerhide",
			keybinding: 'M',
			onclick:    onClickFunc("hideplayer"),
		},
		{
			index:      7,
			title:      "Quit",
			menuid:     "quit",
			keybinding: 'Q',
			onclick:    onClickFunc("quit"),
		},
	}

	deviceOptions := []*MenuOption{
		{
			index:      0,
			title:      "Connect",
			menuid:     "connect",
			togglestr:  "Disconnect",
			keybinding: 'c',
			onclick:    onClickFunc("connect"),
			oncreate:   createConnect,
		},
		{
			index:      1,
			title:      "Pair",
			menuid:     "pair",
			keybinding: 'p',
			onclick:    onClickFunc("pair"),
		},
		{
			index:      2,
			title:      "Trust",
			menuid:     "trust",
			togglestr:  "Untrust",
			keybinding: 't',
			onclick:    onClickFunc("trust"),
			oncreate:   createTrust,
		},
		{
			index:      3,
			title:      "Send Files",
			menuid:     "send",
			keybinding: 'f',
			onclick:    onClickFunc("send"),
			visible:    visibleSend,
		},
		{
			index:      4,
			title:      "Network",
			menuid:     "network",
			keybinding: 'n',
			onclick:    onClickFunc("network"),
			visible:    visibleNetwork,
		},
		{
			index:      5,
			title:      "Audio Profiles",
			menuid:     "profiles",
			keybinding: 'A',
			onclick:    onClickFunc("profiles"),
			visible:    visibleProfile,
		},
		{
			index:      6,
			title:      "Media player",
			menuid:     "player",
			keybinding: 'm',
			onclick:    onClickFunc("showplayer"),
			visible:    visiblePlayer,
		},
		{
			index:      7,
			title:      "Info",
			menuid:     "info",
			keybinding: 'i',
			onclick:    onClickFunc("info"),
		},
		{
			index:      8,
			title:      "Remove",
			menuid:     "remove",
			keybinding: 'd',
			onclick:    onClickFunc("remove"),
		},
	}

	for _, opt := range []string{"adapter", "device"} {
		if menu.options[opt] == nil {
			menu.options[opt] = make(map[string]*MenuOption)
		}

		switch opt {
		case "adapter":
			for _, menuopt := range adapterOptions {
				menuopt.displaystr = menuopt.title
				menu.options[opt][menuopt.menuid] = menuopt
			}

		case "device":
			for _, menuopt := range deviceOptions {
				menuopt.displaystr = menuopt.title
				menu.options[opt][menuopt.menuid] = menuopt
			}
		}
	}
}

// setMenuItemToggle sets the toggled state of the specified menu item using
// the menu's name and the submenu's ID.
func setMenuItemToggle(menuName, menuID string, toggle bool, nodraw ...struct{}) {
	menu.lock.Lock()
	defer menu.lock.Unlock()

	menuItem := menu.options[menuName][menuID]
	if menuItem.togglestr == "" {
		return
	}

	if toggle {
		menuItem.displaystr = menuItem.togglestr
	} else {
		menuItem.displaystr = menuItem.title
	}
	menuItem.toggled = toggle

	if nodraw == nil {
		UI.QueueUpdateDraw(func() {
			highlighted := menu.bar.GetHighlights()

			if menu.modal.Open && highlighted != nil && highlighted[0] == menuName {
				cell := menu.modal.Table.GetCell(menuItem.index, 0)
				if cell == nil {
					return
				}

				cell.Text = menuItem.displaystr
			}
		})
	}
}

// setMenu sets up a submenu for the specified menu.
func setMenu(x, y int, menuID string, options map[string]*MenuOption, device ...struct{}) {
	var menuoptions []*MenuOption

	modal := menu.modal
	modal.Table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			exitMenu()
			return event

		case tcell.KeyTab:
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

		ref.onclick()
	})

	for _, opt := range options {
		if opt.visible != nil && !opt.visible() {
			continue
		}

		menuoptions = append(menuoptions, opt)
	}
	sort.Slice(menuoptions, func(i, j int) bool {
		return menuoptions[i].index < menuoptions[j].index
	})

	modal.Table.Clear()
	for index, menuopt := range menuoptions {
		if menuopt.oncreate != nil {
			setMenuItemToggle(menuID, menuopt.menuid, menuopt.oncreate(), struct{}{})
		}

		modal.Table.SetCell(index, 0, tview.NewTableCell(menuopt.displaystr).
			SetExpansion(1).
			SetReference(menuopt).
			SetAlign(tview.AlignLeft).
			SetClickedFunc(menuopt.onclick).
			SetTextColor(theme.GetColor("MenuItem")).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor("MenuItem")).
				Background(theme.BackgroundColor("MenuItem")),
			),
		)
		modal.Table.SetCell(index, 1, tview.NewTableCell(string(menuopt.keybinding)).
			SetExpansion(1).
			SetAlign(tview.AlignRight).
			SetClickedFunc(menuopt.onclick).
			SetTextColor(theme.GetColor("MenuItem")).
			SetSelectedStyle(tcell.Style{}.
				Foreground(theme.GetColor("MenuItem")).
				Background(theme.BackgroundColor("MenuItem")),
			),
		)
	}

	modal.Table.Select(0, 0)

	drawMenuBox(x, y, 20, device != nil)
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
		switch event.Key() {
		case tcell.KeyEnter:
			if selected != nil {
				selected(modal.Table)
			}

			fallthrough

		case tcell.KeyEscape:
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
	menu.bar.SetText(header + "[-:-:-] " + theme.ColorWrap("Menu", menuBarRegions))
}

// switchMenu switches between menus.
func switchMenu() {
	highlighted := menu.bar.GetHighlights()
	if highlighted == nil {
		return
	}

	for _, region := range menu.bar.GetRegionInfos() {
		if highlighted[0] != region.ID {
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

	key := cmd.KeyOperation(event, cmd.KeyContextDevice, cmd.KeyContextAdapter)

	for _, options := range menu.options {
		for menuKey, option := range options {
			if menuKey == key {
				if option.Visible && !KeyHandler(menuKey, FunctionVisible)() {
					return
				}

				opt.onclick()

				return
			}
		}
	}
}
