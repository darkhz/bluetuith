package ui

import (
	"context"
	"errors"
	"time"

	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

type Status struct {
	// MessageBox is an area to display messages.
	MessageBox *tview.TextView

	// InputField is an area to interact with messages.
	InputField *tview.InputField

	sctx    context.Context
	scancel context.CancelFunc
	msgchan chan message

	*tview.Pages
}

type message struct {
	text    string
	persist bool
}

// statusBar sets up the statusbar.
func statusBar() *tview.Pages {
	UI.Status.Pages = tview.NewPages()
	UI.Status.SetBackgroundColor(theme.GetColor("Background"))

	UI.Status.InputField = tview.NewInputField()
	UI.Status.InputField.SetLabelColor(theme.GetColor("Text"))
	UI.Status.InputField.SetFieldTextColor(theme.GetColor("Text"))
	UI.Status.InputField.SetBackgroundColor(theme.GetColor("Background"))
	UI.Status.InputField.SetFieldBackgroundColor(theme.GetColor("Background"))

	UI.Status.MessageBox = tview.NewTextView()
	UI.Status.MessageBox.SetDynamicColors(true)
	UI.Status.MessageBox.SetBackgroundColor(theme.GetColor("Background"))

	UI.Status.AddPage("input", UI.Status.InputField, true, true)
	UI.Status.AddPage("messages", UI.Status.MessageBox, true, true)
	UI.Status.SwitchToPage("messages")

	UI.Status.msgchan = make(chan message, 10)
	UI.Status.sctx, UI.Status.scancel = context.WithCancel(context.Background())

	go startStatus()

	return UI.Status.Pages
}

// stopStatus stops the message event loop.
func stopStatus() {
	UI.Status.scancel()
}

// SetInput sets the inputfield label and returns the input text.
func SetInput(label string, multichar ...struct{}) string {
	entered := make(chan bool)

	go func(ch chan bool) {
		exit := func() {
			UI.Status.SwitchToPage("messages")

			_, item := UI.Pages.GetFrontPage()
			UI.SetFocus(item)
		}

		UI.QueueUpdateDraw(func() {
			UI.Status.InputField.SetText("")
			UI.Status.InputField.SetLabel("[::b]" + label + " ")

			if multichar != nil {
				UI.Status.InputField.SetAcceptanceFunc(nil)
				UI.Status.InputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					switch event.Key() {
					case tcell.KeyEnter:
						ch <- true

						exit()

					case tcell.KeyEscape:
						ch <- false

						exit()
					}

					return event
				})
			} else {
				UI.Status.InputField.SetAcceptanceFunc(tview.InputFieldMaxLength(1))
				UI.Status.InputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
					switch event.Rune() {
					case 'y':
						ch <- true

						exit()

					default:
						ch <- false

						exit()
					}

					return event
				})
			}

			UI.Status.SwitchToPage("input")
			UI.SetFocus(UI.Status.InputField)
		})
	}(entered)

	hasEntered := <-entered
	if !hasEntered {
		return ""
	}

	return UI.Status.InputField.GetText()
}

// InfoMessage sends an info message to the status bar.
func InfoMessage(text string, persist bool) {
	if UI.Status.msgchan == nil {
		return
	}

	select {
	case UI.Status.msgchan <- message{theme.ColorWrap("StatusInfo", text), persist}:
		return

	default:
	}
}

// ErrorMessage sends an error message to the status bar.
func ErrorMessage(err error) {
	if UI.Status.msgchan == nil {
		return
	}

	if errors.Is(err, context.Canceled) {
		return
	}

	select {
	case UI.Status.msgchan <- message{theme.ColorWrap("StatusError", "Error: "+err.Error()), false}:
		return

	default:
	}
}

// startStatus starts the message event loop
func startStatus() {
	var text string
	var cleared bool

	t := time.NewTicker(2 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-UI.Status.sctx.Done():
			return

		case msg, ok := <-UI.Status.msgchan:
			if !ok {
				return
			}

			t.Reset(2 * time.Second)

			cleared = false

			if msg.persist {
				text = msg.text
			}

			if !msg.persist && text != "" {
				text = ""
			}

			UI.QueueUpdateDraw(func() {
				UI.Status.MessageBox.SetText(msg.text)
			})

		case <-t.C:
			if cleared {
				continue
			}

			cleared = true

			UI.QueueUpdateDraw(func() {
				UI.Status.MessageBox.SetText(text)
			})
		}
	}
}
