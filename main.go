package main

import (
	"github.com/darkhz/bluetuith/agent"
	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/network"
	"github.com/darkhz/bluetuith/ui"
)

func main() {
	var warn string

	cmd.Parse()

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
		warn += "Network connection is disabled since the NetworkManager DBus connection could not be initialized.\n\n"
	}
	cmd.AddProperty("network", err == nil)

	obexConn, err := bluez.NewObex()
	if err != nil {
		warn += "Could not initialize bluez OBEX DBus connection.\n\n"
	} else {
		if err = agent.SetupObexAgent(); err != nil {
			warn += "Send/receive files is disabled since the bluez OBEX agent could not be setup.\n\n"
		}
	}
	cmd.AddProperty("obex", err == nil)

	ui.SetConnections(bluezConn, obexConn, networkConn, warn)
	ui.StartUI()
	ui.StopMediaPlayer()

	agent.RemoveObexAgent()
	agent.RemoveAgent()
}
