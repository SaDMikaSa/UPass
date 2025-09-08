package zero

func SecureZero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
