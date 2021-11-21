package proxy

import (
	"errors"
	"io"
	"net"
	"os"
	"time"
)

type NetConnError error

var (
	ErrNetConnReaderNotSet = errors.New("reader not set")
	ErrNetConnWriterNotSet = errors.New("writer not set")
)

// NetConn test structure implement net.Conn
type NetConn struct {
	net.Conn
	Reader              io.Reader
	Writer              io.Writer
	CloseErr            error
	LocalAddrRet        net.Addr
	RemoteAddrRet       net.Addr
	SetDeadlineErr      error
	SetReadDeadlineErr  error
	readDeadline        *time.Time
	SetWriteDeadlineErr error
	writeDeadline       *time.Time
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (nc *NetConn) Read(b []byte) (n int, err error) {
	if nc.Reader == nil {
		return 0, ErrNetConnReaderNotSet
	}

	if nc.readDeadline != nil {
		d := nc.readDeadline.Sub(time.Now())
		ch := make(chan bool)
		n = 0
		err = nil
		go func() {
			n, err = nc.Reader.Read(b)
			ch <- true
		}()
		select {
		case <-ch:
			return
		case <-time.After(d):
			return 0, os.ErrDeadlineExceeded
		}
		return
	}

	return nc.Reader.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (nc *NetConn) Write(b []byte) (n int, err error) {
	if nc.Writer == nil {
		return 0, ErrNetConnWriterNotSet
	}

	if nc.writeDeadline != nil {
		d := nc.writeDeadline.Sub(time.Now())
		ch := make(chan bool)
		n = 0
		err = nil
		go func() {
			n, err = nc.Writer.Write(b)
			ch <- true
		}()
		select {
		case <-ch:
			return
		case <-time.After(d):
			return 0, os.ErrDeadlineExceeded
		}
		return
	}

	return nc.Writer.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (nc *NetConn) Close() error {
	return nc.CloseErr
}

// LocalAddr returns the local network address.
func (nc *NetConn) LocalAddr() net.Addr {
	return nc.LocalAddrRet
}

// RemoteAddr returns the remote network address.
func (nc *NetConn) RemoteAddr() net.Addr {
	return nc.RemoteAddrRet
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (nc *NetConn) SetDeadline(t time.Time) (err error) {
	err = nc.SetReadDeadline(t)
	if err != nil {
		return err
	}
	err = nc.SetWriteDeadline(t)
	if err != nil {
		return err
	}

	return nc.SetDeadlineErr
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (nc *NetConn) SetReadDeadline(t time.Time) error {
	nc.readDeadline = &t

	return nc.SetReadDeadlineErr
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (nc *NetConn) SetWriteDeadline(t time.Time) error {
	nc.writeDeadline = &t

	return nc.SetWriteDeadlineErr
}
