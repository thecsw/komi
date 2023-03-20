package komi

import (
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
)

// NewPool will allocate a new pool with a required work function and optional settings
// with sensible defaults.
func NewPool[I, O any](optionWork poolWork[I, O], options ...PoolSettingsFunc) *Pool[I, O] {
	// Create the new pool entity with some sensible defaults.
	p := &Pool[I, O]{
		settings: &poolSettings{
			numLaborers:            defaultNumLaborers,
			size:                   defaultNumLaborers,
			ratioSizeToNumLaborers: defaultRatio,
			waitingDelay:           time.Millisecond,
			name:                   defaultName,
			// Show errors by default from logging.
			logLevel: log.WarnLevel,
		},
		jobsWaiting:              &atomic.Int64{},
		jobsCompleted:            &atomic.Int64{},
		closureSignalForChildren: make(chan Signal, 1),
		closedSignal:             make(chan Signal, 1),
		log: log.New(
			log.WithTimestamp(),
			log.WithTimeFormat(time.DateTime),
		),
	}

	// Run the function to set the work performer for the pool.
	optionWork(p)

	// If work received and set is a nil function, then immediately panic.
	if !p.hasWork() {
		panic("pool didn't receive any work")
	}

	// Run all of the optional settings' functions and set user's settings.
	for _, option := range options {
		option(p.settings)
	}

	// Set the logging levels and options.
	p.log.SetLevel(p.settings.logLevel)
	p.log.SetPrefix(p.settings.name)

	// If usar has not provided a manual size setting, then set `size = laborers * ratio`.
	if !p.settings.sizeOverride {
		p.settings.size = p.settings.numLaborers * p.settings.ratioSizeToNumLaborers
	}

	// A nice debug.
	p.log.Debug("Pool settings initialized")

	// Allocate the channel with proper size.
	p.inputs = make(chan I, p.settings.size)

	// If the function given produces outputs, also allocate the outputs channel.
	if p.producesOutputs() {
		p.outputs = make(chan O, p.settings.size)
	}

	// If the function given produces errors, also allocated the errors channel.
	if p.producesErrors() {
		p.errors = make(chan PoolError[I], p.settings.size)
	}

	// Fire off all the laborers.
	p.startLaborers()

	return p
}
