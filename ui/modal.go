package ui

import "github.com/darkhz/tview"

func ShowModal(page string, table *tview.Table, proportions ...int) {
	prevPage, _ := Pages.GetFrontPage()

	height := proportions[0]
	width := proportions[1]

	tableWrap := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 20, false).
		AddItem(table, 0, height, true).
		AddItem(nil, 0, 20, false)

	tableFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(nil, 0, width, false).
		AddItem(tableWrap, 0, 60, true).
		AddItem(nil, 0, width, false)

	Pages.AddAndSwitchToPage(page, tableFlex, true).ShowPage(prevPage)

	App.SetFocus(table)
}
