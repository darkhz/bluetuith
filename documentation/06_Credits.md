# Text User Interface

The TUI was designed using [tview](https://www.github.com/rivo/tview),  which in turn uses [tcell](https://www.github.com/gdamore/tcell).

Thank you, [rivo](https://www.github.com/rivo) and [gdamore](https://www.github.com/gdamore), for these wonderful libraries.

## Progress bars

[progressbar](https://www.github.com/schollz/progressbar) is used to render progress bars

# Configuration and Command-line handling

[koanf](https://www.github.com/knadh/koanf) is used to parse the HJSON configuration and the command-line parameters.

## Text Printing

[color](https://github.com/fatih/color) is used to print colored text to the terminal.

# Bluez and DBus handling

- **vishen**, for the bluez implementation [here](https://github.com/vishen/sluez/blob/master/bluez/device.go).
- **muka**, for the agent implementation [here](https://github.com/muka/go-bluetooth/blob/master/bluez/profile/agent/agent_simple.go).
