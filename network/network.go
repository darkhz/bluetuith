package network

import (
	"sync"

	nm "github.com/Wifx/gonetworkmanager"
	"github.com/google/uuid"
)

// Network holds the network manager and active connections.
type Network struct {
	Manager nm.NetworkManager

	ActiveConnection map[string]nm.ActiveConnection
	connectionLock   sync.Mutex
}

// NewNetwork returns a new Network.
func NewNetwork() (*Network, error) {
	manager, err := nm.NewNetworkManager()
	if err != nil {
		return nil, err
	}

	network := &Network{
		Manager:          manager,
		ActiveConnection: make(map[string]nm.ActiveConnection),
	}

	return network, nil
}

// Connect connects to the device's network interface.
func (n *Network) Connect(name, connType, bdaddr string) error {
	active, err := n.IsConnectionActive(connType, bdaddr)
	if err != nil {
		return err
	}
	if active {
		return NMConnectionAlreadyActive
	}

	activated, err := n.ActivateExistingConnection(connType, bdaddr)
	if err != nil {
		return err
	}
	if activated {
		return nil
	}

	return n.CreateConnection(name, connType, bdaddr)
}

// IsConnectionActive checks if the device's connection is active.
func (n *Network) IsConnectionActive(connType, bdaddr string) (bool, error) {
	activeConnections, err := n.Manager.GetPropertyActiveConnections()
	if err != nil {
		return false, err
	}

	for _, activeConn := range activeConnections {
		ctype, err := activeConn.GetPropertyType()
		if err != nil {
			return false, err
		}
		if ctype != "bluetooth" {
			continue
		}

		conn, err := activeConn.GetPropertyConnection()
		if err != nil {
			return false, err
		}

		exist, err := isDeviceAddrExist(conn, connType, bdaddr)
		if err != nil {
			return false, err
		}
		if exist {
			return true, nil
		}
	}

	return false, nil
}

// ActivateExistingConnection activates an existing device connection profile.
func (n *Network) ActivateExistingConnection(connType, bdaddr string) (bool, error) {
	devices, err := n.Manager.GetPropertyDevices()
	if err != nil {
		return false, err
	}

	for _, device := range devices {
		dtype, err := device.GetPropertyDeviceType()
		if err != nil {
			return false, err
		}
		if dtype != nm.NmDeviceTypeBt {
			continue
		}

		conns, err := device.GetPropertyAvailableConnections()
		if err != nil {
			return false, err
		}

		for _, conn := range conns {
			exist, err := isDeviceAddrExist(conn, connType, bdaddr)
			if err != nil {
				return false, err
			}
			if exist {
				err := checkSettings(conn, connType)
				if err != nil {
					return false, err
				}

				return true, n.ActivateConnection(conn, device, bdaddr)
			}
		}
	}

	return false, nil
}

// CreateConnection creates a new connection.
func (n *Network) CreateConnection(name, connType, bdaddr string) error {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	bdaddrBytes, err := getBDAddress(bdaddr)
	if err != nil {
		return err
	}

	connectionSettings := getSettings(bdaddrBytes, name, connType, newUUID.String())

	settings, err := nm.NewSettings()
	if err != nil {
		return err
	}

	conn, err := settings.AddConnection(connectionSettings)
	if err != nil {
		return err
	}

	device, err := n.Manager.GetDeviceByIpIface(bdaddr)
	if err != nil {
		return err
	}

	return n.ActivateConnection(conn, device, bdaddr)
}

// ActivateConnection activates the connection.
func (n *Network) ActivateConnection(conn nm.Connection, device nm.Device, bdaddr string) error {
	var state nm.StateChange

	activeConn, err := n.Manager.ActivateConnection(conn, device, nil)
	if err != nil {
		return err
	}

	exit := make(chan struct{})
	activeState := make(chan nm.StateChange)

	err = activeConn.SubscribeState(activeState, exit)
	if err != nil {
		return err
	}

	n.connectionLock.Lock()
	n.ActiveConnection[bdaddr] = activeConn
	n.connectionLock.Unlock()

WatchState:
	for {
		select {
		case state = <-activeState:
			if state.State == nm.NmActiveConnectionStateActivating {
				continue
			}

			exit <- struct{}{}
			break WatchState
		}
	}

	if state.State != nm.NmActiveConnectionStateActivated {
		return NMConnectionError
	}

	return nil
}

// DeactivateConnection deactivates the connection.
func (n *Network) DeactivateConnection(bdaddr string) error {
	var activeConn nm.ActiveConnection

	n.connectionLock.Lock()
	activeConn = n.ActiveConnection[bdaddr]
	delete(n.ActiveConnection, bdaddr)
	n.connectionLock.Unlock()

	if activeConn == nil {
		return nil
	}

	return n.Manager.DeactivateConnection(activeConn)
}
