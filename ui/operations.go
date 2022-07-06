package ui

import "sync"

var (
	opLock     sync.Mutex
	cancelFunc func()
)

// startOperation sets up the cancellation handler,
// and starts the operation.
func startOperation(dofunc, cancel func()) {
	opLock.Lock()
	defer opLock.Unlock()

	if cancelFunc != nil {
		InfoMessage("Operation still in progress", false)
		return
	}

	cancelFunc = cancel

	go func() {
		dofunc()
		cancelOperation(false)
	}()
}

// cancelOperation cancels the currently running operation.
func cancelOperation(cancelfunc bool) {
	var cancel func()

	opLock.Lock()
	defer opLock.Unlock()

	if cancelFunc == nil {
		return
	}

	cancel = cancelFunc
	cancelFunc = nil

	if cancelfunc {
		go cancel()
	}
}
