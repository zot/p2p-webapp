//go:build windows

package pidfile

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32     = syscall.NewLazyDLL("kernel32.dll")
	lockFileEx   = kernel32.NewProc("LockFileEx")
	unlockFileEx = kernel32.NewProc("UnlockFileEx")
)

const (
	lockfileExclusiveLock = 0x00000002
)

// lockFile locks a file using Windows LockFileEx
func lockFile(file *os.File) error {
	// Lock the entire file
	ol := syscall.Overlapped{}
	r1, _, err := lockFileEx.Call(
		uintptr(file.Fd()),
		uintptr(lockfileExclusiveLock),
		0,
		1, 0,
		uintptr(unsafe.Pointer(&ol)),
	)
	if r1 == 0 {
		return fmt.Errorf("failed to lock file: %w", err)
	}
	return nil
}

// unlockFile unlocks a file using Windows UnlockFileEx
func unlockFile(file *os.File) error {
	ol := syscall.Overlapped{}
	r1, _, err := unlockFileEx.Call(
		uintptr(file.Fd()),
		0,
		1, 0,
		uintptr(unsafe.Pointer(&ol)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}
