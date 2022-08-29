package ui

// watchEvent listens to DBus events and passes them to
// the event handlers.
func watchEvent() {
	watchSignal := BluezConn.WatchSignal()
	defer BluezConn.Conn().RemoveSignal(watchSignal)

	for signal := range watchSignal {
		signalData := BluezConn.ParseSignalData(signal)

		adapterEvent(signal, signalData)
		deviceEvent(signal, signalData)
	}
}
