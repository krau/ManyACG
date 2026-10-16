//go:build unix

package internalrpc

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func sshStateOpenFlags() int {
	return unix.O_NOFOLLOW | unix.O_NONBLOCK
}

func safeSSHStateDirectory(info os.FileInfo, final bool) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || (stat.Uid != 0 && stat.Uid != uint32(os.Geteuid())) {
		return false
	}
	// A sticky ancestor such as /tmp is safe; the state directory itself is not shared.
	return info.Mode().Perm()&0o022 == 0 || (!final && info.Mode()&os.ModeSticky != 0)
}

func safeSSHStateFile(info os.FileInfo, private, writable bool) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Nlink != 1 || (stat.Uid != 0 && stat.Uid != uint32(os.Geteuid())) {
		return false
	}
	if private {
		return info.Mode().Perm() == 0o600
	}
	return info.Mode().Perm()&0o022 == 0 && (!writable || info.Mode().Perm()&0o200 != 0)
}

func trySSHStateLock(file *os.File) (bool, error) {
	lock := unix.Flock_t{Type: unix.F_WRLCK, Whence: int16(0), Start: 0, Len: 1}
	err := unix.FcntlFlock(file.Fd(), unix.F_SETLK, &lock)
	if errors.Is(err, unix.EACCES) || errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EINTR) {
		return false, nil
	}
	return err == nil, err
}

func syncSSHStateDirectory(root *os.Root) error {
	file, err := root.Open(".")
	if err != nil {
		return fmt.Errorf("open internal SSH state directory for synchronization: %w", err)
	}
	defer file.Close()
	if err := file.Sync(); err != nil {
		return fmt.Errorf("synchronize internal SSH state directory: %w", err)
	}
	return nil
}
