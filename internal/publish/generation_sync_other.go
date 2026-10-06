//go:build !linux

package publish

// Directory fsync and production atomic-rename guarantees are Linux-specific.
func syncGeneration(path string) error          { return nil }
func syncActivationDirectory(path string) error { return nil }
