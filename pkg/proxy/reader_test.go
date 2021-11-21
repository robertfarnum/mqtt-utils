package proxy

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"
)

// ErrTerminate Test indicates the test was terminated normally
var ErrTerminateTest = errors.New("terminate test")

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

// ErrorReader test structure
type ErrorReader struct {
	Err error
}

// Read implements the io.Reader interface for an Error Reader
func (r *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

// ControlPacketReader for testing with a io.Reader interface and a control packet
type ControlPacketReader struct {
	ControlPacketList Packets
	index             int
	buf               *bytes.Buffer
}

// Read a control packet into the []byte
func (r *ControlPacketReader) Read(p []byte) (n int, err error) {
	if r.buf == nil {
		if r.ControlPacketList.Len() <= r.index {
			return 0, ErrTerminateTest
		}
		r.buf = &bytes.Buffer{}
		cp := r.ControlPacketList.packets[r.index].GetControlPacket()
		r.index = r.index + 1

		if err = cp.Write(r.buf); err != nil {
			return 0, err
		}
	}

	n, err = r.buf.Read(p)

	if r.buf.Len() == 0 {
		r.buf = nil
	}

	return n, err
}

// ControlPacketWriter for testing with an io.Writer interface and a control packet
type ControlPacketWriter struct {
	err error
}

// Write a control packet into the []byte
func (r *ControlPacketWriter) Write(p []byte) (n int, err error) {
	return len(p), r.err
}

// DeadlineReader for testing io.Reader timeout
type DeadlineReader struct {
	reader  io.Reader
	timeout time.Duration
}

// NewDeadlineReader create a new DeadlineReader
func NewDeadlineReader(reader io.Reader, timeout time.Duration) io.Reader {
	ret := new(DeadlineReader)
	ret.reader = reader
	ret.timeout = timeout
	return ret
}

// Read from the DeadlineReader
func (r *DeadlineReader) Read(buf []byte) (n int, err error) {
	ch := make(chan bool)
	n = 0
	err = nil
	go func() {
		n, err = r.reader.Read(buf)
		ch <- true
	}()
	select {
	case <-ch:
	case <-time.After(r.timeout):
		n = 0
		err = os.ErrDeadlineExceeded
	}
	return n, err
}
