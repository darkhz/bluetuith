package ui

import (
	"context"
	"errors"
	"time"

	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/theme"
	"github.com/darkhz/tview"
	"github.com/gdamore/tcell/v2"
)

type Status struct {
	// MessageBox is an area to display messages.
	MessageBox *tview.TextView

	// Help is an area to display help keybindings.
	Help *tview.TextView

	// InputField is an area to interact with messages.
	InputField *tview.InputField

	sctx    context.Context
	scancel context.CancelFunc
	msgchan chan message

	itemCount int

	*tview.Pages
}

type message struct {
	text    string
	persist bool
}

// statusBar sets up the statusbar.
func statusBar() *tview.Flex {
	UI.Status.Pages = tview.NewPages()
	UI.Status.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	UI.Status.InputField = tview.NewInputField()
	UI.Status.InputField.SetLabelColor(theme.GetColor(theme.ThemeText))
	UI.Status.InputField.SetFieldTextColor(theme.GetColor(theme.ThemeText))
	UI.Status.InputField.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))
	UI.Status.InputField.SetFieldBackgroundColor(theme.GetColor(theme.ThemeBackground))

	UI.Status.MessageBox = tview.NewTextView()
	UI.Status.MessageBox.SetDynamicColors(true)
	UI.Status.MessageBox.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	UI.Status.Help = tview.NewTextView()
	UI.Status.Help.SetDynamicColors(true)
	UI.Status.Help.SetBackgroundColor(theme.GetColor(theme.ThemeBackground))

	UI.Status.AddPage("input", UI.Status.InputField, true, true)
	UI.Status.AddPage("messages", UI.Status.MessageBox, true, true)
	UI.Status.SwitchToPage("messages")

	UI.Status.msgchan = make(chan message, 10)
	UI.Status.sctx, UI.Status.scancel = context.WithCancel(context.Background())

	go startStatus()

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(UI.Status.Pages, 1, 0, false)
	if !cmd.IsPropertyEnabled("no-help-display") {
		flex.AddItem(horizontalLine(), 1, 0, false)
		flex.AddItem(UI.Status.Help, 1, 0, false)
	}

	UI.Status.itemCount = flex.GetItemCount()

	return flex
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
					switch cmd.KeyOperation(event) {
					case cmd.KeySelect:
						ch <- true

						exit()

					case cmd.KeyClose:
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
	case UI.Status.msgchan <- message{theme.ColorWrap(theme.ThemeStatusInfo, text), persist}:
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
	case UI.Status.msgchan <- message{theme.ColorWrap(theme.ThemeStatusError, "Error: "+err.Error()), false}:
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
