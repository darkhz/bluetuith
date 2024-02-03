package cmd

import "github.com/darkhz/bluetuith/bluez"

// Version stores the version information.
var Version string

// Init initializes the application.
func Init(bluez *bluez.Bluez) {
	cmdOptionListAdapters(bluez)
	cmdOptionAdapter(bluez)
	cmdOptionConnectBDAddr(bluez)
	cmdOptionAdapterStates()

	validateKeybindings()
	cmdOptionGenerate()
	cmdOptionTheme()

	cmdOptionGsm()

	cmdOptionReceiveDir()
}

// Parse parses the command-line parameters.
func Parse() {
	config.setup()
	parse()

	cmdOptionVersion()
}
