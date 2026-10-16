package update

import (
	"os"
	"path/filepath"
	"syscall"
)

func lockFile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func syncReportDirectories(directory string) error {
	for i := 0; i < 3; i++ {
		file, err := os.Open(directory)
		if err != nil {
			return err
		}
		err = file.Sync()
		closeErr := file.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		directory = filepath.Dir(directory)
	}
	return nil
}
