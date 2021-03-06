package main

import (
	"fmt"
	"os"

	"github.com/darkhz/bluetuith/agent"
	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/ui"
)

func errMessage(err string) {
	fmt.Fprintf(os.Stderr, "\rError: %s\n", err)
}

func main() {
	bluezConn, err := bluez.NewBluez()
	if err != nil {
		errMessage("Could not initialize bluez DBus connection")
		return
	}

	obexConn, err := bluez.NewObex()
	if err != nil {
		errMessage("Could not initialize bluez OBEX DBus connection")
		return
	}

	cmd.ParseCmdFlags(bluezConn)

	if err := agent.SetupAgent(bluezConn.Conn()); err != nil {
		errMessage("Could not setup bluez agent")
		return
	}

	if err := agent.SetupObexAgent(); err != nil {
		errMessage("Could not setup bluez OBEX agent")
		return
	}

	ui.SetBluezConn(bluezConn)
	ui.SetObexConn(obexConn)

	ui.StartUI()

	agent.RemoveObexAgent()
	agent.RemoveAgent()
}
