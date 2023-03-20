package komi

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/charmbracelet/log"
)

// Signal is a cheap type to be sent through channels.
type Signal struct{}

const (
	// defaultName is the unnamed's pool's name.
	defaultName = "Komi 🍡 "

	// defaultRatio sets the size to laborers ratio.
	defaultRatio = 2
)

var (
	// signal is the implementation of `Signal`.
	signal = struct{}{}

	// defaultNumLaborers defaults to the number of CPU (logical) cores.
	defaultNumLaborers = runtime.NumCPU()
)

// Pool is a fantastic golang pool that can take work of any form and
// perform it in a go-idiomatic way of producing outputs (optional) and
// errors (optional).
type Pool[I, O any] struct {
	// settings is the typeless configuration of the pool.
	settings *poolSettings

	// log is the pool's logger to be used.
	log log.Logger

	// workSimple could be set by the user if the kind of work
	// they want the pool to perform has no outputs or errors.
	workSimple func(I)

	// workSimpleWithErrors could be set by the user if the kind
	// of work they want the pool to perform only produces errors.
	workSimpleWithErrors func(I) error

	// workRegular could be set by the user if the kind of work
	// they want the pool to perform produces outputs with no errors.
	workRegular func(I) O

	// workRegularWithErrors could be set by the user if the kind of
	// work they want the pool to perform produces outputs and errors.
	workRegularWithErrors func(I) (O, error)

	// workPerformer is a function signature that will be set to
	// whatever work that the user gave for the pool.
	workPerformer func(I)

	// tellChildrenToClose is a channel where this pool will send a
	// signal to tell all the dependent (child) pools (the ones that send
	// their outputs to here) to start shutting down.
	tellChildrenToClose chan Signal

	// closed will be set to true when the pool is fully closed.
	closed bool

	// closedSignal is a channel that is set by dependent (child) channels,
	// so they can tell their connected (parent) pools that they are closed.
	// This is because the parent will close if and only if ALL their dependent
	// (child) pools have closed.
	closedSignal chan Signal

	// inputs channel is where the jobs are coming from.
	inputs chan I

	// outputs channel is where `workPerformer` will send jobs' outputs (if work
	// is at least "Regular") to.
	outputs chan O

	// errors channel is where `workPerformer` will send jobs' errors (if work has
	// an error return) to, unless, a user supplied an error handler func, which will
	// immediately consume the element (or not push it in the first place).
	errors chan PoolError[I]

	// jobsWaiting is an atomic counter used to count to how many jobs are currently
	// waiting in the `inputs` channel AND the number of jobs that are currently
	// performing work.
	jobsWaiting *atomic.Int64

	// jobsCompleted is an atomic counter used to count how many jobs have performed work
	// (it doesn't if the errors are enabled and return a non-nil result).
	jobsCompleted *atomic.Int64

	// laborersStopSignal is a channel used by the pool to tell all laborers to quit,
	// consumed by laborers.
	laborersStopSignal chan Signal

	// laborersActive is a wait group to block a closure request until all laborers
	// have gracefully quit.
	laborersActive *sync.WaitGroup

	// connectorsStopSignal is a channel used by the pool to tell all connectors to quit,
	// consumed by connectors.
	connectorsStopSignal chan Signal

	// connectorsActive is a wait group to block a closure request until all connectors
	// have gracefully quit.
	connectorsActive *sync.WaitGroup

	// connectorRequestedClosure is set to true if the closure request has been supplied
	// by one of the connectors, this will happen if the connected (parent) pool has let this
	// dependent (child) pool know that its closing, therefore the child should also shutdown.
	connectorRequestedClosure bool

	// childsClosureSignal is a back-channel that is used by the child to tell the parent that the
	// child left, therefore continuing parent's active closure request.
	childsClosureSignal <-chan Signal

	// parent is a handle that the child can use to communicate with its parent.
	parent PoolConnector[O]

	// childsWait is dependent (child) pool's waiting function.
	childsWait func()
}

// PoolError is produced by the pool when a work performed by the pool fails
// with a non-nil error, so the user can debug and look what happened.
type PoolError[I any] struct {
	// Job is the job that returned a non-nil error when work was performed on it.
	Job I

	// Error is the error returned by pool's work performer.
	Error error
}
