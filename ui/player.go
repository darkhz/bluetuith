package ui

import (
	"sync"
	"time"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
	"golang.org/x/sync/semaphore"
)

var (
	playerStop        chan struct{}
	playerKeyEvent    chan string
	playerButtonEvent chan struct{}

	mediaLock  sync.Mutex
	playerLock *semaphore.Weighted

	playerSkip bool
)

const mediaButtons = `["rewind"][::b][<<][""] ["prev"][::b][<][""] ["play"][::b][|>][""] ["next"][::b][>][""] ["fastforward"][::b][>>][""]`

// StartMediaPlayer shows the media player.
func StartMediaPlayer() {
	mediaLock.Lock()
	defer mediaLock.Unlock()

	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return
	}

	err := BluezConn.InitMediaPlayer(device.Path)
	if err != nil {
		ErrorMessage(err)
		return
	}

	if playerKeyEvent == nil {
		playerStop = make(chan struct{})
		playerKeyEvent = make(chan string, 1)
		playerButtonEvent = make(chan struct{}, 1)

		playerLock = semaphore.NewWeighted(1)
	}

	go mediaPlayerLoop(device.Name)
}

// StopMediaPlayer closes the media player.
func StopMediaPlayer() {
	select {
	case <-playerStop:

	case playerStop <- struct{}{}:

	default:
	}
}

// mediaPlayerLoop updates the media player.
func mediaPlayerLoop(deviceName string) {
	if !playerLock.TryAcquire(1) {
		return
	}
	defer playerLock.Release(1)

	player, views := setupMediaPlayer(deviceName)
	playerInfo := views[0]
	playerTitle := views[1]
	playerProgress := views[2]
	playerTrack := views[3]
	playerButtons := views[4]

	App.QueueUpdateDraw(func() {
		UILayout.AddItem(player, 8, 0, false)
	})

	mediaSignal := BluezConn.WatchSignal()

	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

PlayerLoop:
	for {
		media, err := BluezConn.GetMediaProperties()
		if err != nil {
			break PlayerLoop
		}

		_, _, width, _ := Pages.GetRect()
		title, buttons, tracknum, progress := getProgress(media, mediaButtons, width, isPlayerSkip())

		App.QueueUpdateDraw(func() {
			playerInfo.SetText(media.Track.Artist + " - " + media.Track.Album)

			playerTitle.SetText(title)
			playerTrack.SetText(tracknum)
			playerButtons.SetText(buttons)
			playerProgress.SetText(progress)
		})

		select {
		case <-playerStop:
			break PlayerLoop

		case highlight, ok := <-playerKeyEvent:
			if !ok {
				break PlayerLoop
			}

			playerButtons.Highlight(highlight)
			t.Reset(1 * time.Second)

		case <-playerButtonEvent:
			t.Reset(1 * time.Second)

		case signal, ok := <-mediaSignal:
			if !ok {
				break PlayerLoop
			}

			_, ok = BluezConn.ParseSignalData(signal).(bluez.MediaProperties)
			if !ok {
				continue PlayerLoop
			}

			t.Reset(1 * time.Second)

		case <-t.C:
		}
	}

	App.QueueUpdateDraw(func() {
		UILayout.RemoveItem(player)
	})

	BluezConn.Conn().RemoveSignal(mediaSignal)
}

// setupMediaPlayer sets up the media player elements.
func setupMediaPlayer(deviceName string) (*tview.Flex, []*tview.TextView) {
	info := tview.NewTextView()
	info.SetDynamicColors(true)
	info.SetTextAlign(tview.AlignCenter)
	info.SetTextColor(theme.GetColor("Text"))
	info.SetBackgroundColor(theme.GetColor("Background"))

	title := tview.NewTextView()
	title.SetDynamicColors(true)
	title.SetTextAlign(tview.AlignCenter)
	title.SetTextColor(theme.GetColor("Text"))
	title.SetBackgroundColor(theme.GetColor("Background"))

	progress := tview.NewTextView()
	progress.SetDynamicColors(true)
	progress.SetTextAlign(tview.AlignCenter)
	progress.SetTextColor(theme.GetColor("Text"))
	progress.SetBackgroundColor(theme.GetColor("Background"))

	track := tview.NewTextView()
	track.SetDynamicColors(true)
	track.SetTextAlign(tview.AlignLeft)
	track.SetTextColor(theme.GetColor("Text"))
	track.SetBackgroundColor(theme.GetColor("Background"))

	device := tview.NewTextView()
	device.SetText(deviceName)
	device.SetDynamicColors(true)
	device.SetTextAlign(tview.AlignRight)
	device.SetTextColor(theme.GetColor("Text"))
	device.SetBackgroundColor(theme.GetColor("Background"))

	buttons := tview.NewTextView()
	buttons.SetRegions(true)
	buttons.SetText(mediaButtons)
	buttons.SetDynamicColors(true)
	buttons.SetTextAlign(tview.AlignCenter)
	buttons.SetTextColor(theme.GetColor("Text"))
	buttons.SetBackgroundColor(theme.GetColor("Background"))
	buttons.SetHighlightedFunc(func(added, removed, remaining []string) {
		var ch rune
		var key tcell.Key
		var highlight string

		if added == nil {
			return
		}

		switch added[0] {
		case "play":
			key, ch = tcell.KeyRune, ' '

		case "next":
			key, ch = tcell.KeyRune, '>'

		case "prev":
			key, ch = tcell.KeyRune, '<'

		case "fastforward":
			highlight = added[0]
			key, ch = tcell.KeyRight, '-'

		case "rewind":
			highlight = added[0]
			key, ch = tcell.KeyLeft, '-'

		default:
			return
		}

		playerEvents(tcell.NewEventKey(key, ch, tcell.ModNone), true)
		buttons.Highlight(highlight)
	})

	buttonFlex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(track, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(buttons, 0, 1, false).
		AddItem(nil, 1, 0, false).
		AddItem(device, 0, 1, false)
	buttonFlex.SetBackgroundColor(theme.GetColor("Background"))

	player := tview.NewFlex().
		AddItem(nil, 1, 0, false).
		AddItem(title, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(info, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(progress, 1, 0, false).
		AddItem(nil, 1, 0, false).
		AddItem(buttonFlex, 1, 0, false).
		SetDirection(tview.FlexRow)
	player.SetBackgroundColor(theme.GetColor("Background"))

	return player, []*tview.TextView{info, title, progress, track, buttons}
}

// playerEvents handles the media player events.
func playerEvents(event *tcell.EventKey, button bool) {
	var highlight string
	var nokey, norune bool

	if UILayout.GetItemCount() <= 2 {
		return
	}

	switch event.Key() {
	case tcell.KeyRight:
		BluezConn.FastForward()

		setPlayerSkip(true)
		highlight = "fastforward"

	case tcell.KeyLeft:
		BluezConn.Rewind()

		setPlayerSkip(true)
		highlight = "rewind"

	default:
		nokey = true
	}

	switch event.Rune() {
	case '<':
		BluezConn.Previous()

	case '>':
		BluezConn.Next()

	case ']':
		BluezConn.Stop()

	case ' ':
		if isPlayerSkip() {
			BluezConn.Play()
			setPlayerSkip(false)

			break
		}

		BluezConn.TogglePlayPause()

	default:
		norune = true
	}

	if !nokey || !norune {
		if button {
			select {
			case playerButtonEvent <- struct{}{}:

			default:
			}

			return
		}

		select {
		case playerKeyEvent <- highlight:

		default:
		}
	}
}

func isPlayerSkip() bool {
	mediaLock.Lock()
	defer mediaLock.Unlock()

	return playerSkip
}

func setPlayerSkip(skip bool) {
	mediaLock.Lock()
	defer mediaLock.Unlock()

	playerSkip = skip
}
