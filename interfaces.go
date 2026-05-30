// Package hotloader provides hot-reloading of configuration files at runtime
// via OS signals. It watches for signals (e.g., SIGUSR1) and automatically
// re-reloads a configuration file into memory when received.
package hotloader

// Loader defines the interface that configuration types must implement
// to be hot-reloaded by the library.
type Loader interface {
	// Load parses the configuration file at the given filename.
	// It returns an error if the file cannot be read or parsed.
	Load(filename string) error

	// InitDefault initializes the configuration with default values.
	// This is called before Load to ensure a clean state.
	InitDefault()
}

// HotLoader manages the hot-reloading lifecycle of a configuration.
// It watches for OS signals and reloads the configuration when triggered.
type HotLoader interface {
	// Load manually triggers a configuration reload from the file.
	// It returns an error if the reload fails.
	Load() error

	// Get returns the current loaded configuration in a thread-safe manner.
	Get() Loader

	// SetBeforeFunc sets a callback that is called before each reload.
	// The callback receives the current configuration being replaced.
	SetBeforeFunc(f func(l Loader))

	// SetAfterFunc sets a callback that is called after a successful reload.
	// The callback receives the new configuration.
	SetAfterFunc(f func(l Loader))

	// SetErrorFunc sets a callback that is called when a reload fails.
	// The callback receives the error that occurred.
	SetErrorFunc(f func(e error))
}
