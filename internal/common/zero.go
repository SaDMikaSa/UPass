package common

import "runtime"

// ZeroBytes overwrites the provided byte slice with zeroes, unlocks the memory
// (if it was locked), and uses runtime.KeepAlive to reduce the likelihood that
// the slice contents are garbage-collected or moved before being cleared.
//
//go:noinline
func ZeroBytes(data []byte) {
	for i := range data {
		data[i] = 0
	}

	_ = UnlockMemory(data)
	runtime.KeepAlive(data)
}
