package main

import (
	"github.com/darkhz/bluetuith/agent"
	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/network"
	"github.com/darkhz/bluetuith/ui"
)

func main() {
	bluezConn, err := bluez.NewBluez()
	if err != nil {
		cmd.PrintError("Could not initialize bluez DBus connection", err)
	}

	if err := agent.SetupAgent(bluezConn.Conn()); err != nil {
		cmd.PrintError("Could not setup bluez agent", err)
	}

	cmd.Init(bluezConn)

	networkConn, err := network.NewNetwork()
	if err != nil {
		cmd.Print("Network connection is disabled since the NetworkManager DBus connection could not be initialized.", 0)
	}
	cmd.AddProperty("network", err == nil)

	obexConn, err := bluez.NewObex()
	if err != nil {
		cmd.Print("Could not initialize bluez OBEX DBus connection.", 0)
	} else {
		if err = agent.SetupObexAgent(); err != nil {
			cmd.Print("Send/receive files is disabled since the bluez OBEX agent could not be setup.", 0)
		}
	}
	cmd.AddProperty("obex", err == nil)

	ui.SetConnections(bluezConn, obexConn, networkConn)
	ui.StartUI()
	ui.StopMediaPlayer()

	agent.RemoveObexAgent()
	agent.RemoveAgent()
}
