package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/darkhz/bluetuith/agent"
	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/cmd"
	"github.com/darkhz/bluetuith/network"
	"github.com/darkhz/bluetuith/ui"
)

func errMessage(err string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
}

func warnMessage(warn string) {
	fmt.Fprintf(os.Stderr, "Warning: %s\n", warn)
}

func main() {
	if err := cmd.SetupConfig(); err != nil {
		fmt.Println(err)
		return
	}

	bluezConn, err := bluez.NewBluez()
	if err != nil {
		errMessage("Could not initialize bluez DBus connection")
		return
	}

	if err := agent.SetupAgent(bluezConn.Conn()); err != nil {
		errMessage("Could not setup bluez agent")
		return
	}

	cmd.ParseCmdFlags(bluezConn)

	networkConn, err := network.NewNetwork()
	if err != nil {
		warnMessage("Could not initialize NetworkManager DBus connection")
		warnMessage("Network connection is disabled")
	}
	cmd.AddConfigProperty("network", strconv.FormatBool(err == nil))

	obexConn, err := bluez.NewObex()
	if err != nil {
		warnMessage("Could not initialize bluez OBEX DBus connection")
	} else {
		err = agent.SetupObexAgent()
		if err != nil {
			warnMessage("Could not setup bluez OBEX agent")
			warnMessage("Send/receive files is disabled")
		}
	}
	cmd.AddConfigProperty("obex", strconv.FormatBool(err == nil))

	ui.SetBluezConn(bluezConn)
	ui.SetObexConn(obexConn)
	ui.SetNetworkConn(networkConn)

	ui.StartUI()

	ui.StopMediaPlayer()

	agent.RemoveObexAgent()
	agent.RemoveAgent()
}
