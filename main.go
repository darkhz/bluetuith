package main

import (
	"github.com/darkhz/bluetuith/agent"
	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/ui"
)

func main() {
	bluezConn, err := bluez.NewBluez()
	if err != nil {
		return
	}

	cmd.ParseCmdFlags(bluezConn)

	if err := agent.SetupAgent(bluezConn.Conn()); err != nil {
		return
	}

	ui.SetBluezConn(bluezConn)

	ui.StartUI()

	agent.RemoveAgent()
}
