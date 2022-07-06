package ui

// watchEvent listens to DBus events and passes them to
// the event handlers.
func watchEvent() {
	for signal := range BluezConn.WatchSignal() {
		signalData := BluezConn.ParseSignalData(signal)

		adapterEvent(signal, signalData)
		deviceEvent(signal, signalData)
	}
}
