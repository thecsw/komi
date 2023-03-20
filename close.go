package komi

// signalForChildren will have a signal sent when this pool
// is getting closed. Use this for children to know when the
// parent is leaving.
func (p Pool[_, _]) signalForChildren() <-chan Signal {
	return p.closureSignalForChildren
}

// IsClosed returns true if the pool is closed, false otherwise.
func (p Pool[_, _]) IsClosed() bool {
	return p.closed
}

// anotherPoolIsSendingJobsHere return true if another pool is feeding
// jobs into this pool, false otherwise.
func (p Pool[_, _]) anotherPoolIsSendingJobsHere() bool {
	return p.childPoolLeft != nil
}

// Close will issue a pool closure request and takes a bool value, if true,
// any pending jobs will be ignored and forcefully closed. Note that the user
// can request a pool closure if and only if it is not connected to another
// pool. In that case, the parent pool will have to issue the closure request.
func (p *Pool[_, _]) Close(force ...bool) {
	// Refuse to close if it had already been done.
	if p.IsClosed() {
		p.log.Warn("Pool is already closed")
		return
	}
	if p.IsConnected() && !p.connectorRequestedClosure {
		p.log.Warn("Only the parent can close this pool", "parent", p.parent.Name())
		return
	}

	// This is a flag that will force closure (override waiting).
	shouldForceNonetheless := false

	// Wait until the child pools are closed.
	if p.anotherPoolIsSendingJobsHere() {
		// Send the closed signal to any connected pools. We need to issue a closure
		// request to the dependent (child) pools before locking ourselves (optionally)
		// and waiting for those dependent (child) pools to leave.
		p.closureSignalForChildren <- signal

		p.log.Info("Waiting for child pools to close...")
		<-p.childPoolLeft
		p.log.Info("Child left, resuming closure...")
		close(p.closureSignalForChildren)

		shouldForceNonetheless = true
		drain(p.inputs)
		drain(p.outputs)
	}

	// If force flag has been passed, wait until no jobs are waiting.
	if (len(force) <= 0 || !force[0]) && !shouldForceNonetheless {
		p.Wait()
	}

	// Start sending a signal for all laborers to quit.
	p.stopLaborers()

	// Close the inputs channel so no new work is processed.
	close(p.inputs)

	// If we have been writing outputs, close the channel.
	if p.producesOutputs() {
		drain(p.outputs)
		close(p.outputs)
	}

	// If we have been writing errors, close the channel.
	if p.producesErrors() {
		drain(p.errors)
		close(p.errors)
	}

	// If we the pool had a connector enabled, close it down too.
	if p.IsConnected() && !p.connectorRequestedClosure {
		p.log.Debug("Sent a shutdown signal to connectors...")
		p.connectorsStopSignal <- signal
		p.connectorsActive.Wait()
		close(p.connectorsStopSignal)
		p.log.Debug("Connectors quit")
	}

	// Mark the flag that the pool is closed.
	p.closed = true
	p.closedSignal <- signal

	// I like Internet Historian.
	p.log.Debug("Pool is closed", "completed", p.JobsCompleted())
}
