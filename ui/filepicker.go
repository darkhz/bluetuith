package ui

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

const filePickButtonRegion = `["ok"][::b][OK[][""] ["cancel"][::b][Cancel[][""] ["hidden"][::b][Toggle hidden[][""] ["invert"][Invert selection[][""] ["all"][Select All[][""]`

var (
	filePickTable   *tview.Table
	filePickTitle   *tview.TextView
	filePickButtons *tview.TextView
	filePickerLock  sync.Mutex

	prevDir      string
	currentPath  string
	prevFileInfo fs.FileInfo

	isHidden bool
	hideLock sync.Mutex

	fileSelection     map[string]fs.FileInfo
	fileSelectionLock sync.Mutex

	fileListChan chan []string
)

// filePicker shows a file picker, and returns
// a list of all the selected files.
func filePicker() []string {
	App.QueueUpdateDraw(func() {
		setupFilePicker()
	})

	return <-fileListChan
}

// setupFilePicker sets up the file picker.
func setupFilePicker() {
	fileListChan = make(chan []string)
	fileSelection = make(map[string]fs.FileInfo)

	infoTitle := tview.NewTextView()
	infoTitle.SetDynamicColors(true)
	infoTitle.SetTextAlign(tview.AlignCenter)
	infoTitle.SetText("[::bu]Select files to send")
	infoTitle.SetBackgroundColor(tcell.ColorDefault)

	pickerflex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(infoTitle, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(filePickerTitle(), 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(filePickerTable(), 0, 10, true).
		AddItem(nil, 1, 0, false).
		AddItem(filePickerButtons(), 2, 0, false)
	pickerflex.SetBackgroundColor(tcell.ColorDefault)

	Pages.AddAndSwitchToPage("filepicker", pickerflex, true)

	if !getHidden() {
		toggleHidden()
	}

	go changeDir(false, false)
}

// filePickerTable sets up and returns the filepicker table.
func filePickerTable() *tview.Table {
	filePickTable = tview.NewTable()
	filePickTable.SetSelectorWrap(true)
	filePickTable.SetSelectable(true, false)
	filePickTable.SetBackgroundColor(tcell.ColorDefault)
	filePickTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			go changeDir(false, true)

		case tcell.KeyRight:
			go changeDir(true, false)

		case tcell.KeyCtrlH:
			toggleHidden()
			fallthrough

		case tcell.KeyCtrlR:
			go changeDir(false, false)

		case tcell.KeyCtrlS:
			sendFileList()
			fallthrough

		case tcell.KeyEscape:
			close(fileListChan)
			Pages.RemovePage("filepicker")
		}

		switch event.Rune() {
		case 'q':
			go quit()

		case '?':
			showHelp()

		case 'A', 'a', ' ':
			selectFile(event.Rune())
		}

		return event
	})

	return filePickTable
}

// filePickerTitle sets up and returns the filepicker title area.
// This will be used to show the current directory path.
func filePickerTitle() *tview.TextView {
	filePickTitle = tview.NewTextView()
	filePickTitle.SetDynamicColors(true)
	filePickTitle.SetBackgroundColor(tcell.ColorDefault)
	filePickTitle.SetTextAlign(tview.AlignLeft)

	return filePickTitle
}

// filePickerButtons sets up and returns the filepicker buttons.
func filePickerButtons() *tview.TextView {
	filePickButtons = tview.NewTextView()
	filePickButtons.SetRegions(true)
	filePickButtons.SetDynamicColors(true)
	filePickButtons.SetBackgroundColor(tcell.ColorDefault)
	filePickButtons.SetHighlightedFunc(func(added, removed, remaining []string) {
		if added == nil {
			return
		}

		for _, region := range filePickButtons.GetRegionInfos() {
			if region.ID == added[0] {
				buttonHandler(added[0])
				break
			}
		}
	})

	filePickButtons.SetText(filePickButtonRegion)
	filePickButtons.SetTextAlign(tview.AlignLeft)

	return filePickButtons
}

// sendFileList sends a slice of all the selected files
// to the file list channel, which is received by filePicker().
func sendFileList() {
	fileSelectionLock.Lock()
	defer fileSelectionLock.Unlock()

	var fileList []string

	for path := range fileSelection {
		fileList = append(fileList, path)
	}

	fileListChan <- fileList
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

	filePickerLock.Lock()
	defer filePickerLock.Unlock()

	if currentPath == "" {
		var err error

		currentPath, err = os.UserHomeDir()
		if err != nil {
			ErrorMessage(err)
			return
		}
	}

	testPath = currentPath

	row, _ := filePickTable.GetSelection()
	cell := filePickTable.GetCell(row, 1)
	if cell == nil {
		return
	}
	if cdFwd && tview.Escape(cell.Text) == "../" {
		cdFwd = false
		cdBack = true
	}

	switch {
	case cdFwd:
		entry, ok := cell.GetReference().(fs.FileInfo)
		if !ok {
			return
		}

		testPath = trimPath(testPath, false)
		testPath = filepath.Join(testPath, entry.Name())

	case cdBack:
		prevDir = filepath.Base(testPath)
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

	currentPath = testPath

	createDirList(dlist, cdBack)
}

// createDirList displays the contents of the directory in the filepicker.
func createDirList(dlist []fs.FileInfo, cdBack bool) {
	App.QueueUpdateDraw(func() {
		var pos int
		var prevrow int = -1
		var rowpadding int = -1

		filePickTable.SetSelectable(false, false)
		filePickTable.Clear()

		for row, entry := range dlist {
			var color tcell.Color
			var attr tcell.AttrMask

			name := entry.Name()
			fileTotalSize := formatSize(entry.Size())
			fileModifiedTime := entry.ModTime().Format("02 Jan 2006 03:04 PM")
			permissions := strings.ToLower(entry.Mode().String())
			if len(permissions) > 10 {
				permissions = permissions[1:]
			}

			if entry.IsDir() {
				if currentPath != "/" {
					rowpadding = 0

					if cdBack && name == prevDir {
						pos = row
					}

					if entry == prevFileInfo {
						name = ".."
						prevrow = row

						row = 0
						filePickTable.InsertRow(0)

						if filePickTable.GetRowCount() > 0 {
							pos++
						}
					}
				} else if currentPath == "/" && name == "/" {
					rowpadding = -1
					continue
				}

				attr = tcell.AttrBold
				color = tcell.ColorBlue
				name += string(os.PathSeparator)
			} else {
				color = tcell.ColorWhite
			}

			filePickTable.SetCell(row+rowpadding, 0, tview.NewTableCell(" ").
				SetSelectable(false),
			)

			filePickTable.SetCell(row+rowpadding, 1, tview.NewTableCell(tview.Escape(name)).
				SetExpansion(1).
				SetReference(entry).
				SetTextColor(color).
				SetAttributes(attr).
				SetAlign(tview.AlignLeft).
				SetOnClickedFunc(cellHandler).
				SetSelectedStyle(tcell.Style{}.
					Foreground(color).
					Background(tcell.Color16).
					Attributes(tcell.AttrBold),
				),
			)

			for col, text := range []string{
				permissions,
				fileTotalSize,
				fileModifiedTime,
			} {
				filePickTable.SetCell(row+rowpadding, col+2, tview.NewTableCell(text).
					SetAlign(tview.AlignRight).
					SetTextColor(tcell.ColorGrey).
					SetSelectedStyle(tcell.Style{}.
						Attributes(tcell.AttrBold).
						Background(tcell.ColorWhite).
						Foreground(tcell.ColorDefault),
					),
				)
			}

			if prevrow > -1 {
				row = prevrow
				prevrow = -1
			}

			markFileSelection(row, entry, checkFileSelected(filepath.Join(currentPath, name)))
		}

		filePickTitle.SetText("[::b]Directory: " + currentPath)

		filePickTable.ScrollToBeginning()
		filePickTable.SetSelectable(true, false)
		filePickTable.Select(pos, 0)
	})
}

// dirList lists a directory's contents.
func dirList(testPath string) ([]fs.FileInfo, bool) {
	var dlist []fs.FileInfo

	_, err := os.Lstat(testPath)
	if err != nil {
		return nil, false
	}

	dir, err := os.Lstat(trimPath(testPath, true))
	if err != nil {
		return nil, false
	}

	list, err := ioutil.ReadDir(testPath)
	if err != nil {
		return nil, false
	}

	prevFileInfo = dir
	dlist = append(dlist, dir)

	for _, entry := range list {
		if getHidden() && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		dlist = append(dlist, entry)
	}

	return dlist, true
}

// cellHandler handles on-click events for a table cell.
func cellHandler(row, col int) {
	selectFileHandler(false, false, row)
}

// buttonHandler handles button on-click events for the filepicker buttons.
func buttonHandler(button string) {
	switch button {
	case "ok":
		sendFileList()
		fallthrough

	case "cancel":
		close(fileListChan)
		Pages.RemovePage("filepicker")

	case "hidden":
		toggleHidden()
		go changeDir(false, false)

	case "invert":
		selectFileHandler(false, true)

	case "all":
		selectFileHandler(true, false)
	}

	filePickButtons.Highlight("")
}

// selectFileHandler iterates over the filePickTable's rows,
// determines the type of selection to be made (single, inverse or all),
// and marks the selections.
func selectFileHandler(all, inverse bool, row ...int) {
	var pos int

	singleSelection := !all && !inverse
	inverseSelection := !all && inverse

	if row != nil {
		pos = row[0]
	} else {
		pos, _ = filePickTable.GetSelection()
	}
	totalrows := filePickTable.GetRowCount()

	for i := 0; i < totalrows; i++ {
		var checkSelected bool
		var userSelected []struct{}

		if singleSelection {
			i = pos
			userSelected = append(userSelected, struct{}{})
		}

		cell := filePickTable.GetCell(i, 1)
		if cell == nil {
			return
		}

		entry, ok := cell.GetReference().(fs.FileInfo)
		if !ok {
			return
		}

		fullpath := filepath.Join(currentPath, entry.Name())
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
				filePickTable.Select(i+1, 0)
				return
			}

			break
		}
	}

	filePickTable.Select(pos, 0)
}

// addFileSelection adds a file to the fileSelection list.
func addFileSelection(path string, info fs.FileInfo) {
	fileSelectionLock.Lock()
	defer fileSelectionLock.Unlock()

	if !info.Mode().IsRegular() {
		return
	}

	fileSelection[path] = info
}

// removeFileSelection removes a file from the fileSelection list.
func removeFileSelection(path string) {
	fileSelectionLock.Lock()
	defer fileSelectionLock.Unlock()

	delete(fileSelection, path)
}

// checkFileSelected checks if a file is selected.
func checkFileSelected(path string) bool {
	fileSelectionLock.Lock()
	defer fileSelectionLock.Unlock()

	_, selected := fileSelection[path]

	return selected
}

// markFileSelection marks the selection for files only, directories are skipped.
func markFileSelection(row int, info fs.FileInfo, selected bool, userSelected ...struct{}) {
	if !info.Mode().IsRegular() {
		if info.IsDir() && userSelected != nil {
			go changeDir(true, false)
		}

		return
	}

	cell := filePickTable.GetCell(row, 0)

	if selected {
		cell.Text = "+"
	} else {
		cell.Text = " "
	}
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
	hideLock.Lock()
	defer hideLock.Unlock()

	return isHidden
}

// toggleHidden toggles the hidden files mode.
func toggleHidden() {
	hideLock.Lock()
	defer hideLock.Unlock()

	isHidden = !isHidden
}
