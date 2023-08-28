package ui

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// FilePicker describes a filepicker.
type FilePicker struct {
	table          *tview.Table
	title, buttons *tview.TextView

	prevDir, currentPath string
	isHidden             bool

	listChan      chan []string
	prevFileInfo  fs.DirEntry
	selectedFiles map[string]fs.DirEntry

	lock, selection, hide sync.Mutex
}

const filePickButtonRegion = `["ok"][::b][OK[][""] ["cancel"][::b][Cancel[][""] ["hidden"][::b][Toggle hidden[][""] ["invert"][Invert selection[][""] ["all"][Select All[][""]`

var filepicker FilePicker

// filePicker shows a file picker, and returns
// a list of all the selected files.
func filePicker() []string {
	UI.QueueUpdateDraw(func() {
		setupFilePicker()
	})

	return <-filepicker.listChan
}

// setupFilePicker sets up the file picker.
func setupFilePicker() {
	filepicker.listChan = make(chan []string)
	filepicker.selectedFiles = make(map[string]fs.DirEntry)

	infoTitle := tview.NewTextView()
	infoTitle.SetDynamicColors(true)
	infoTitle.SetTextAlign(tview.AlignCenter)
	infoTitle.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	infoTitle.SetText(theme.ColorWrap(theme.ThemeText, "Select files to send", "::bu"))

	pickerflex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(infoTitle, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(filePickerTitle(), 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(filePickerTable(), 0, 10, true).
		AddItem(nil, 1, 0, false).
		AddItem(filePickerButtons(), 2, 0, false)
	pickerflex.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	UI.Pages.AddAndSwitchToPage("filepicker", pickerflex, true)

	if !getHidden() {
		toggleHidden()
	}

	go changeDir(false, false)
}

// filePickerTable sets up and returns the filepicker table.
func filePickerTable() *tview.Table {
	filepicker.table = tview.NewTable()
	filepicker.table.SetSelectorWrap(true)
	filepicker.table.SetSelectable(true, false)
	filepicker.table.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	filepicker.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch cmd.KeyOperation(event, cmd.KeyContextFiles) {
		case cmd.KeyFilebrowserDirForward:
			go changeDir(true, false)

		case cmd.KeyFilebrowserDirBack:
			go changeDir(false, true)

		case cmd.KeyFilebrowserToggleHidden:
			toggleHidden()
			fallthrough

		case cmd.KeyFilebrowserRefresh:
			go changeDir(false, false)

		case cmd.KeyFilebrowserConfirmSelection:
			sendFileList()
			fallthrough

		case cmd.KeyClose:
			buttonHandler("cancel")

		case cmd.KeyQuit:
			go quit()

		case cmd.KeyHelp:
			showHelp()

		case cmd.KeyFilebrowserSelectAll, cmd.KeyFilebrowserInvertSelection, cmd.KeyFilebrowserSelect:
			selectFile(event.Rune())
		}

		return event
	})

	return filepicker.table
}

// filePickerTitle sets up and returns the filepicker title area.
// This will be used to show the current directory path.
func filePickerTitle() *tview.TextView {
	filepicker.title = tview.NewTextView()
	filepicker.title.SetDynamicColors(true)
	filepicker.title.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	filepicker.title.SetTextAlign(tview.AlignLeft)

	return filepicker.title
}

// filePickerButtons sets up and returns the filepicker buttons.
func filePickerButtons() *tview.TextView {
	filepicker.buttons = tview.NewTextView()
	filepicker.buttons.SetRegions(true)
	filepicker.buttons.SetDynamicColors(true)
	filepicker.buttons.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	filepicker.buttons.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		for _, region := range filepicker.buttons.GetRegionInfos() {
			if region.ID == added[0] {
				buttonHandler(added[0])
				break
			}
		}
	})

	filepicker.buttons.SetTextAlign(tview.AlignLeft)
	filepicker.buttons.SetText(theme.ColorWrap(theme.ThemeText, filePickButtonRegion))

	return filepicker.buttons
}

// sendFileList sends a slice of all the selected files
// to the file list channel, which is received by filePicker().
func sendFileList() {
	filepicker.selection.Lock()
	defer filepicker.selection.Unlock()

	var fileList []string

	for path := range filepicker.selectedFiles {
		fileList = append(fileList, path)
	}

	filepicker.listChan <- fileList
}

// selectFile sets the parameters for the file selection handler.
func selectFile(key rune) {
	var all, inverse bool

	switch key {
	case 'A':
		all = true
		inverse = false

	case 'a':
		all = false
		inverse = true

	case ' ':
		all = false
		inverse = false
	}

	selectFileHandler(all, inverse)
}

// changeDir changes to a directory and lists its contents.
func changeDir(cdFwd bool, cdBack bool) {
	var testPath string

	filepicker.lock.Lock()
	defer filepicker.lock.Unlock()

	if filepicker.currentPath == "" {
		var err error

		filepicker.currentPath, err = os.UserHomeDir()
		if err != nil {
			ErrorMessage(err)
			return
		}
	}

	testPath = filepicker.currentPath

	row, _ := filepicker.table.GetSelection()
	cell := filepicker.table.GetCell(row, 1)
	if cell == nil {
		return
	}
	if cdFwd && tview.Escape(cell.Text) == "../" {
		cdFwd = false
		cdBack = true
	}

	switch {
	case cdFwd:
		entry, ok := cell.GetReference().(fs.DirEntry)
		if !ok {
			return
		}

		testPath = trimPath(testPath, false)
		testPath = filepath.Join(testPath, entry.Name())

	case cdBack:
		filepicker.prevDir = filepath.Base(testPath)
		testPath = trimPath(testPath, cdBack)
	}

	dlist, listed := dirList(filepath.FromSlash(testPath))
	if !listed {
		return
	}

	sort.Slice(dlist, func(i, j int) bool {
		if dlist[i].IsDir() != dlist[j].IsDir() {
			return dlist[i].IsDir()
		}

		return dlist[i].Name() < dlist[j].Name()
	})

	filepicker.currentPath = testPath

	createDirList(dlist, cdBack)
}

// createDirList displays the contents of the directory in the filepicker.
func createDirList(dlist []fs.DirEntry, cdBack bool) {
	UI.QueueUpdateDraw(func() {
		var pos int
		var prevrow int = -1
		var rowpadding int = -1

		filepicker.table.SetSelectable(false, false)
		filepicker.table.Clear()

		for row, entry := range dlist {
			var attr tcell.AttrMask
			var entryColor tcell.Color

			info, err := entry.Info()
			if err != nil {
				continue
			}

			name := info.Name()
			fileTotalSize := formatSize(info.Size())
			fileModifiedTime := info.ModTime().Format("02 Jan 2006 03:04 PM")
			permissions := strings.ToLower(entry.Type().String())
			if len(permissions) > 10 {
				permissions = permissions[1:]
			}

			if entry.IsDir() {
				if filepicker.currentPath != "/" {
					rowpadding = 0

					if cdBack && name == filepicker.prevDir {
						pos = row
					}

					if entry == filepicker.prevFileInfo {
						name = ".."
						prevrow = row

						row = 0
						filepicker.table.InsertRow(0)

						if filepicker.table.GetRowCount() > 0 {
							pos++
						}
					}
				} else if filepicker.currentPath == "/" && name == "/" {
					rowpadding = -1
					continue
				}

				attr = tcell.AttrBold
				entryColor = tcell.ColorBlue
				name += string(os.PathSeparator)
			} else {
				entryColor = theme.GetColor(theme.ThemeText)
			}

			filepicker.table.SetCell(row+rowpadding, 0, tview.NewTableCell(" ").
				SetSelectable(false),
			)

			filepicker.table.SetCell(row+rowpadding, 1, tview.NewTableCell(tview.Escape(name)).
				SetExpansion(1).
				SetReference(entry).
				SetAttributes(attr).
				SetTextColor(entryColor).
				SetAlign(tview.AlignLeft).
				SetOnClickedFunc(cellHandler).
				SetSelectedStyle(tcell.Style{}.
					Bold(true).
					Foreground(entryColor).
					Background(theme.BackgroundColor(theme.ThemeText)),
				),
			)

			for col, text := range []string{
				permissions,
				fileTotalSize,
				fileModifiedTime,
			} {
				filepicker.table.SetCell(row+rowpadding, col+2, tview.NewTableCell(text).
					SetAlign(tview.AlignRight).
					SetTextColor(tcell.ColorGrey).
					SetSelectedStyle(tcell.Style{}.
						Bold(true),
					),
				)
			}

			if prevrow > -1 {
				row = prevrow
				prevrow = -1
			}

			markFileSelection(row, entry, checkFileSelected(filepath.Join(filepicker.currentPath, name)))
		}

		filepicker.title.SetText(theme.ColorWrap(theme.ThemeText, "Directory: "+filepicker.currentPath))

		filepicker.table.ScrollToBeginning()
		filepicker.table.SetSelectable(true, false)
		filepicker.table.Select(pos, 0)
	})
}

// dirList lists a directory's contents.
func dirList(testPath string) ([]fs.DirEntry, bool) {
	var dlist []fs.DirEntry

	_, err := os.Lstat(testPath)
	if err != nil {
		return nil, false
	}

	dir, err := os.Lstat(trimPath(testPath, true))
	if err != nil {
		return nil, false
	}

	list, err := os.ReadDir(testPath)
	if err != nil {
		return nil, false
	}

	dirEntry := fs.FileInfoToDirEntry(dir)

	filepicker.prevFileInfo = dirEntry
	dlist = append(dlist, dirEntry)

	for _, entry := range list {
		if getHidden() && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		dlist = append(dlist, entry)
	}

	return dlist, true
}

// cellHandler handles on-click events for a table cell.
func cellHandler(table *tview.Table, row, col int) {
	selectFileHandler(false, false, row)
}

// buttonHandler handles button on-click events for the filepicker buttons.
func buttonHandler(button string) {
	switch button {
	case "ok":
		sendFileList()
		fallthrough

	case "cancel":
		close(filepicker.listChan)

		UI.Pages.RemovePage("filepicker")
		UI.Pages.SwitchToPage("main")

	case "hidden":
		toggleHidden()
		go changeDir(false, false)

	case "invert":
		selectFileHandler(false, true)

	case "all":
		selectFileHandler(true, false)
	}

	filepicker.buttons.Highlight("")
}

// selectFileHandler iterates over the filepicker.table's rows,
// determines the type of selection to be made (single, inverse or all),
// and marks the selections.
func selectFileHandler(all, inverse bool, row ...int) {
	var pos int

	singleSelection := !all && !inverse
	inverseSelection := !all && inverse

	if row != nil {
		pos = row[0]
	} else {
		pos, _ = filepicker.table.GetSelection()
	}
	totalrows := filepicker.table.GetRowCount()

	for i := 0; i < totalrows; i++ {
		var checkSelected bool
		var userSelected []struct{}

		if singleSelection {
			i = pos
			userSelected = append(userSelected, struct{}{})
		}

		cell := filepicker.table.GetCell(i, 1)
		if cell == nil {
			return
		}

		entry, ok := cell.GetReference().(fs.DirEntry)
		if !ok {
			return
		}

		fullpath := filepath.Join(filepicker.currentPath, entry.Name())
		if singleSelection || inverseSelection {
			checkSelected = checkFileSelected(fullpath)
		}
		if !checkSelected {
			addFileSelection(fullpath, entry)
		} else {
			removeFileSelection(fullpath)
		}

		markFileSelection(i, entry, !checkSelected, userSelected...)

		if singleSelection {
			if i+1 < totalrows {
				filepicker.table.Select(i+1, 0)
				return
			}

			break
		}
	}

	filepicker.table.Select(pos, 0)
}

// addFileSelection adds a file to the filepicker.selectedFiles list.
func addFileSelection(path string, info fs.DirEntry) {
	filepicker.selection.Lock()
	defer filepicker.selection.Unlock()

	if !info.Type().IsRegular() {
		return
	}

	filepicker.selectedFiles[path] = info
}

// removeFileSelection removes a file from the filepicker.selectedFiles list.
func removeFileSelection(path string) {
	filepicker.selection.Lock()
	defer filepicker.selection.Unlock()

	delete(filepicker.selectedFiles, path)
}

// checkFileSelected checks if a file is selected.
func checkFileSelected(path string) bool {
	filepicker.selection.Lock()
	defer filepicker.selection.Unlock()

	_, selected := filepicker.selectedFiles[path]

	return selected
}

// markFileSelection marks the selection for files only, directories are skipped.
func markFileSelection(row int, info fs.DirEntry, selected bool, userSelected ...struct{}) {
	if !info.Type().IsRegular() {
		if info.IsDir() && userSelected != nil {
			go changeDir(true, false)
		}

		return
	}

	cell := filepicker.table.GetCell(row, 0)

	if selected {
		cell.Text = "+"
	} else {
		cell.Text = " "
	}

	cell.Text = theme.ColorWrap(theme.ThemeText, cell.Text)
}

// trimPath trims a given path and appends a path separator where appropriate.
func trimPath(testPath string, cdBack bool) string {
	testPath = filepath.Clean(testPath)

	if cdBack {
		testPath = filepath.Dir(testPath)
	}

	return filepath.FromSlash(testPath)
}

// getHidden checks if hidden files can be shown or not.
func getHidden() bool {
	filepicker.hide.Lock()
	defer filepicker.hide.Unlock()

	return filepicker.isHidden
}

// toggleHidden toggles the hidden files mode.
func toggleHidden() {
	filepicker.hide.Lock()
	defer filepicker.hide.Unlock()

	filepicker.isHidden = !filepicker.isHidden
}
