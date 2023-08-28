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

type MediaPlayer struct {
	skip bool

	keyEvent               chan string
	stopEvent, buttonEvent chan struct{}

	playerLock *semaphore.Weighted
	lock       sync.Mutex
}

const mediaButtons = `["rewind"][::b][<<][""] ["prev"][::b][<][""] ["play"][::b][|>][""] ["next"][::b][>][""] ["fastforward"][::b][>>][""]`

var mediaplayer MediaPlayer

// StartMediaPlayer shows the media player.
func StartMediaPlayer() {
	mediaplayer.lock.Lock()
	defer mediaplayer.lock.Unlock()

	device := getDeviceFromSelection(false)
	if device.Path == "" {
		return
	}

	err := UI.Bluez.InitMediaPlayer(device.Path)
	if err != nil {
		ErrorMessage(err)
		return
	}

	if mediaplayer.keyEvent == nil {
		mediaplayer.stopEvent = make(chan struct{})
		mediaplayer.keyEvent = make(chan string, 1)
		mediaplayer.buttonEvent = make(chan struct{}, 1)

		mediaplayer.playerLock = semaphore.NewWeighted(1)
	}

	go mediaPlayerLoop(device.Name)
}

// StopMediaPlayer closes the media player.
func StopMediaPlayer() {
	select {
	case <-mediaplayer.stopEvent:

	case mediaplayer.stopEvent <- struct{}{}:

	default:
	}
}

// mediaPlayerLoop updates the media player.
func mediaPlayerLoop(deviceName string) {
	if !mediaplayer.playerLock.TryAcquire(1) {
		return
	}
	defer mediaplayer.playerLock.Release(1)

	player, views := setupMediaPlayer(deviceName)
	playerInfo := views[0]
	playerTitle := views[1]
	playerProgress := views[2]
	playerTrack := views[3]
	playerButtons := views[4]

	UI.QueueUpdateDraw(func() {
		UI.Layout.AddItem(player, 8, 0, false)
	})

	mediaSignal := UI.Bluez.WatchSignal()

	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

PlayerLoop:
	for {
		media, err := UI.Bluez.GetMediaProperties()
		if err != nil {
			break PlayerLoop
		}

		_, _, width, _ := UI.Pages.GetRect()
		title, buttons, tracknum, progress := getProgress(media, mediaButtons, width, isPlayerSkip())

		UI.QueueUpdateDraw(func() {
			playerInfo.SetText(media.Track.Artist + " - " + media.Track.Album)

			playerTitle.SetText(title)
			playerTrack.SetText(tracknum)
			playerButtons.SetText(buttons)
			playerProgress.SetText(progress)
		})

		select {
		case <-mediaplayer.stopEvent:
			break PlayerLoop

		case highlight, ok := <-mediaplayer.keyEvent:
			if !ok {
				break PlayerLoop
			}

			playerButtons.Highlight(highlight)
			t.Reset(1 * time.Second)

		case <-mediaplayer.buttonEvent:
			t.Reset(1 * time.Second)

		case signal, ok := <-mediaSignal:
			if !ok {
				break PlayerLoop
			}

			_, ok = UI.Bluez.ParseSignalData(signal).(bluez.MediaProperties)
			if !ok {
				continue PlayerLoop
			}

			t.Reset(1 * time.Second)

		case <-t.C:
		}
	}

	UI.QueueUpdateDraw(func() {
		UI.Layout.RemoveItem(player)
	})

	UI.Bluez.Conn().RemoveSignal(mediaSignal)
}

// setupMediaPlayer sets up the media player elements.
func setupMediaPlayer(deviceName string) (*tview.Flex, []*tview.TextView) {
	info := tview.NewTextView()
	info.SetDynamicColors(true)
	info.SetTextAlign(tview.AlignCenter)
	info.SetTextColor(theme.GetColor(theme.ThemeText))
	info.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	title := tview.NewTextView()
	title.SetDynamicColors(true)
	title.SetTextAlign(tview.AlignCenter)
	title.SetTextColor(theme.GetColor(theme.ThemeText))
	title.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	progress := tview.NewTextView()
	progress.SetDynamicColors(true)
	progress.SetTextAlign(tview.AlignCenter)
	progress.SetTextColor(theme.GetColor(theme.ThemeText))
	progress.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	track := tview.NewTextView()
	track.SetDynamicColors(true)
	track.SetTextAlign(tview.AlignLeft)
	track.SetTextColor(theme.GetColor(theme.ThemeText))
	track.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	device := tview.NewTextView()
	device.SetText(deviceName)
	device.SetDynamicColors(true)
	device.SetTextAlign(tview.AlignRight)
	device.SetTextColor(theme.GetColor(theme.ThemeText))
	device.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	buttons := tview.NewTextView()
	buttons.SetRegions(true)
	buttons.SetText(mediaButtons)
	buttons.SetDynamicColors(true)
	buttons.SetTextAlign(tview.AlignCenter)
	buttons.SetTextColor(theme.GetColor(theme.ThemeText))
	buttons.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
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
	buttonFlex.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

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
	player.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	return player, []*tview.TextView{info, title, progress, track, buttons}
}

// playerEvents handles the media player events.
func playerEvents(event *tcell.EventKey, button bool) {
	var highlight string
	var nokey, norune bool

	if UI.Layout.GetItemCount() <= 2 {
		return
	}

	switch event.Key() {
	case tcell.KeyRight:
		UI.Bluez.FastForward()

		setPlayerSkip(true)
		highlight = "fastforward"

	case tcell.KeyLeft:
		UI.Bluez.Rewind()

		setPlayerSkip(true)
		highlight = "rewind"

	default:
		nokey = true
	}

	switch event.Rune() {
	case '<':
		UI.Bluez.Previous()

	case '>':
		UI.Bluez.Next()

	case ']':
		UI.Bluez.Stop()

	case ' ':
		if isPlayerSkip() {
			UI.Bluez.Play()
			setPlayerSkip(false)

			break
		}

		UI.Bluez.TogglePlayPause()

	default:
		norune = true
	}

	if !nokey || !norune {
		if button {
			select {
			case mediaplayer.buttonEvent <- struct{}{}:

			default:
			}

			return
		}

		select {
		case mediaplayer.keyEvent <- highlight:

		default:
		}
	}
}

func isPlayerSkip() bool {
	mediaplayer.lock.Lock()
	defer mediaplayer.lock.Unlock()

	return mediaplayer.skip
}

func setPlayerSkip(skip bool) {
	mediaplayer.lock.Lock()
	defer mediaplayer.lock.Unlock()

	mediaplayer.skip = skip
}
