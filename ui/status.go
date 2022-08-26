package ui

import (
	"context"
	"errors"
	"time"

	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

type message struct {
	text    string
	persist bool
}

var (
	// Status enables switching between
	// MessageBox and InputBox.
	Status *tview.Pages

	// MessageBox is an area to display messages.
	MessageBox *tview.TextView

	// InputField is an area to interact with messages.
	InputField *tview.InputField

	sctx    context.Context
	scancel context.CancelFunc
	msgchan chan message
)

// statusBar sets up the statusbar.
func statusBar() *tview.Pages {
	Status = tview.NewPages()
	Status.SetBackgroundColor(theme.GetColor("Background"))

	InputField = tview.NewInputField()
	InputField.SetLabelColor(theme.GetColor("Text"))
	InputField.SetFieldTextColor(theme.GetColor("Text"))
	InputField.SetBackgroundColor(theme.GetColor("Background"))
	InputField.SetFieldBackgroundColor(theme.GetColor("Background"))

	MessageBox = tview.NewTextView()
	MessageBox.SetDynamicColors(true)
	MessageBox.SetBackgroundColor(theme.GetColor("Background"))

	Status.AddPage("input", InputField, true, true)
	Status.AddPage("messages", MessageBox, true, true)
	Status.SwitchToPage("messages")

	msgchan = make(chan message, 10)
	sctx, scancel = context.WithCancel(context.Background())

	go startStatus()

	return Status
}

// stopStatus stops the message event loop.
func stopStatus() {
	scancel()
}

// SetInput sets the inputfield label and returns the input text.
func SetInput(label string, multichar ...struct{}) string {
	entered := make(chan bool)

	go func(ch chan bool) {
		exit := func() {
			Status.SwitchToPage("messages")

			_, item := Pages.GetFrontPage()
			App.SetFocus(item)
		}

		App.QueueUpdateDraw(func() {
			InputField.SetText("")
			InputField.SetLabel("[::b]" + label + " ")
			InputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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
			if multichar != nil {
				InputField.SetAcceptanceFunc(nil)
			} else {
				InputField.SetAcceptanceFunc(tview.InputFieldMaxLength(1))
			}

			Status.SwitchToPage("input")
			App.SetFocus(InputField)
		})
	}(entered)

	hasEntered := <-entered
	if !hasEntered {
		return ""
	}

	return InputField.GetText()
}

// InfoMessage sends an info message to the status bar.
func InfoMessage(text string, persist bool) {
	if msgchan == nil {
		return
	}

	select {
	case msgchan <- message{theme.ColorWrap("StatusInfo", text), persist}:
		return

	default:
	}
}

// ErrorMessage sends an error message to the status bar.
func ErrorMessage(err error) {
	if msgchan == nil {
		return
	}

	if errors.Is(err, context.Canceled) {
		return
	}

	select {
	case msgchan <- message{theme.ColorWrap("StatusError", "Error: "+err.Error()), false}:
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
		case <-sctx.Done():
			return

		case msg, ok := <-msgchan:
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

			App.QueueUpdateDraw(func() {
				MessageBox.SetText(msg.text)
			})

		case <-t.C:
			if cleared {
				continue
			}

			cleared = true

			App.QueueUpdateDraw(func() {
				MessageBox.SetText(text)
			})
		}
	}
}
