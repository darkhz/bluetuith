package network

import "errors"

var (
	NMConnectionAlreadyActive = errors.New("Connection is already active")
	NMConnectionError         = errors.New("Connection error occurred")

	NMSettingModifyError = errors.New("Cannot modify connection settings")
)
