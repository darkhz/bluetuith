package ui

// watchEvent listens to DBus events and passes them to
// the event handlers.
func watchEvent() {
	watchSignal := UI.Bluez.WatchSignal()
	defer UI.Bluez.Conn().RemoveSignal(watchSignal)

	for signal := range watchSignal {
		signalData := UI.Bluez.ParseSignalData(signal)

		adapterEvent(signal, signalData)
		deviceEvent(signal, signalData)
	}
}
