
Generally, bluetuith will work out-of-the-box, with no configuration required.

In case you need to specify settings without using command-line options, a config file can be used.

Typing `bluetuith --help` will show you the location of the config file.<br />

The configuration file is in the [HJSON](https://hjson.github.io/) format.
You can use the [--generate](03_Usage/01_Command_Line_Options.md#generate) flag to generate the configuration.<br /><br />
For example:
```
{
  adapter: hci0
  receive-dir: /home/user/files
  keybindings: {
  	Menu: Alt+m
  }
  theme: {
  	Adapter: red
  }
}

```
# Keybindings
The keybinding configuration is a list of `KeybindingType: keybinding` values.<br />

While defining keybindings, the global keybindings must not conflict with the non-global ones.<br />
It is possible to have duplicate keybindings amongst various non-global keybindings, provided<br />
they are not part of the same context.

For example, this is allowed:
```
keybindings: {
	FilebrowserSelect: Space
	PlayerTogglePlay: Space
}
```
But this isn't:
```
keybindings: {
	Menu: Alt+m
	Close: Alt+m
}
```

## Modifiers
The modifiers currently supported for keybindings are `Ctrl`, `Alt` and `Shift`. `Shift` should only be used in rare cases.
For example, instead of :
- `Shift+a`, type `A`
- `Alt+Shift+e`, type `Alt+E`

and so on.

For the '+' and the space characters, type Plus and Space respectively.
For example,
```
keybindings: {
	AdapterChange: Ctrl+Plus
	PlayerStop: Ctrl+Space
}
```

## Types
Note that some keybinding combinations may be valid, but may not work due to the way your terminal/environment handles it.

The keybinding types are as follows:
| Type                          | Global   | Context  | Default Keybinding   | Description                         |
| ----------------------------- | -------- | -------  | -------------------- | ----------------------------------- |
| Menu                          | Yes      | App      | Alt+m                | Menu                                |
| Select                        | Yes      | App      | Enter                | Select an item                      |
| Cancel                        | Yes      | App      | Ctrl+X               | Cancel an operation                 |
| Suspend                       | Yes      | App      | Ctrl+Z               | Suspend the application             |
| Quit                          | Yes      | App      | Q                    | Quit                                |
| Switch                        | Yes      | App      | Tab                  | Switch between menuitems/buttons    |
| Close                         | Yes      | App      | Esc                  | Close popups/pages                  |
| Help                          | Yes      | App      | ?                    | Show help                           |
| AdapterChange                 | No       | Device   | a                    | Change adapters                     |
| AdapterTogglePower            | No       | Device   | o                    | Toggle adapter power state          |
| AdapterToggleDiscoverable     | No       | Device   | S                    | Toggle adapter discoverable state   |
| AdapterTogglePairable         | No       | Device   | P                    | Toggle adapter pairable state       |
| AdapterToggleScan             | No       | Device   | s                    | Toggle adapter scan state           |
| DeviceSendFiles               | No       | Device   | f                    | Start file transfer session         |
| DeviceNetwork                 | No       | Device   | n                    | Show network Options                |
| DeviceConnect                 | No       | Device   | c                    | Connect to device                   |
| DevicePair                    | No       | Device   | p                    | Pair with device                    |
| DeviceTrust                   | No       | Device   | t                    | Trust device                        |
| DeviceAudioProfiles           | No       | Device   | A                    | Show device's audio profiles        |
| DeviceInfo                    | No       | Device   | i                    | Show device information             |
| DeviceRemove                  | No       | Device   | d                    | Remove device                       |
| PlayerShow                    | No       | Device   | m                    | Show Media Player                   |
| PlayerHide                    | No       | Device   | M                    | Hide Media Player                   |
| PlayerTogglePlay              | No       | Device   | Space                | Play/Pause                          |
| PlayerNext                    | No       | Device   | >                    | Next Track                          |
| PlayerPrevious                | No       | Device   | <                    | Previous Track                      |
| PlayerSeekForward             | No       | Device   | Left                 | Seek Forward                        |
| PlayerSeekBackward            | No       | Device   | Right                | Seek Backward                       |
| PlayerStop                    | No       | Device   | ]                    | Stop Playback                       |
| FilebrowserDirForward         | No       | Files    | Right                | Go forward a directory              |
| FilebrowserDirBack            | No       | Files    | Left                 | Go back a directory                 |
| FilebrowserSelect             | No       | Files    | Space                | Select item                         |
| FilebrowserInvertSelection    | No       | Files    | a                    | Invert selection                    |
| FilebrowserSelectAll          | No       | Files    | A                    | Select all                          |
| FilebrowserRefresh            | No       | Files    | Ctrl+R               | Refresh                             |
| FilebrowserToggleHidden       | No       | Files    | h                    | Toggle hidden files                 |
| FilebrowserConfirmSelection   | No       | Files    | Ctrl+S               | Confirm Selection                   |
| ProgressView                  | No       | Progress | v                    | View Downloads                      |
| ProgressTransferSuspend       | No       | Progress | z                    | Suspend Transfer                    |
| ProgressTransferResume        | No       | Progress | g                    | Resume Transfer                     |
| ProgressTransferCancel        | No       | Progress | x                    | Cancel Transfer                     |

# Themes
The theme configuration is a list of `ElementType: color` values.<br />
Color names or hexadecimal values can be provided for each element type.

For example:
```
theme: {
	Adapter: red
	Device: #000000
}
```

To get a list of available element types and colors, use the `--help` command-line option.
