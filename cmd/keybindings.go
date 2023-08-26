package cmd

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// KeyData stores the metadata for the key.
type KeyData struct {
	Title   string
	Context KeyContext
	Kb      Keybinding
	Global  bool
}

// Keybinding stores the keybinding.
type Keybinding struct {
	Key  tcell.Key
	Rune rune
	Mod  tcell.ModMask
}

// Key describes the application keybinding type.
type Key string

// The different application keybinding types.
const (
	KeyMenu                        Key = "Menu"
	KeySelect                      Key = "Select"
	KeyCancel                      Key = "Cancel"
	KeySuspend                     Key = "Suspend"
	KeyQuit                        Key = "Quit"
	KeySwitch                      Key = "Switch"
	KeyClose                       Key = "Close"
	KeyHelp                        Key = "Help"
	KeyAdapterChange               Key = "AdapterChange"
	KeyAdapterTogglePower          Key = "AdapterTogglePower"
	KeyAdapterToggleDiscoverable   Key = "AdapterToggleDiscoverable"
	KeyAdapterTogglePairable       Key = "AdapterTogglePairable"
	KeyAdapterToggleScan           Key = "AdapterToggleScan"
	KeyDeviceSendFiles             Key = "DeviceSendFiles"
	KeyDeviceNetwork               Key = "DeviceNetwork"
	KeyDeviceConnect               Key = "DeviceConnect"
	KeyDevicePair                  Key = "DevicePair"
	KeyDeviceTrust                 Key = "DeviceTrust"
	KeyDeviceAudioProfiles         Key = "DeviceAudioProfiles"
	KeyDeviceInfo                  Key = "DeviceInfo"
	KeyDeviceRemove                Key = "DeviceRemove"
	KeyPlayerShow                  Key = "PlayerShow"
	KeyPlayerHide                  Key = "PlayerHide"
	KeyFilebrowserDirForward       Key = "FilebrowserDirForward"
	KeyFilebrowserDirBack          Key = "FilebrowserDirBack"
	KeyFilebrowserSelect           Key = "FilebrowserSelect"
	KeyFilebrowserInvertSelection  Key = "FilebrowserInvertSelection"
	KeyFilebrowserSelectAll        Key = "FilebrowserSelectAll"
	KeyFilebrowserRefresh          Key = "FilebrowserRefresh"
	KeyFilebrowserToggleHidden     Key = "FilebrowserToggleHidden"
	KeyFilebrowserConfirmSelection Key = "FilebrowserConfirmSelection"
	KeyProgressView                Key = "ProgressView"
	KeyProgressTransferSuspend     Key = "ProgressTransferSuspend"
	KeyProgressTransferResume      Key = "ProgressTransferResume"
	KeyProgressTransferCancel      Key = "ProgressTransferCancel"
	KeyPlayerTogglePlay            Key = "PlayerTogglePlay"
	KeyPlayerNext                  Key = "PlayerNext"
	KeyPlayerPrevious              Key = "PlayerPrevious"
	KeyPlayerSeekForward           Key = "PlayerSeekForward"
	KeyPlayerSeekBackward          Key = "PlayerSeekBackward"
	KeyPlayerStop                  Key = "PlayerStop"
)

// KeyContext describes the context where the keybinding is
// supposed to be applied in.
type KeyContext string

// The different context types for keybindings.
const (
	KeyContextApp      KeyContext = "App"
	KeyContextDevice   KeyContext = "Device"
	KeyContextFiles    KeyContext = "Files"
	KeyContextProgress KeyContext = "Progress"
)

var (
	// OperationKeys matches the operation name (or the menu ID) with the keybinding.
	OperationKeys = map[Key]*KeyData{
		KeySwitch: {
			Title:   "Switch",
			Context: KeyContextApp,
			Kb:      Keybinding{tcell.KeyTab, ' ', tcell.ModNone},
			Global:  true,
		},
		KeyClose: {
			Title:   "Close",
			Context: KeyContextApp,
			Kb:      Keybinding{tcell.KeyEscape, ' ', tcell.ModNone},
			Global:  true,
		},
		KeyQuit: {
			Title:   "Quit",
			Context: KeyContextApp,
			Kb:      Keybinding{tcell.KeyRune, 'Q', tcell.ModNone},
			Global:  true,
		},
		KeyMenu: {
			Title:   "Menu",
			Context: KeyContextApp,
			Kb:      Keybinding{tcell.KeyRune, 'm', tcell.ModAlt},
		},
		KeySelect: {
			Title:   "Select",
			Context: KeyContextApp,
			Kb:      Keybinding{tcell.KeyEnter, ' ', tcell.ModNone},
			Global:  true,
		},
		KeyCancel: {
			Title:   "Cancel",
			Context: KeyContextApp,
			Kb:      Keybinding{tcell.KeyCtrlX, ' ', tcell.ModCtrl},
			Global:  true,
		},
		KeySuspend: {
			Title:   "Suspend",
			Context: KeyContextApp,
			Kb:      Keybinding{tcell.KeyCtrlZ, ' ', tcell.ModCtrl},
			Global:  true,
		},
		KeyHelp: {
			Title:   "Help",
			Context: KeyContextApp,
			Kb:      Keybinding{tcell.KeyRune, '?', tcell.ModNone},
			Global:  true,
		},
		KeyAdapterTogglePower: {
			Title:   "Power",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'o', tcell.ModNone},
		},
		KeyAdapterToggleDiscoverable: {
			Title:   "Discoverable",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'S', tcell.ModNone},
		},
		KeyAdapterTogglePairable: {
			Title:   "Pairable",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'P', tcell.ModNone},
		},
		KeyAdapterToggleScan: {
			Title:   "Scan",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 's', tcell.ModNone},
		},
		KeyAdapterChange: {
			Title:   "Change",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'a', tcell.ModNone},
		},
		KeyDeviceConnect: {
			Title:   "Connect",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'c', tcell.ModNone},
		},
		KeyDevicePair: {
			Title:   "Pair",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'p', tcell.ModNone},
		},
		KeyDeviceTrust: {
			Title:   "Trust",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 't', tcell.ModNone},
		},
		KeyDeviceSendFiles: {
			Title:   "Send",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'f', tcell.ModNone},
		},
		KeyDeviceNetwork: {
			Title:   "Network Options",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'n', tcell.ModNone},
		},
		KeyDeviceAudioProfiles: {
			Title:   "Audio Profiles",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'A', tcell.ModNone},
		},
		KeyDeviceInfo: {
			Title:   "Info",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'i', tcell.ModNone},
		},
		KeyDeviceRemove: {
			Title:   "Remove",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'd', tcell.ModNone},
		},
		KeyPlayerShow: {
			Title:   "Show Media Player",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'm', tcell.ModNone},
		},
		KeyPlayerHide: {
			Title:   "Hide Media Player",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, 'M', tcell.ModNone},
		},
		KeyPlayerTogglePlay: {
			Title:   "Play/Pause",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, ' ', tcell.ModNone},
		},
		KeyPlayerNext: {
			Title:   "Next",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, '>', tcell.ModNone},
		},
		KeyPlayerPrevious: {
			Title:   "Previous",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, '<', tcell.ModNone},
		},
		KeyPlayerSeekForward: {
			Title:   "Seek Forward",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyLeft, ' ', tcell.ModNone},
		},
		KeyPlayerSeekBackward: {
			Title:   "Seek Backward",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRight, ' ', tcell.ModNone},
		},
		KeyPlayerStop: {
			Title:   "Stop",
			Context: KeyContextDevice,
			Kb:      Keybinding{tcell.KeyRune, ']', tcell.ModNone},
		},
		KeyFilebrowserConfirmSelection: {
			Title:   "Confirm Selection",
			Context: KeyContextFiles,
			Kb:      Keybinding{tcell.KeyCtrlS, ' ', tcell.ModCtrl},
		},
		KeyFilebrowserDirForward: {
			Title:   "Go Forward",
			Context: KeyContextFiles,
			Kb:      Keybinding{tcell.KeyRight, ' ', tcell.ModNone},
		},
		KeyFilebrowserDirBack: {
			Title:   "Go Back",
			Context: KeyContextFiles,
			Kb:      Keybinding{tcell.KeyLeft, ' ', tcell.ModNone},
		},
		KeyFilebrowserSelect: {
			Title:   "Select",
			Context: KeyContextFiles,
			Kb:      Keybinding{tcell.KeyRune, ' ', tcell.ModNone},
		},
		KeyFilebrowserInvertSelection: {
			Title:   "Invert Selection",
			Context: KeyContextFiles,
			Kb:      Keybinding{tcell.KeyRune, 'a', tcell.ModNone},
		},
		KeyFilebrowserSelectAll: {
			Title:   "Select All",
			Context: KeyContextFiles,
			Kb:      Keybinding{tcell.KeyRune, 'A', tcell.ModNone},
		},
		KeyFilebrowserRefresh: {
			Title:   "Refresh",
			Context: KeyContextFiles,
			Kb:      Keybinding{tcell.KeyCtrlR, ' ', tcell.ModCtrl},
		},
		KeyFilebrowserToggleHidden: {
			Title:   "Hidden",
			Context: KeyContextFiles,
			Kb:      Keybinding{tcell.KeyRune, 'h', tcell.ModCtrl},
		},
		KeyProgressTransferResume: {
			Title:   "Resume Transfer",
			Context: KeyContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'g', tcell.ModNone},
		},
		KeyProgressTransferCancel: {
			Title:   "Cancel Transfer",
			Context: KeyContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'x', tcell.ModNone},
		},
		KeyProgressView: {
			Title:   "View Downloads",
			Context: KeyContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'v', tcell.ModNone},
		},
		KeyProgressTransferSuspend: {
			Title:   "Suspend Transfer",
			Context: KeyContextProgress,
			Kb:      Keybinding{tcell.KeyRune, 'z', tcell.ModNone},
		},
	}

	// Keys match the keybinding to the key type.
	Keys map[KeyContext]map[Keybinding]Key

	translateKeys = map[string]string{
		"Pgup":      "PgUp",
		"Pgdn":      "PgDn",
		"Pageup":    "PgUp",
		"Pagedown":  "PgDn",
		"Upright":   "UpRight",
		"Downright": "DownRight",
		"Upleft":    "UpLeft",
		"Downleft":  "DownLeft",
		"Prtsc":     "Print",
		"Backspace": "Backspace2",
	}
)

// OperationData returns the key data associated with
// the provided keyID and operation name.
func OperationData(operation Key) *KeyData {
	return OperationKeys[operation]
}

// KeyOperation returns the operation name for the provided keyID
// and the keyboard event.
func KeyOperation(event *tcell.EventKey, keyContexts ...KeyContext) Key {
	if Keys == nil {
		Keys = make(map[KeyContext]map[Keybinding]Key)
		for keyName, key := range OperationKeys {
			if Keys[key.Context] == nil {
				Keys[key.Context] = make(map[Keybinding]Key)
			}

			Keys[key.Context][key.Kb] = keyName
		}
	}

	ch := event.Rune()
	if event.Key() != tcell.KeyRune {
		ch = ' '
	}

	kb := Keybinding{event.Key(), ch, event.Modifiers()}

	for _, contexts := range [][]KeyContext{
		keyContexts,
		{
			KeyContextApp,
			KeyContextDevice,
		},
	} {
		for _, context := range contexts {
			if operation, ok := Keys[context][kb]; ok {
				return operation
			}
		}
	}

	return ""
}

// KeyName formats and returns the key's name.
func KeyName(kb Keybinding) string {
	if kb.Key == tcell.KeyRune {
		keyname := string(kb.Rune)
		if kb.Rune == ' ' {
			keyname = "Space"
		}

		if kb.Mod&tcell.ModAlt != 0 {
			keyname = "Alt+" + keyname
		}

		return keyname
	}

	return tcell.NewEventKey(kb.Key, kb.Rune, kb.Mod).Name()
}

// parseKeybindings parses the keybindings from the configuration.
func validateKeybindings() {
	if !config.Exists("keybindings") {
		return
	}

	kbMap := config.StringMap("keybindings")
	if len(kbMap) == 0 {
		return
	}

	keyNames := make(map[string]tcell.Key)
	for key, names := range tcell.KeyNames {
		keyNames[names] = key
	}

	for keyType, key := range kbMap {
		checkBindings(keyType, key, keyNames)
	}

	keyErrors := make(map[Keybinding]string)

	for keyType, keydata := range OperationKeys {
		for existing, data := range OperationKeys {
			if data.Kb == keydata.Kb && data.Title != keydata.Title {
				if data.Context == keydata.Context || data.Global || keydata.Global {
					goto KeyError
				}

				continue

			KeyError:
				if _, ok := keyErrors[keydata.Kb]; !ok {
					keyErrors[keydata.Kb] = fmt.Sprintf("- %s will override %s (%s)", keyType, existing, KeyName(keydata.Kb))
				}
			}
		}
	}

	if len(keyErrors) > 0 {
		err := "Config: The following keybindings will conflict:\n"
		for _, ke := range keyErrors {
			err += ke + "\n"
		}

		PrintError(strings.TrimRight(err, "\n"))
	}
}

// checkBindings validates the provided keybinding.
//
//gocyclo:ignore
func checkBindings(keyType, key string, keyNames map[string]tcell.Key) {
	var runes []rune
	var keys []tcell.Key

	if _, ok := OperationKeys[Key(keyType)]; !ok {
		PrintError(fmt.Sprintf("Config: Invalid key type %s", keyType))
	}

	keybinding := Keybinding{
		Key:  tcell.KeyRune,
		Rune: ' ',
		Mod:  tcell.ModNone,
	}

	tokens := strings.FieldsFunc(key, func(c rune) bool {
		return unicode.IsSpace(c) || c == '+'
	})

	for _, token := range tokens {
		if len(token) > 1 {
			token = cases.Title(language.Und, cases.NoLower).String(token)
		} else if len(token) == 1 {
			keybinding.Rune = rune(token[0])
			runes = append(runes, keybinding.Rune)

			continue
		}

		if translated, ok := translateKeys[token]; ok {
			token = translated
		}

		switch token {
		case "Ctrl":
			keybinding.Mod |= tcell.ModCtrl

		case "Alt":
			keybinding.Mod |= tcell.ModAlt

		case "Shift":
			keybinding.Mod |= tcell.ModShift

		case "Space", "Plus":
			keybinding.Rune = ' '
			if token == "Plus" {
				keybinding.Rune = '+'
			}

			runes = append(runes, keybinding.Rune)

		default:
			if key, ok := keyNames[token]; ok {
				keybinding.Key = key
				keybinding.Rune = ' '
				keys = append(keys, keybinding.Key)
			}
		}
	}

	if keys != nil && runes != nil || len(runes) > 1 || len(keys) > 1 {
		PrintError(
			fmt.Sprintf("Config: More than one key entered for %s (%s)", keyType, key),
		)
	}

	if keybinding.Mod&tcell.ModShift != 0 {
		keybinding.Rune = unicode.ToUpper(keybinding.Rune)

		if unicode.IsLetter(keybinding.Rune) {
			keybinding.Mod &^= tcell.ModShift
		}
	}

	if keybinding.Mod&tcell.ModCtrl != 0 {
		var modKey string

		switch {
		case len(keys) > 0:
			if key, ok := tcell.KeyNames[keybinding.Key]; ok {
				modKey = key
			}

		case len(runes) > 0:
			if keybinding.Rune == ' ' {
				modKey = "Space"
			} else {
				modKey = string(unicode.ToUpper(keybinding.Rune))
			}
		}

		if modKey != "" {
			modKey = "Ctrl-" + modKey
			if key, ok := keyNames[modKey]; ok {
				keybinding.Key = key
				keybinding.Rune = ' '
				keys = append(keys, keybinding.Key)
			}
		}
	}

	if keys == nil && runes == nil {
		PrintError(
			fmt.Sprintf("Config: No key specified or invalid keybinding for %s (%s)", keyType, key),
		)
	}

	OperationKeys[Key(keyType)].Kb = keybinding
}
