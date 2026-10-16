package internalrpc

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

func sshStateOpenFlags() int { return 0 }

// Windows enforces directory/file access with inherited ACLs, not Unix mode bits.
func safeSSHStateDirectory(os.FileInfo, bool) bool  { return true }
func safeSSHStateFile(os.FileInfo, bool, bool) bool { return true }

func trySSHStateLock(file *os.File) (bool, error) {
	err := windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{})
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) || errors.Is(err, windows.ERROR_IO_PENDING) {
		return false, nil
	}
	return err == nil, err
}

// File.Sync completes before publishing; Windows does not support directory fsync.
func syncSSHStateDirectory(*os.Root) error { return nil }
