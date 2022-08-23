package ui

import (
	"sort"
	"sync"

	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

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

var (
	// MenuBar holds the menu bar.
	MenuBar *tview.TextView

	menuList       *tview.Table
	menuOptionLock sync.Mutex
	menuOptions    map[string]map[string]*MenuOption
)

const menuBarRegions = `["adapter"][::b][Adapter[][""] ["device"][::b][Device[][""]`

// menuBar sets up and returns the menu bar.
func menuBar() *tview.TextView {
	setupMenuOptions()

	MenuBar = tview.NewTextView()
	MenuBar.SetRegions(true)
	MenuBar.SetDynamicColors(true)
	MenuBar.SetBackgroundColor(theme.GetColor("MenuBar"))
	MenuBar.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		for _, region := range MenuBar.GetRegionInfos() {
			if region.ID == added[0] {
				setMenuList(region.FromX, 1, added[0], menuOptions[added[0]])
				break
			}
		}
	})

	return MenuBar
}

// setupMenuOptions sets up the menu options with its attributes.
func setupMenuOptions() {
	menuOptions = make(map[string]map[string]*MenuOption)

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
			title:      "Scan",
			menuid:     "scan",
			togglestr:  "Stop scan",
			keybinding: 's',
			onclick:    onClickFunc("scan"),
		},
		{
			index:      2,
			title:      "Change",
			menuid:     "change",
			keybinding: 'a',
			onclick:    onClickFunc("change"),
		},
		{
			index:      3,
			title:      "Progress View",
			menuid:     "progress",
			keybinding: 'v',
			onclick:    onClickFunc("progress"),
		},
		{
			index:      4,
			title:      "Quit",
			menuid:     "quit",
			keybinding: 'q',
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
			title:      "Info",
			menuid:     "info",
			keybinding: 'i',
			onclick:    onClickFunc("info"),
		},
		{
			index:      7,
			title:      "Remove",
			menuid:     "remove",
			keybinding: 'd',
			onclick:    onClickFunc("remove"),
		},
	}

	for _, opt := range []string{"adapter", "device"} {
		if menuOptions[opt] == nil {
			menuOptions[opt] = make(map[string]*MenuOption)
		}

		switch opt {
		case "adapter":
			for _, menuopt := range adapterOptions {
				menuopt.displaystr = menuopt.title
				menuOptions[opt][menuopt.menuid] = menuopt
			}

		case "device":
			for _, menuopt := range deviceOptions {
				menuopt.displaystr = menuopt.title
				menuOptions[opt][menuopt.menuid] = menuopt
			}
		}
	}
}

// setMenuItemToggle sets the toggled state of the specified menu item using
// the menu's name and the submenu's ID.
func setMenuItemToggle(menuName, menuID string, toggle bool, nodraw ...struct{}) {
	menuOptionLock.Lock()
	defer menuOptionLock.Unlock()

	menuItem := menuOptions[menuName][menuID]
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
		App.QueueUpdateDraw(func() {
			highlighted := MenuBar.GetHighlights()

			if Pages.HasPage("menulist") && highlighted != nil && highlighted[0] == menuName {
				cell := menuList.GetCell(menuItem.index, 0)
				if cell == nil {
					return
				}

				cell.Text = menuItem.displaystr
			}
		})
	}
}

// setMenuList sets up a submenu for the specified menu.
func setMenuList(x, y int, menu string, options map[string]*MenuOption, selection ...struct{}) {
	if menuList != nil {
		goto AddOptions
	}

	menuList = tview.NewTable()
	menuList.SetBorder(true)
	menuList.SetEnableFocus(false)
	menuList.SetSelectable(true, false)
	menuList.SetBorderColor(theme.GetColor("Border"))
	menuList.SetBackgroundColor(theme.GetColor("Background"))
	menuList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			exitMenu("menulist")
			return event

		case tcell.KeyTab:
			switchMenuList()
			return event
		}

		menuListInputHandler(event)

		return event
	})
	menuList.SetSelectedFunc(func(row, col int) {
		cell := menuList.GetCell(row, 0)
		if cell == nil {
			return
		}

		ref, ok := cell.GetReference().(*MenuOption)
		if !ok {
			return
		}

		ref.onclick()
	})

AddOptions:
	var menuoptions []*MenuOption

	for _, opt := range options {
		if opt.visible != nil && !opt.visible() {
			continue
		}

		menuoptions = append(menuoptions, opt)
	}
	sort.Slice(menuoptions, func(i, j int) bool {
		return menuoptions[i].index < menuoptions[j].index
	})

	menuList.Clear()
	for index, menuopt := range menuoptions {
		if menuopt.oncreate != nil {
			setMenuItemToggle(menu, menuopt.menuid, menuopt.oncreate(), struct{}{})
		}

		menuList.SetCell(index, 0, tview.NewTableCell(menuopt.displaystr).
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
		menuList.SetCell(index, 1, tview.NewTableCell(string(menuopt.keybinding)).
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
	menuList.Select(0, 0)

	Pages.AddAndSwitchToPage(
		"menulist",
		drawMenuBox(menuList, menuList.GetRowCount()+2, 20, x, y, selection...),
		true,
	).ShowPage("main")

	App.SetFocus(menuList)
}

// setSelectorMenu sets up a selector menu.
func setSelectorMenu(
	menuID string,
	selected func(table *tview.Table),
	selectchg func(table *tview.Table, row, col int),
	listContents func(table *tview.Table) (int, int),
) *tview.Table {
	var x, y int
	var selchgEnable bool
	var selection []struct{}
	var selectorMenu *tview.Table

	if menuID == "adapter" {
	}

	switch menuID {
	case "adapter":
		x, y = 0, 1

	case "device":
		selection = append(selection, struct{}{})
	}

	exitMenu("menulist")

	selectorMenu = tview.NewTable()
	selectorMenu.SetBorder(true)
	selectorMenu.SetEnableFocus(false)
	selectorMenu.SetSelectable(true, false)
	selectorMenu.SetBorderColor(theme.GetColor("Border"))
	selectorMenu.SetBackgroundColor(theme.GetColor("Background"))
	selectorMenu.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if selected != nil {
				selected(selectorMenu)
			}

			fallthrough

		case tcell.KeyEscape:
			exitMenu("selectormenu")
		}

		return event
	})
	selectorMenu.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		x, y := event.Position()

		if action == tview.MouseLeftClick && selectorMenu.InRect(x, y) {
			if selected != nil {
				selectorMenu.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone), nil)
			}
		}

		return action, event
	})
	selectorMenu.SetSelectionChangedFunc(func(row, col int) {
		if selectchg == nil {
			return
		}

		if !selchgEnable {
			selchgEnable = true
			return
		}

		selectchg(selectorMenu, row, col)
	})

	width, index := listContents(selectorMenu)

	selectorMenu.Select(index, 0)

	Pages.AddAndSwitchToPage(
		"selectormenu",
		drawMenuBox(selectorMenu, selectorMenu.GetRowCount()+2, width+20, x, y, selection...),
		true,
	).ShowPage("main")

	App.SetFocus(selectorMenu)

	return selectorMenu
}

// drawMenuBox draws the submenu.
func drawMenuBox(list *tview.Table, height, width, x, y int, selection ...struct{}) *tview.Flex {
	if selection != nil {
		_, _, _, tableHeight := DeviceTable.GetInnerRect()
		deviceX, deviceY := getSelectionXY(DeviceTable)

		x = deviceX + 10
		if deviceY >= tableHeight-6 {
			y = deviceY - list.GetRowCount()
		} else {
			y = deviceY + 1
		}
	}

	wrapList := tview.NewFlex().
		AddItem(nil, y, 0, false).
		AddItem(list, height, 0, true).
		AddItem(nil, 1, 0, false).
		SetDirection(tview.FlexRow)

	return tview.NewFlex().
		AddItem(nil, x, 0, false).
		AddItem(wrapList, width, 0, true).
		AddItem(nil, 1, 0, false).
		SetDirection(tview.FlexColumn)
}

// setMenuBarHeader appends the header text with
// the menu bar's regions.
func setMenuBarHeader(header string) {
	MenuBar.SetText(header + "[-:-:-] " + theme.ColorWrap("Menu", menuBarRegions))
}

// switchMenuList switches between menus.
func switchMenuList() {
	highlighted := MenuBar.GetHighlights()
	if highlighted == nil {
		return
	}

	for _, region := range MenuBar.GetRegionInfos() {
		if highlighted[0] != region.ID {
			exitMenu("menulist")
			MenuBar.Highlight(region.ID)
		}
	}
}

// exitMenuList closes the submenu and exits the menubar.
func exitMenu(menu string) {
	MenuBar.Highlight("")
	Pages.RemovePage(menu)

	App.SetFocus(DeviceTable)
}

// menuListInputHandler handles key events for a submenu.
func menuListInputHandler(event *tcell.EventKey) {
	menuOptionLock.Lock()
	defer menuOptionLock.Unlock()

	if menuOptions == nil {
		return
	}

	for _, option := range menuOptions {
		for _, opt := range option {
			if event.Rune() == opt.keybinding {
				if opt.visible != nil && !opt.visible() {
					return
				}

				opt.onclick()

				return
			}
		}
	}
}

//gocyclo: ignore
// menuListMouseHandler handles mouse events for a submenu.
func menuListMouseHandler(action tview.MouseAction, event *tcell.EventMouse) *tcell.EventMouse {
	if !Pages.HasPage("menulist") && !Pages.HasPage("selectormenu") {
		return event
	}

	if action != tview.MouseLeftClick {
		return event
	}

	x, y := event.Position()

	switch {
	case Pages.HasPage("selectormenu"):
		pg, item := Pages.GetFrontPage()
		if pg != "selectormenu" {
			return event
		}

		flex, ok := item.(*tview.Flex)
		if !ok {
			return event
		}

		for i := 0; i < flex.GetItemCount(); i++ {
			tflex, ok := flex.GetItem(i).(*tview.Flex)
			if !ok {
				continue
			}

			for j := 0; j < tflex.GetItemCount(); j++ {
				table, ok := tflex.GetItem(j).(*tview.Table)
				if !ok {
					continue
				}

				if !table.InRect(x, y) {
					exitMenu("selectormenu")
					break
				} else {
					panic(nil)
				}
			}
		}

	case menuList != nil && menuList.InRect(x, y):
		App.SetFocus(menuList)
		return nil

	case !MenuBar.InRect(x, y) && menuList != nil && !menuList.InRect(x, y):
		exitMenu("menulist")
	}

	return event
}
