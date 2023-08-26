package cmd

import "github.com/darkhz/bluetuith/bluez"

// Init parses the command-line parameters and initializes the application.
func Init(bluez *bluez.Bluez) {
	config.setup()

	parse()
	validateKeybindings()

	cmdOptionAdapter(bluez)
	cmdOptionListAdapters(bluez)

	cmdOptionGenerate()
	cmdOptionTheme()

	cmdOptionGsm()

	cmdOptionReceiveDir()
}
