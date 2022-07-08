[![Go Report Card](https://goreportcard.com/badge/github.com/darkhz/bluetuith)](https://goreportcard.com/report/github.com/darkhz/bluetuith)

![demo](demo/demo.gif)

# bluetuith
bluetuith is a TUI-based bluetooth connection manager, which can interact with bluetooth adapters and devices.
It aims to be a replacement to most bluetooth managers, like blueman.

This is only available on Linux.

This project is currently in the alpha stage.

## Features
- Perform pairing with authentication.
- Perform connection/disconnection to different devices.
- Interact with bluetooth adapters, toggle power and discovery states
- Mouse support

## Requirements
- Bluez
- DBus

## Installation
If the `go` compiler is present in your system, you can install it via the following command:
`go install github.com/darkhz/bluetuith@latest`. Make sure that `$HOME/go/bin` is in your `$PATH`.

Or you can navigate to the releases section and download a binary that matches your architecture.

## Keybindings
|Operation                       |Keybinding                   |
|--------------------------------|-----------------------------|
|Open the menu                   |<kbd>Ctrl</kbd>+<kbd>m</kbd> |
|Navigate between menus          |<kbd>Tab</kbd>               |
|Navigate between devices/options|<kbd>Up</kbd>/<kbd>Down</kbd>|
|Toggle adapter power state      |<kbd>o</kbd>                 |
|Toggle scan (discovery state)   |<kbd>s</kbd>                 |
|Change adapter                  |<kbd>a</kbd>                 |
|Connect to selected device      |<kbd>c</kbd>                 |
|Pair with selected device       |<kbd>p</kbd>                 |
|Trust selected device           |<kbd>t</kbd>                 |
|Remove device from adapter      |<kbd>r</kbd>                 |
|Quit                            |<kbd>q</kbd>                 |

## Planned features

 - [ ] OBEX file transfer.
 - [ ] Display the device type and icon.
 - [ ] Display range (RSSI) of the connected device.

## Additional notes
- Ensure that the bluetooth service is up and running, and it is visible to DBus before launching the application. With systemd you can find out the status using the following command: `systemctl status bluetooth.service`.

## Credits
Special thanks to:
- **vishen**, for the bluez implementation [here](https://github.com/vishen/sluez/blob/master/bluez/device.go).
- **muka**, for the agent implementation [here](https://github.com/muka/go-bluetooth/blob/master/bluez/profile/agent/agent_simple.go).
