package ui

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/godbus/dbus/v5"
	"github.com/schollz/progressbar/v3"
)

// ProgressUI describes a file transfer progress display.
type ProgressUI struct {
	view, status *tview.Table
	flex         *tview.Flex

	total int

	lock sync.Mutex
}

// ProgressIndicator describes a progress indicator, which will display
// a description and a progress bar.
type ProgressIndicator struct {
	desc        *tview.TableCell
	progress    *tview.TableCell
	progressBar *progressbar.ProgressBar

	recv   bool
	status string

	signal chan *dbus.Signal
}

const progressViewButtonRegion = `["resume"][::b][Resume[][""] ["suspend"][::b][Pause[][""] ["cancel"][::b][Cancel[][""]`

var progressUI ProgressUI

// NewProgress returns a new Progress.
func NewProgress(transferPath dbus.ObjectPath, props bluez.ObexTransferProperties, recv bool) *ProgressIndicator {
	var progress ProgressIndicator
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
		SetAlign(tview.AlignLeft).
		SetTextColor(theme.GetColor(theme.ThemeProgressText))

	progress.progress = tview.NewTableCell("").
		SetExpansion(1).
		SetSelectable(false).
		SetReference(&progress).
		SetAlign(tview.AlignRight).
		SetTextColor(theme.GetColor(theme.ThemeProgressBar))

	progress.progressBar = progressbar.NewOptions64(
		int64(props.Size),
		progressbar.OptionSpinnerType(34),
		progressbar.OptionSetWriter(&progress),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionThrottle(200*time.Millisecond),
	)

	progress.recv = recv
	progress.signal = UI.Obex.WatchSignal()

	UI.QueueUpdateDraw(func() {
		progressView(false)
		statusProgressView(true)

		rows := progressUI.view.GetRowCount()

		progressUI.status.SetCell(0, 0, progress.desc)
		progressUI.status.SetCell(0, 1, progress.progress)

		progressUI.view.SetCell(rows+1, 0, tview.NewTableCell("#"+strconv.Itoa(count)).
			SetReference(transferPath).
			SetAlign(tview.AlignCenter),
		)
		progressUI.view.SetCell(rows+1, 1, progress.desc)
		progressUI.view.SetCell(rows+1, 2, progress.progress)
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

			props, ok := UI.Obex.ParseSignalData(signal).(bluez.ObexProperties)
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

	UI.Obex.SuspendTransfer(transferPath)
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

	UI.Obex.ResumeTransfer(transferPath)
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

	UI.Obex.CancelTransfer(transferPath)
	UI.Obex.Conn().RemoveSignal(progress.signal)

	close(progress.signal)
}

// FinishProgress removes the progress indicator from view. If a file was received, as indicated by the path parameter,
// the file is moved from the "root" (usually the ~/.cache/obexd folder) to the user's home directory.
func (p *ProgressIndicator) FinishProgress(transferPath dbus.ObjectPath, path ...string) {
	decProgressCount()
	UI.Obex.Conn().RemoveSignal(p.signal)

	UI.QueueUpdateDraw(func() {
		for row := 0; row < progressUI.view.GetRowCount(); row++ {
			cell := progressUI.view.GetCell(row, 0)
			if cell == nil {
				continue
			}

			path, ok := cell.GetReference().(dbus.ObjectPath)
			if !ok {
				continue
			}

			if path == transferPath {
				progressUI.view.RemoveRow(row)
				progressUI.view.RemoveRow(row - 1)

				break
			}
		}

		if getProgressCount() == 0 {
			if pg, _ := UI.Status.GetFrontPage(); pg == "progressview" {
				UI.Status.RemovePage("progressview")
			}

			if pg, _ := UI.Pages.GetFrontPage(); pg == "progressview" {
				UI.Pages.RemovePage("progressview")
			}

			UI.Status.SwitchToPage("messages")
		}
	})

	if path != nil && p.status == "complete" {
		if err := savefile(path[0]); err != nil {
			ErrorMessage(err)
		}
	}
}

// Write is used by the progressbar to display the progress on the screen.
func (p *ProgressIndicator) Write(b []byte) (int, error) {
	UI.QueueUpdateDraw(func() {
		p.progress.SetText(string(b))
	})

	return 0, nil
}

// progressView initializes and, if switchToView is set, displays the progress view.
//
//gocyclo:ignore
func progressView(switchToView bool) {
	if progressUI.flex == nil {
		title := tview.NewTextView()
		title.SetDynamicColors(true)
		title.SetTextAlign(tview.AlignLeft)
		title.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
		title.SetText(theme.ColorWrap(theme.ThemeText, "Progress View", "::bu"))

		progressUI.view = tview.NewTable()
		progressUI.view.SetSelectable(true, false)
		progressUI.view.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
		progressUI.view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch cmd.KeyOperation(event, cmd.KeyContextProgress) {
			case cmd.KeyClose:
				if UI.Status.HasPage("progressview") && getProgressCount() > 0 {
					UI.Status.SwitchToPage("progressview")
				}

				UI.Pages.SwitchToPage("main")

			case cmd.KeyProgressTransferCancel:
				CancelProgress()

			case cmd.KeyProgressTransferSuspend:
				SuspendProgress()

			case cmd.KeyProgressTransferResume:
				ResumeProgress()

			case cmd.KeyQuit:
				go quit()
			}

			return ignoreDefaultEvent(event)
		})

		progressViewButtons := tview.NewTextView()
		progressViewButtons.SetRegions(true)
		progressViewButtons.SetDynamicColors(true)
		progressViewButtons.SetTextAlign(tview.AlignLeft)
		progressViewButtons.SetText(progressViewButtonRegion)
		progressViewButtons.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
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

		progressUI.flex = tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(title, 1, 0, false).
			AddItem(progressUI.view, 0, 10, true).
			AddItem(progressViewButtons, 2, 0, false)
	}

	if switchToView {
		if pg, _ := UI.Status.GetFrontPage(); pg == "progressview" {
			UI.Status.SwitchToPage("messages")
		}

		if getProgressCount() == 0 {
			InfoMessage("No transfers are in progress", false)
			return
		}

		UI.Pages.AddAndSwitchToPage("progressview", progressUI.flex, true)
	}
}

// statusProgressView initializes and, if switchToView is set, displays the progress view in the status bar.
func statusProgressView(switchToView bool) {
	if progressUI.status == nil {
		progressUI.status = tview.NewTable()
		progressUI.status.SetSelectable(true, true)
		progressUI.status.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	}

	UI.Status.AddPage("progressview", progressUI.status, true, false)

	if pg, _ := UI.Pages.GetFrontPage(); pg != "progressview" && switchToView {
		UI.Status.SwitchToPage("progressview")
	}
}

// getProgressData gets the transfer DBus object path and the progress data
// from the current selection in the progressUI.view.
func getProgressData() (dbus.ObjectPath, *ProgressIndicator) {
	row, _ := progressUI.view.GetSelection()

	pathCell := progressUI.view.GetCell(row, 0)
	if pathCell == nil {
		return "", nil
	}

	progCell := progressUI.view.GetCell(row, 2)
	if progCell == nil {
		return "", nil
	}

	transferPath, ok := pathCell.GetReference().(dbus.ObjectPath)
	if !ok {
		return "", nil
	}

	progress, ok := progCell.GetReference().(*ProgressIndicator)
	if !ok {
		return "", nil
	}

	return transferPath, progress
}

// getProgressCount returns the progress count.
func getProgressCount() int {
	progressUI.lock.Lock()
	defer progressUI.lock.Unlock()

	return progressUI.total
}

// incProgressCount increments the progress count.
func incProgressCount() {
	progressUI.lock.Lock()
	defer progressUI.lock.Unlock()

	progressUI.total++
}

// decProgressCount decrements the progress count.
func decProgressCount() {
	progressUI.lock.Lock()
	defer progressUI.lock.Unlock()

	progressUI.total--
}
