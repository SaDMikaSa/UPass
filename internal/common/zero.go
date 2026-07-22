package common

import "runtime"

// ZeroBytes overwrites the provided byte slice with zeroes and uses
// runtime.KeepAlive to reduce the likelihood that the slice contents are
// garbage-collected or moved before being cleared. This mitigates, but
// does not fully eliminate, secret retention in Go memory.
//
//go:noinline
func ZeroBytes(data []byte) {
	for i := range data {
		data[i] = 0
	}
	runtime.KeepAlive(data)
}
