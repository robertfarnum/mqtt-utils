package proxy_test

import "io"

// writer test structure
type writer struct {
	p []byte
}

// NewWriter create a writer with a byte slice
func NewWriter(p []byte) io.Writer {
	return &writer{
		p: p,
	}
}

// Write implements the io.Writer interface for a writer
func (w *writer) Write(p []byte) (n int, err error) {
	w.p = p

	return len(p), nil
}
