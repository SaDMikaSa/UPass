//go:build windows

package common

// LockMemory is a no-op on Windows. Windows does not have a direct equivalent
// to mlock. For maximum security on Windows, use Full-Disk Encryption (BitLocker).
func LockMemory(data []byte) error {
	return nil
}

// UnlockMemory is a no-op on Windows.
func UnlockMemory(data []byte) error {
	return nil
}
