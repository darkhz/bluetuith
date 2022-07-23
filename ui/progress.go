package ui

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/godbus/dbus/v5"
	"github.com/schollz/progressbar/v3"
)

// Progress describes a progress indicator, which will display
// a description and a progress bar.
type Progress struct {
	desc        *tview.TableCell
	progress    *tview.TableCell
	progressBar *progressbar.ProgressBar

	recv   bool
	status string

	signal chan *dbus.Signal
}

var (
	ProgressView       *tview.Table
	StatusProgressView *tview.Table
	progressFlex       *tview.Flex

	progressCount     int
	progressCountLock sync.Mutex
)

const progressViewButtonRegion = `["resume"][::b][Resume[][""] ["suspend"][::b][Pause[][""] ["cancel"][::b][Cancel[][""]`

// NewProgress returns a new Progress.
func NewProgress(transferPath dbus.ObjectPath, props bluez.ObexTransferProperties, recv bool) *Progress {
	var progress Progress
	var progressText string

	if recv {
		progressText = "Receiving"
	} else {
		progressText = "Sending"
	}

	incProgressCount()

	count := getProgressCount()
	title := fmt.Sprintf(" [::b]%s %s[-:-:-]", progressText, props.Name)

	progress.desc = tview.NewTableCell(title).
		SetExpansion(1).
		SetSelectable(false).
		SetAlign(tview.AlignLeft)

	progress.progress = tview.NewTableCell("").
		SetExpansion(1).
		SetSelectable(false).
		SetReference(&progress).
		SetAlign(tview.AlignRight)

	progress.progressBar = progressbar.NewOptions64(
		int64(props.Size),
		progressbar.OptionSpinnerType(34),
		progressbar.OptionSetWriter(&progress),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionThrottle(200*time.Millisecond),
	)

	progress.recv = recv
	progress.signal = ObexConn.WatchSignal()

	App.QueueUpdateDraw(func() {
		progressView(false)
		statusProgressView(true)

		rows := ProgressView.GetRowCount()

		StatusProgressView.SetCell(0, 0, progress.desc)
		StatusProgressView.SetCell(0, 1, progress.progress)

		ProgressView.SetCell(rows+1, 0, tview.NewTableCell("#"+strconv.Itoa(count)).
			SetReference(transferPath).
			SetAlign(tview.AlignCenter),
		)
		ProgressView.SetCell(rows+1, 1, progress.desc)
		ProgressView.SetCell(rows+1, 2, progress.progress)
	})

	return &progress
}

// StartProgress creates a new progress indicator, monitors the OBEX DBus interface for transfer events,
// and displays the progress on the screen. If the optional path parameter is provided, it means that
// a file is being received, and on transfer completion, the received file should be moved to a user-accessible
// directory.
func StartProgress(transferPath dbus.ObjectPath, props bluez.ObexTransferProperties, path ...string) bool {
	progress := NewProgress(transferPath, props, path != nil)

	for {
		select {
		case signal, ok := <-progress.signal:
			if !ok {
				progress.status = "error"
				progress.FinishProgress(transferPath, path...)
				return false
			}

			props, ok := ObexConn.ParseSignalData(signal).(bluez.ObexProperties)
			if !ok {
				continue
			}

			if transferPath != signal.Path {
				continue
			}

			switch props.TransferProperties.Status {
			case "error":
				ErrorMessage(errors.New("Transfer has failed for " + props.TransferProperties.Name))
				fallthrough

			case "complete":
				progress.status = props.TransferProperties.Status
				progress.FinishProgress(transferPath, path...)
				return true
			}

			progress.progressBar.Set64(int64(props.TransferProperties.Transferred))
		}
	}
}

// SuspendProgress suspends the transfer.
// This does not work when a file is being received.
func SuspendProgress() {
	transferPath, progress := getProgressData()
	if transferPath == "" {
		return
	}

	if progress.recv {
		InfoMessage("Cannot resume/suspend receiving transfer", false)
		return
	}

	ObexConn.SuspendTransfer(transferPath)
}

// ResumeProgress resumes the transfer.
// This does not work when a file is being received.
func ResumeProgress() {
	transferPath, progress := getProgressData()
	if transferPath == "" {
		return
	}

	if progress.recv {
		InfoMessage("Cannot resume/suspend receiving transfer", false)
		return
	}

	ObexConn.ResumeTransfer(transferPath)
}

// CancelProgress cancels the transfer.
// This does not work when a file is being received.
func CancelProgress() {
	transferPath, progress := getProgressData()
	if transferPath == "" {
		return
	}

	if progress.recv {
		InfoMessage("Cannot cancel receiving transfer", false)
		return
	}

	ObexConn.CancelTransfer(transferPath)
	ObexConn.Conn().RemoveSignal(progress.signal)

	close(progress.signal)
}

// FinishProgress removes the progress indicator from view. If a file was received, as indicated by the path parameter,
// the file is moved from the "root" (usually the ~/.cache/obexd folder) to the user's home directory.
func (p *Progress) FinishProgress(transferPath dbus.ObjectPath, path ...string) {
	decProgressCount()
	ObexConn.Conn().RemoveSignal(p.signal)

	App.QueueUpdateDraw(func() {
		for row := 0; row < ProgressView.GetRowCount(); row++ {
			cell := ProgressView.GetCell(row, 0)
			if cell == nil {
				continue
			}

			path, ok := cell.GetReference().(dbus.ObjectPath)
			if !ok {
				continue
			}

			if path == transferPath {
				ProgressView.RemoveRow(row)
				ProgressView.RemoveRow(row - 1)

				break
			}
		}

		if getProgressCount() == 0 {
			if pg, _ := Status.GetFrontPage(); pg == "progressview" {
				Status.RemovePage("progressview")
			}

			if pg, _ := Pages.GetFrontPage(); pg == "progressview" {
				Pages.RemovePage("progressview")
			}

			Status.SwitchToPage("messages")
		}
	})

	if path != nil && p.status == "complete" {
		if err := savefile(path[0]); err != nil {
			ErrorMessage(err)
		}
	}
}

// Write is used by the progressbar to display the progress on the screen.
func (p *Progress) Write(b []byte) (int, error) {
	App.QueueUpdateDraw(func() {
		p.progress.SetText(string(b))
	})

	return 0, nil
}

//gocyclo: ignore
// progressView initializes and, if switchToView is set, displays the progress view.
func progressView(switchToView bool) {
	if progressFlex == nil {
		title := tview.NewTextView()
		title.SetDynamicColors(true)
		title.SetTextAlign(tview.AlignLeft)
		title.SetText("[::bu]Progress View")
		title.SetBackgroundColor(tcell.ColorDefault)

		ProgressView = tview.NewTable()
		ProgressView.SetSelectable(true, false)
		ProgressView.SetBackgroundColor(tcell.ColorDefault)
		ProgressView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEscape:
				if Status.HasPage("progressview") && getProgressCount() > 0 {
					Status.SwitchToPage("progressview")
				}

				Pages.SwitchToPage("main")
			}

			switch event.Rune() {
			case 'x':
				CancelProgress()

			case 'z':
				SuspendProgress()

			case 'g':
				ResumeProgress()

			case 'q':
				go quit()
			}

			return event
		})

		progressViewButtons := tview.NewTextView()
		progressViewButtons.SetRegions(true)
		progressViewButtons.SetDynamicColors(true)
		progressViewButtons.SetTextAlign(tview.AlignLeft)
		progressViewButtons.SetText(progressViewButtonRegion)
		progressViewButtons.SetBackgroundColor(tcell.ColorDefault)
		progressViewButtons.SetHighlightedFunc(func(added, removed, remaining []string) {
			if added == nil {
				return
			}

			for _, region := range progressViewButtons.GetRegionInfos() {
				if region.ID == added[0] {
					switch added[0] {
					case "resume":
						ResumeProgress()

					case "suspend":
						SuspendProgress()

					case "cancel":
						CancelProgress()
					}

					progressViewButtons.Highlight("")

					break
				}
			}
		})

		progressFlex = tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(title, 1, 0, false).
			AddItem(ProgressView, 0, 10, true).
			AddItem(progressViewButtons, 2, 0, false)
	}

	if switchToView {
		if pg, _ := Status.GetFrontPage(); pg == "progressview" {
			Status.SwitchToPage("messages")
		}

		if getProgressCount() == 0 {
			InfoMessage("No transfers are in progress", false)
			return
		}

		Pages.AddAndSwitchToPage("progressview", progressFlex, true)
	}
}

// statusProgressView initializes and, if switchToView is set, displays the progress view in the status bar.
func statusProgressView(switchToView bool) {
	if StatusProgressView == nil {
		StatusProgressView = tview.NewTable()
		StatusProgressView.SetSelectable(true, true)
		StatusProgressView.SetBackgroundColor(tcell.ColorDefault)
	}

	Status.AddPage("progressview", StatusProgressView, true, false)

	if pg, _ := Pages.GetFrontPage(); pg != "progressview" && switchToView {
		Status.SwitchToPage("progressview")
	}
}

// getProgressData gets the transfer DBus object path and the progress data
// from the current selection in the ProgressView.
func getProgressData() (dbus.ObjectPath, *Progress) {
	row, _ := ProgressView.GetSelection()

	pathCell := ProgressView.GetCell(row, 0)
	if pathCell == nil {
		return "", nil
	}

	progCell := ProgressView.GetCell(row, 2)
	if progCell == nil {
		return "", nil
	}

	transferPath, ok := pathCell.GetReference().(dbus.ObjectPath)
	if !ok {
		return "", nil
	}

	progress, ok := progCell.GetReference().(*Progress)
	if !ok {
		return "", nil
	}

	return transferPath, progress
}

// getProgressCount returns the progress count.
func getProgressCount() int {
	progressCountLock.Lock()
	defer progressCountLock.Unlock()

	return progressCount
}

// incProgressCount increments the progress count.
func incProgressCount() {
	progressCountLock.Lock()
	defer progressCountLock.Unlock()

	progressCount++
}

// decProgressCount decrements the progress count.
func decProgressCount() {
	progressCountLock.Lock()
	defer progressCountLock.Unlock()

	progressCount--
}
