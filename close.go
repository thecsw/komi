package komi

// signalForChildren will have a signal sent when this pool
// is getting closed. Use this for children to know when the
// parent is leaving.
func (p Pool[_, _]) signalForChildren() <-chan Signal {
	return p.tellChildrenToClose
}

// IsClosed returns true if the pool is closed, false otherwise.
func (p Pool[_, _]) IsClosed() bool {
	return p.closed
}

// anotherPoolIsSendingJobsHere return true if another pool is feeding
// jobs into this pool, false otherwise.
func (p Pool[_, _]) anotherPoolIsSendingJobsHere() bool {
	return p.childsClosureSignal != nil
}

// Close will issue a pool closure request and takes a bool value, if true,
// any pending jobs will be ignored and forcefully closed. Note that the user
// can request a pool closure if and only if it is not connected to another
// pool. In that case, the parent pool will have to issue the closure request.
func (p *Pool[_, _]) Close(force ...bool) {
	p.closureInternalWait.Add(1)
	p.closureRequest <- (len(force) > 0 && force[0])
	p.closureInternalWait.Wait()
}

func (p *Pool[_, _]) closureRequestListener() {
waiting:
	// Block until a request comes in
	forced := <-p.closureRequest
	p.log.Debug("Got a request to close", "forced", forced, "parent_requested", p.connectorRequestedClosure)
	// Refuse to close if it had already been done.
	if p.IsClosed() {
		p.log.Warn("Pool is already closed")
		p.closureInternalWait.Done()
		goto waiting
	}
	if p.IsConnected() && !p.connectorRequestedClosure {
		p.log.Warn("Only the parent can close this pool", "parent", p.parent.Name())
		p.closureInternalWait.Done()
		goto waiting
	}

	if p.childsWait != nil && !forced {
		p.log.Debug("Waiting for the child's Wait")
		p.childsWait()
	}

	// This is a flag that will force closure (override waiting).
	shouldForceNonetheless := false

	// Wait until the child pools are closed.
	if p.anotherPoolIsSendingJobsHere() {
		// Send the closed signal to any connected pools. We need to issue a closure
		// request to the dependent (child) pools before locking ourselves (optionally)
		// and waiting for those dependent (child) pools to leave.
		p.log.Info("Sending a signal for the child to leave...")
		p.tellChildrenToClose <- signal
		<-p.childsClosureSignal
		p.log.Info("Child left, resuming closure...")
		close(p.tellChildrenToClose)

		shouldForceNonetheless = true
		drain(p.inputs)
		drain(p.outputs)
	}

	// If force flag has been passed, wait until no jobs are waiting.
	if forced && !shouldForceNonetheless {
		p.Wait()
	}

	// Start sending a signal for all laborers to quit.
	p.stopLaborers()

	// If we the pool had a connector enabled, close it down too.
	if p.IsConnected() {
		p.log.Debug("Sent a shutdown signal to connectors...")
		p.connectorsStopSignal <- signal
		p.connectorsActive.Wait()
		close(p.connectorsStopSignal)
		p.log.Debug("Connectors quit")
	}

	// Close the inputs channel so no new work is processed.
	drain(p.inputs)
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

	// Mark the flag that the pool is closed.
	p.closed = true

	// I like Internet Historian.
	p.log.Debug("Pool is closed", "completed", p.JobsCompleted())

	p.closedSignal <- signal
	close(p.closedSignal)

	if !p.connectorRequestedClosure {
		p.closureInternalWait.Done()
	}
}
