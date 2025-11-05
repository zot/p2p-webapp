//go:build unix

package pidfile

import (
	"fmt"
	"os"
	"syscall"
)

// lockFile locks a file using Unix flock
func lockFile(file *os.File) error {
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("failed to lock file: %w", err)
	}
	return nil
}

// unlockFile unlocks a file using Unix flock
func unlockFile(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
