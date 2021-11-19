package proxy_test

import (
	"time"
)

// SleepReader test structure
type SleepReader struct {
	Timeout time.Duration
	Err     error
	Len     int
}

// Read implements the io.Reader interface for a SleepReader
func (r *SleepReader) Read(p []byte) (n int, err error) {
	time.Sleep(r.Timeout)

	return r.Len, r.Err
}
