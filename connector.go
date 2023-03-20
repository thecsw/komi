package komi

import "sync"

// PoolConnector is an interface that should be used by other pools
// when connecting to them and by users to send pools around as well.
type PoolConnector[O any] interface {
	// Submit will submit a job to the connected (parent) pool.
	Submit(O)

	// signalForChildren will have a signal go through it when the
	// connected (parent) pool is closing, therefore, letting know
	// all the children pools that they should themselves close.
	signalForChildren() <-chan Signal

	// waitBeforeClosure will force the connected (parent) pool to
	// wait for a signal from this channel before proceeding with
	// a closure request.
	waitBeforeClosure(<-chan Signal)

	// setChildsWait is useful for parents gracefully waiting for
	// their children to wrap up work.
	setChildsWait(func())

	// IsClosed returns true if the connected (parent) pool is closed,
	// false otherwise.
	IsClosed() bool

	// Name returns the name of the connected (parent) pool.
	Name() string
}

func (p *Pool[I, O]) Connect(parent PoolConnector[O]) {
	// This should not trigger, because `noValue` is a package internal,
	// so it shouldn't be accessible to the user to connect outputless
	// pools to other pools. Consider this as a last defense line.
	if !p.producesOutputs() {
		p.log.Warn("Can't connect because not producing outputs.")
		return
	}

	// This pool is already sending its outputs to a connected (parent)
	// pool, therefore, refuse this connection request.
	if p.IsConnected() {
		p.log.Warn("A connector is already running.")
		return
	}

	// Create a wait group that will let us know if there are any
	// running connectors in this pool.
	p.connectorsActive = &sync.WaitGroup{}

	// Create a connector stop signal that will be used to tell connector(s)
	// to quit their execution.
	p.connectorsStopSignal = make(chan Signal, 1)

	// Set the connected (parent) pool.
	p.parent = parent

	// Tell the connected (parent) pool to wait for this dependent (child)
	// pool's closure before they can close themselves.
	p.parent.waitBeforeClosure(p.closedSignal)

	// Set child's wait.
	p.parent.setChildsWait(p.Wait)

	// Kick off the connector.
	go func(p *Pool[I, O]) {
		for {
			select {
			case result := <-p.outputs:
				// If the pool produced a new output, grab it and send it
				// as a new job to the connected pool.
				parent.Submit(result)
				// ---
			case <-p.connectorsStopSignal:
				// If the stop connector signal received, mark this connector
				// as done and kill the scope.
				p.connectorsActive.Done()
				return
			case <-p.parent.signalForChildren():
				// If the target pool is closed, this pool should also get
				// automatically closed, as no one would be continuing to
				// consume this pool's outputs.
				p.log.Info("Closing because the parent pool is leaving...", "parent", p.parent.Name())

				// Mark this flag, so the closure subroutine doesn't hang until
				// this connector responds back, because it is the one, which
				// requested closure, and not the user.
				p.connectorRequestedClosure = true

				// Request the closure of this pool.
				p.Close()
				return
			}
		}
	}(p)

	// Mark this new connector as a running instance.
	p.connectorsActive.Add(1)

	// Log the connected (parent) pool.
	p.log.Info("Connected to the parent pool", "parent", p.parent.Name())
}

// IsConnected will return true if this pool already has an active connector.
// This is equivalent to having a connected (parent) pool.
func (p Pool[_, _]) IsConnected() bool {
	return p.parent != nil
}

// waitBeforeClosure will force the pool to wait for a signal from this channel
// before it can proceed with a closure request.
func (p *Pool[_, _]) waitBeforeClosure(waitForThis <-chan Signal) {
	p.childPoolLeft = waitForThis
}

// Name returns the name of the pool.
func (p *Pool[_, _]) Name() string {
	return p.settings.name
}

// setChildsWait sets the child's wait function.
func (p *Pool[_, _]) setChildsWait(childWait func()) {
	p.childsWait = childWait
}
