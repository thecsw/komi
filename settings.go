package komi

import (
	"github.com/charmbracelet/log"
)

// poolSettings is an internal struct that tunes the pool as requested.
type poolSettings struct {
	// numLaborers is the number of laborers (work performers) that should
	// run in parallel.
	numLaborers int

	// size sets the size of the pool, or how many inputs and outputs (each
	// has separate size) can be set by the pool. If size is reached, submissions
	// or work results will be blocked.
	size int

	// sizeOverride is true if the user chose to set the size manuall, instead of
	// letting the pool set it automatically.
	sizeOverride bool

	// ratioSizeToNumLaborers is the ratio that is used (unless size is set manually)
	// for setting the size to the number of laborers.
	ratioSizeToNumLaborers int

	// debug will set the logger to show all debug logs too.
	debug bool

	// name is the name of the pool.
	name string

	// logLevel defaults to warn, can be set by the user.
	logLevel log.Level
}

// PoolSettingsFunc is the function type for binding custom settings.
type PoolSettingsFunc func(*poolSettings)

// WithLaborers sets the number of laborers to activate in the pool,
// will default to number of logical CPU cores if less or equal to 0.
func WithLaborers(num int) PoolSettingsFunc {
	return func(ps *poolSettings) {
		if num <= 0 {
			return
		}
		ps.numLaborers = num
	}
}

// WithSize sets the size of the pool, of how many jobs can wait at any
// moment's time, defaults to `ratio * number of laborers`
func WithSize(size int) PoolSettingsFunc {
	return func(ps *poolSettings) {
		if size <= 0 {
			return
		}
		ps.size = size
		ps.sizeOverride = true
	}
}

// WithSizeToLaborersRatio sets the `ratio` in `size = ratio * laborers` equation,
// unless `size` has been manually set with `WithSize`.
func WithSizeToLaborersRatio(ratio int) PoolSettingsFunc {
	return func(ps *poolSettings) {
		if ratio <= 0 {
			return
		}
		ps.ratioSizeToNumLaborers = ratio
	}
}

// WithDebug will set the log level to `DebugLevel`.
func WithDebug() PoolSettingsFunc {
	return WithLogLevel(log.DebugLevel)
}

// WithName sets the name of the pool, which is shown in logs.
func WithName(name string) PoolSettingsFunc {
	return func(ps *poolSettings) {
		ps.name = name
	}
}

// WithLogLevel sets the logging level of the pool's internals,
// defaults to `WarnLevel`.
func WithLogLevel(level log.Level) PoolSettingsFunc {
	return func(ps *poolSettings) {
		ps.logLevel = level
	}
}
