package zero

func SecureZero(b []byte) {
	if b == nil {
		return
	}
	for i := range b {
		b[i] = 0
	}
}
