//go:build !linux && !windows

package update

import (
	"errors"
	"os"
)

func lockFile(file *os.File) error {
	return errors.New("update locking is supported on Linux and Windows")
}

func syncReportDirectories(directory string) error { return nil }
