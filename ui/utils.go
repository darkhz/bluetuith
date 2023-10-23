package ui

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

// SetTrusted sets the trusted state of a device.
func SetTrusted(devicePath string, enable bool) error {
	UI.Bluez.SetDeviceProperty(devicePath, "Trusted", true)

	return nil
}

// GetDeviceFromPath gets a device from the device path.
func GetDeviceFromPath(devicePath string) (bluez.Device, error) {
	device := UI.Bluez.GetDevice(devicePath)
	if device.Path == "" {
		return bluez.Device{}, errors.New("Device not found")
	}

	return device, nil
}

// formatSize returns the human readable form of a size value in bytes.
// Adapted from: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func formatSize(size int64) string {
	const unit = 1000
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "kMGTPE"[exp])
}

// savefile moves a file from the obex cache to a specified user-accessible directory.
// If the directory is not specified, it automatically creates a directory in the
// user's home path and moves the file there.
func savefile(path string) error {
	userpath := cmd.GetProperty("receive-dir")
	if userpath == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		userpath = filepath.Join(homedir, "bluetuith")

		if _, err := os.Stat(userpath); err != nil {
			err = os.Mkdir(userpath, 0700)
			if err != nil {
				return err
			}
		}
	}

	return os.Rename(path, filepath.Join(userpath, filepath.Base(path)))
}

// getSelectionXY gets the coordinates of the current table selection.
func getSelectionXY(table *tview.Table) (int, int) {
	row, _ := table.GetSelection()

	cell := table.GetCell(row, 0)
	x, y, _ := cell.GetLastPosition()

	return x, y
}

// ignoreDefaultEvent ignores the default keyevents in the provided event.
func ignoreDefaultEvent(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlF, tcell.KeyCtrlB:
		return nil
	}

	switch event.Rune() {
	case 'g', 'G', 'j', 'k', 'h', 'l':
		return nil
	}

	return event
}

// getProgress returns the title and progress of the currently playing media.
func getProgress(media bluez.MediaProperties, buttons string, width int, skip bool) (string, string, string, string) {
	var length int

	title := media.Track.Title
	position := media.Position
	duration := media.Track.Duration
	number := strconv.FormatUint(uint64(media.Track.TrackNumber), 10)
	total := strconv.FormatUint(uint64(media.Track.TotalTracks), 10)

	button := "|>"
	oldButton := button

	if !skip {
		switch media.Status {
		case "playing":
			button = "||"

		case "paused":
			button = "|>"

		case "stopped":
			button = "[]"
		}
	}

	buttons = strings.Replace(buttons, oldButton, button, 1)

	width /= 2
	if position >= math.MaxUint32 {
		position = 0
	}
	if position >= duration {
		position = duration
	}

	if duration <= 0 {
		length = 0
	} else {
		length = width * int(position) / int(duration)
	}

	endlength := width - length
	if endlength < 0 {
		endlength = width
	}

	track := "Track " + number + "/" + total
	progress := " " + formatDuration(position) +
		" |" + strings.Repeat("█", length) + strings.Repeat(" ", endlength) + "| " +
		formatDuration(duration)

	return title, buttons, track, progress
}

// horizontalLine returns a box with a thick horizontal line.
func horizontalLine() *tview.Box {
	return tview.NewBox().
		SetBackgroundColor(tcell.ColorDefault).
		SetDrawFunc(func(
			screen tcell.Screen,
			x, y, width, height int) (int, int, int, int) {
			centerY := y + height/2
			for cx := x; cx < x+width; cx++ {
				screen.SetContent(
					cx,
					centerY,
					tview.BoxDrawingsLightHorizontal,
					nil,
					tcell.StyleDefault.Foreground(tcell.ColorWhite),
				)
			}

			return x + 1,
				centerY + 1,
				width - 2,
				height - (centerY + 1 - y)
		})
}

// formatDuration converts a duration into a human-readable format.
func formatDuration(duration uint32) string {
	var durationtext string

	input, err := time.ParseDuration(strconv.FormatUint(uint64(duration), 10) + "ms")
	if err != nil {
		return "00:00"
	}

	d := input.Round(time.Second)

	h := d / time.Hour
	d -= h * time.Hour

	m := d / time.Minute
	d -= m * time.Minute

	s := d / time.Second

	if h > 0 {
		if h < 10 {
			durationtext += "0"
		}

		durationtext += strconv.Itoa(int(h))
		durationtext += ":"
	}

	if m > 0 {
		if m < 10 {
			durationtext += "0"
		}

		durationtext += strconv.Itoa(int(m))
	} else {
		durationtext += "00"
	}

	durationtext += ":"

	if s < 10 {
		durationtext += "0"
	}

	durationtext += strconv.Itoa(int(s))

	return durationtext
}
