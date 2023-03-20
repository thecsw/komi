package komi

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
)

// New creates a new pool with sensible defaults.
func New[I, O any](optionWork poolWork[I, O]) *Pool[I, O] {
	return NewWithSettings(optionWork, nil)
}

// NewWithSettings creates a new pool with custom pool tunings enabled.
func NewWithSettings[I, O any](optionWork poolWork[I, O], settings *Settings) *Pool[I, O] {
	p := &Pool[I, O]{
		settings:            settings,
		jobsWaiting:         &atomic.Int64{},
		jobsCompleted:       &atomic.Int64{},
		tellChildrenToClose: make(chan Signal),
		closedSignal:        make(chan Signal, 1),
		log: log.NewWithOptions(os.Stderr, log.Options{
			TimeFormat:      time.DateTime,
			ReportTimestamp: true,
			ReportCaller:    false,
		}),
		currentlyWaitingForJobs:      &atomic.Bool{},
		noJobsCurrentlyWaitingSignal: make(chan Signal),
	}

	// Run the function to set the work performer for the pool.
	optionWork(p)

	// If work received and set is a nil function, then immediately panic.
	if !p.hasWork() {
		panic("pool didn't receive any work")
	}

	// Verify that all settings have been correctly set after work has been confirmed.
	if p.settings == nil {
		p.settings = &Settings{}
	}
	verifySettings(p.settings)

	// Set the logging levels and options.
	p.log.SetLevel(p.settings.LogLevel)
	p.log.SetPrefix(p.settings.Name)

	// If usar has not provided a manual size setting, then set `size = laborers * ratio`.
	if !p.settings.sizeOverride {
		p.settings.Size = p.settings.Laborers * p.settings.Ratio
	}

	// A nice debug.
	p.log.Debug("Pool settings initialized")

	// Allocate the channel with proper size.
	p.inputs = make(chan I, p.settings.Size)

	// If the function given produces outputs, also allocate the outputs channel.
	if p.producesOutputs() {
		p.outputs = make(chan O, p.settings.Size)
	}

	// If the function given produces errors, also allocated the errors channel.
	if p.producesErrors() {
		p.errors = make(chan PoolError[I], p.settings.Size)
	}

	// Fire off all the laborers.
	p.startLaborers()

	return p
}

// SetLevel the logging level of the pool.
func (p *Pool[_, _]) SetLevel(level log.Level) {
	p.log.SetLevel(level)
}

// Debug enables the debug logging in the pool.
func (p *Pool[_, _]) Debug() {
	p.log.SetLevel(log.DebugLevel)
}
