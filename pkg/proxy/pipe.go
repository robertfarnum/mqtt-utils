package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Pipe interface to Run the Pipe and fetch the Nozzle (channel)
type Pipe interface {
	Nozzle() Nozzle
	Run(ctx context.Context) error
}

const (
	// DefaultReadTimeout is the default read timeout for DefaultHostConfig
	DefaultReadTimeout = 10 * time.Second
	// DefaultWriteTimeout is the default read timeout for DefaultHostConfig
	DefaultWriteTimeout = 10 * time.Second
)

// PipeStatus of the Pipe
type PipeStatus error

var (
	// ErrClose indicates the Pipe was closed
	ErrPipeClosed PipeStatus = errors.New("closed")
	// ErrCanceled indicates the Pipe was canceled
	ErrPipeCanceled PipeStatus = errors.New("canceled")
	// ErrPipeReadTimeout indicates the Pipe has timed out
	ErrPipeReadTimeout PipeStatus = errors.New("read timeout")
	// ErrPipeWriteTimeout indicates the Pipe has timed out
	ErrPipeWriteTimeout PipeStatus = errors.New("write timeout")
)

// pipeStatus wraps an error as a PipeStatus
func pipeStatus(err error) PipeStatus {
	return fmt.Errorf("pipe error %w", err)
}

// Errors
var (
	// ErrPipeIsNotSet indicates the Pipe is not set
	ErrPipeIsNotSet = errors.New("pipe is not set")
	// ErrPipeReaderIsNotSet indicates the Pipe io.Reader is not set
	ErrPipeReaderIsNotSet = errors.New("pipe io.Reader is not set")
	// ErrPipeCtxIsNotSet indicates the Pipe ctx is not set
	ErrPipeCtxIsNotSet = errors.New("pipe ctx is not set")
	// ErrPipeChannelIsNotSet indicates the Pipe channel is not set")
	ErrPipeChannelIsNotSet = errors.New("pipe channel is not set")
)

// DefaultPipeConfig is the default PipeConfig settings if not provided
var DefaultTimeoutConfig = &TimeoutConfig{
	ReadTimeout:  DefaultReadTimeout,
	WriteTimeout: DefaultWriteTimeout,
}

// TimeoutConfig for newtwork read and channel write operations
type TimeoutConfig struct {
	// ReadTimeout for network connection read operations
	ReadTimeout time.Duration
	// WriteTimeout for network connection write operations
	WriteTimeout time.Duration
}

// PipeConfig configures the host tuneable parameters
type PipeConfig struct {
	// Conn is the network connection
	Reader io.Reader
	// TimeoutConfig for Pipe network operations
	TimeoutConfig *TimeoutConfig
}

// Nozzle is the channel type for the Packet output channel
type Nozzle chan Packet

// pipeImpl implements the Pipe interface
type pipeImpl struct {
	// config for the Pipe
	config PipeConfig
	// out is the channel output Nozzle
	out Nozzle
}

// NewPipe create a new bi-directional communication channel between the client and broker
func NewPipe(config PipeConfig) Pipe {
	if config.TimeoutConfig == nil {
		config.TimeoutConfig = DefaultTimeoutConfig
	}

	return &pipeImpl{
		config: config,
		out:    make(Nozzle),
	}
}

func (p *pipeImpl) Nozzle() Nozzle {
	if p == nil {
		return nil
	}

	return p.out
}

// Run the Pipe
func (p *pipeImpl) Run(ctx context.Context) (err error) {
	if p == nil {
		return pipeStatus(ErrPipeIsNotSet)
	}

	if ctx == nil {
		return pipeStatus(ErrPipeCtxIsNotSet)
	}

	if p.config.Reader == nil {
		return pipeStatus(ErrPipeReaderIsNotSet)
	}

	if p.out == nil {
		return pipeStatus(ErrPipeChannelIsNotSet)
	}

	cp := NewProcessControlPacket(ReadyPacketType)

LOOP:
	for {
		// Create the packet
		pkt := NewPacketWithError(cp, Process, err)

		select {
		case <-time.NewTimer(p.config.TimeoutConfig.WriteTimeout).C:
			err = pipeStatus(ErrPipeWriteTimeout)
			break LOOP
		case <-ctx.Done():
			err = pipeStatus(ErrPipeCanceled)
			break LOOP
		case p.out <- pkt:
		}

		// If the Reader is a new.Conn then set the read deadline
		if conn, ok := p.config.Reader.(net.Conn); ok {
			err = conn.SetReadDeadline(time.Now().Add(p.config.TimeoutConfig.ReadTimeout))
			if err != nil {
				err = pipeStatus(err)
				break LOOP
			}
		}

		// Read the incoming packet
		cp, err = packets.ReadPacket(p.config.Reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = pipeStatus(ErrPipeClosed)
			} else if errors.Is(err, os.ErrDeadlineExceeded) {
				err = pipeStatus(ErrPipeReadTimeout)
			} else {
				err = pipeStatus(err)
			}
			break LOOP
		}
	}

	close(p.out)

	return err
}
