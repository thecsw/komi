package komi

// isWorkSimple returns true if the work produces no outputs nor errors.
func (p *Pool[_, _]) isWorkSimple() bool { return p.workSimple != nil }

// isWorkSimpleWithErrors returns true if the work produces errors but no outputs.
func (p *Pool[_, _]) isWorkSimpleWithErrors() bool { return p.workSimpleWithErrors != nil }

// isWorkRegular returns true if the work produces outputs but no errors.
func (p *Pool[_, _]) isWorkRegular() bool { return p.workRegular != nil }

// isWorkRegularWithErrors returns true if the work produces outputs and errors.
func (p *Pool[_, _]) isWorkRegularWithErrors() bool { return p.workRegularWithErrors != nil }

// hasWork returns true work has been set and is non-nil.
func (p *Pool[_, _]) hasWork() bool {
	return p.isWorkSimple() || p.isWorkSimpleWithErrors() || p.isWorkRegular() || p.isWorkRegularWithErrors()
}

// producesOutputs returns true if the work produces outputs.
func (p *Pool[_, _]) producesOutputs() bool {
	return p.isWorkRegular() || p.isWorkRegularWithErrors()
}

// producesErrors returns true if the work produces errors.
func (p *Pool[_, _]) producesErrors() bool {
	return p.isWorkSimpleWithErrors() || p.isWorkRegularWithErrors()
}

// performWorkSimple will perform the simple work.
func (p *Pool[I, _]) performWorkSimple(job I) {
	defer p.performedWork(true)
	p.workSimple(job)
}

// performWorkSimpleWithErrors will perform simple work with errors.
func (p *Pool[I, _]) performWorkSimpleWithErrors(job I) {
	err := p.workSimpleWithErrors(job)
	if err != nil {
		p.errors <- PoolError[I]{
			Job:   job,
			Error: err,
		}
		p.performedWork(false)
		return
	}
	p.performedWork(true)
}

// performWorkRegular will perform regular work.
func (p *Pool[I, O]) performWorkRegular(job I) {
	defer p.performedWork(true)
	p.outputs <- p.workRegular(job)
}

// performWorkWithErrors will perform regular work with errors.
func (p *Pool[I, O]) performWorkWithErrors(job I) {
	res, err := p.workRegularWithErrors(job)
	if err != nil {
		p.errors <- PoolError[I]{
			Job:   job,
			Error: err,
		}
		p.performedWork(false)
		return
	}
	p.outputs <- res
	p.performedWork(true)
}

// performedWork will reduce the number of waiting jobs and increase
// the number of completed jobs.
func (p *Pool[_, _]) performedWork(success bool) {
	p.jobsWaiting.Add(-1)
	p.jobsCompleted.Add(1)
	if success {
		p.jobsSucceeded.Add(1)
	}

	// If we are concurrently waiting for jobs, send a signal
	// if the number of waiting jobs is 0.
	if p.jobsWaiting.Load() < 1 && p.currentlyWaitingForJobs.Load() {
		p.noJobsCurrentlyWaitingSignal <- signal
	}
}
