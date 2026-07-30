//go:build !windows

package common

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// LockMemory locks the memory region occupied by the byte slice, preventing
// the OS from swapping it to disk. This is critical for sensitive data like
// passwords and encryption keys.
func LockMemory(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	if err := unix.Mlock(data); err != nil {
		return fmt.Errorf("mlock failed: %w", err)
	}

	return nil
}

// UnlockMemory unlocks the memory region, allowing the OS to swap it again.
func UnlockMemory(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	if err := unix.Munlock(data); err != nil {
		return fmt.Errorf("munlock failed: %w", err)
	}

	return nil
}
