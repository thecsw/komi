package komi

import (
	"github.com/charmbracelet/log"
)

// Settings is an internal struct that tunes the pool as requested.
type Settings struct {
	// Laborers is the number of laborers (work performers) that should
	// run in parallel.
	Laborers int

	// Size sets the Size of the pool, or how many inputs and outputs (each
	// has separate Size) can be set by the pool. If Size is reached, submissions
	// or work results will be blocked.
	Size int

	// sizeOverride is true if the user chose to set the size manually, instead of
	// letting the pool set it automatically.
	sizeOverride bool

	// Ratio is the ratio that is used (unless size is set manually)
	// for setting the size to the number of laborers.
	Ratio int

	// Debug will set the logger to show all Debug logs too.
	Debug bool

	// Name is the Name of the pool.
	Name string

	// LogLevel defaults to warn, can be set by the user.
	LogLevel log.Level
}

// verifySettings will make sure the settings are proper and
// set sensible defaults if user hasn't set them manually.
func verifySettings(settings *Settings) {
	// If laborers have not been set, default to number of CPUs.
	if settings.Laborers <= 0 {
		settings.Laborers = defaultNumLaborers
	}
	// If the user has manually set the size, that shall be used.
	if settings.Size > 0 {
		settings.sizeOverride = true
	}
	// If ratio is default, set the default size ratio.
	if settings.Ratio <= 0 {
		settings.Ratio = defaultRatio
	}
	// If size has not been manually set, default to `size = ratio * laborers`
	if !settings.sizeOverride {
		settings.Size = settings.Ratio * settings.Laborers
	}
	// If log level is default, set it to at least to warn level.
	if settings.LogLevel <= 0 {
		settings.LogLevel = log.WarnLevel
	}
	// If debug is set, set the log level to debug.
	if settings.Debug {
		settings.LogLevel = log.DebugLevel
	}
	// If name is empty, set it to default.
	if len(settings.Name) < 1 {
		settings.Name = defaultName
	}
}
