package ui

import "sync"

type Operation struct {
	cancel func()

	lock sync.Mutex
}

var operation Operation

// startOperation sets up the cancellation handler,
// and starts the operation.
func startOperation(dofunc, cancel func()) {
	operation.lock.Lock()
	defer operation.lock.Unlock()

	if operation.cancel != nil {
		InfoMessage("Operation still in progress", false)
		return
	}

	operation.cancel = cancel

	go func() {
		dofunc()
		cancelOperation(false)
	}()
}

// cancelOperation cancels the currently running operation.
func cancelOperation(cancelfunc bool) {
	var cancel func()

	operation.lock.Lock()
	defer operation.lock.Unlock()

	if operation.cancel == nil {
		return
	}

	cancel = operation.cancel
	operation.cancel = nil

	if cancelfunc {
		go cancel()
	}
}
