package ui

import (
	"sync"

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
	MenuBar.SetBackgroundColor(tcell.ColorDefault)
	MenuBar.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		for _, region := range MenuBar.GetRegionInfos() {
			if region.ID == added[0] {
				setMenuList(region.FromX, region.ToY+2, added[0], menuOptions[added[0]])
				break
			}
		}
	})

	MenuBar.SetText(menuBarRegions)

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
		},
		{
			index:      4,
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
func setMenuList(x, y int, menu string, options map[string]*MenuOption) {
	if menuList != nil {
		goto AddOptions
	}

	menuList = tview.NewTable()
	menuList.SetBorder(true)
	menuList.SetSelectable(true, false)
	menuList.SetBackgroundColor(tcell.ColorDefault)
	menuList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			exitMenuList()
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

	menuList.Clear()
	for _, opt := range options {
		if opt.oncreate != nil {
			setMenuItemToggle(menu, opt.menuid, opt.oncreate(), struct{}{})
		}

		menuList.SetCell(opt.index, 0, tview.NewTableCell(opt.displaystr).
			SetExpansion(1).
			SetReference(opt).
			SetAlign(tview.AlignLeft).
			SetClickedFunc(opt.onclick),
		)
		menuList.SetCell(opt.index, 1, tview.NewTableCell(string(opt.keybinding)).
			SetExpansion(1).
			SetAlign(tview.AlignRight).
			SetClickedFunc(opt.onclick),
		)
	}
	menuList.Select(0, 0)

	Pages.AddAndSwitchToPage(
		"menulist",
		drawMenuBox(menuList, menuList.GetRowCount()+2, 20, x, y),
		true,
	).ShowPage("main")
	App.SetFocus(menuList)
}

// drawMenuBox draws the submenu.
func drawMenuBox(list tview.Primitive, height, width, x, y int) *tview.Flex {
	wrapList := tview.NewFlex().
		AddItem(nil, 1, 0, false).
		AddItem(list, height, 0, true).
		AddItem(nil, 1, 0, false).
		SetDirection(tview.FlexRow)

	return tview.NewFlex().
		AddItem(nil, x, 0, false).
		AddItem(wrapList, width, 0, true).
		AddItem(nil, y, 0, false).
		SetDirection(tview.FlexColumn)
}

// setMenuBarHeader appends the header text with
// the menu bar's regions.
func setMenuBarHeader(header string) {
	MenuBar.SetText(header + "[-:-:-] " + menuBarRegions)
}

// switchMenuList switches between menus.
func switchMenuList() {
	highlighted := MenuBar.GetHighlights()
	if highlighted == nil {
		return
	}

	for _, region := range MenuBar.GetRegionInfos() {
		if highlighted[0] != region.ID {
			exitMenuList()
			MenuBar.Highlight(region.ID)
		}
	}
}

// exitMenuList closes the submenu and exits the menubar.
func exitMenuList() {
	MenuBar.Highlight("")
	Pages.RemovePage("menulist")

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
				opt.onclick()
				return
			}
		}
	}
}

//gocyclo: ignore
// menuListMouseHandler handles mouse events for a submenu.
func menuListMouseHandler(action tview.MouseAction, event *tcell.EventMouse) *tcell.EventMouse {
	if !Pages.HasPage("menulist") && !Pages.HasPage("adaptermenu") {
		return event
	}

	x, y := event.Position()

	switch {
	case Pages.HasPage("adaptermenu"):
		pg, item := Pages.GetFrontPage()
		if pg != "adaptermenu" {
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
					Pages.RemovePage("adaptermenu")
					App.SetFocus(DeviceTable)

					break
				}
			}
		}

	case menuList != nil && menuList.InRect(x, y):
		App.SetFocus(menuList)
		return nil

	case !MenuBar.InRect(x, y) && menuList != nil && !menuList.InRect(x, y):
		exitMenuList()
	}

	return event
}
