package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

const (
	// DefaultReadTimeout is the default read timeout for DefaultHostConfig
	DefaultReadTimeout = 10 * time.Second
	// DefaultWriteTimeout is the default read timeout for DefaultHostConfig
	DefaultWriteTimeout = 10 * time.Second
)

// HostStatus of Hose in Packet
type HostStatus error

var (
	ErrClosed   HostStatus = fmt.Errorf("closed")
	ErrCanceled HostStatus = fmt.Errorf("canceled")
	ErrTimeout  HostStatus = fmt.Errorf("timeout")
)

// Nozzle is the channel type for the Packet output channel
type Nozzle chan Packet

// Hose is the interface to start a flow and get the nozzle (channel) to be sent Packet(s)
type Hose interface {
	GetNozzle() Nozzle
	Flow(ctx context.Context) error
}

// hoseImpl represents the MQTT network connection to packet channel (Nozzle)
type hoseImpl struct {
	in     io.Reader
	out    Nozzle
	config HoseConfig
}

// HoseConfig configures the host tuneable parameters
type HoseConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

var DefaultHoseConfig = &HoseConfig{
	ReadTimeout:  DefaultReadTimeout,
	WriteTimeout: DefaultWriteTimeout,
}

// NewHose creates a Hose from an input network connection and build a packet output channel (Nozzle)
func NewHose(in io.Reader, config *HoseConfig) Hose {
	if config == nil {
		config = DefaultHoseConfig
	}

	return &hoseImpl{
		in:     in,
		out:    make(Nozzle),
		config: *config,
	}
}

func (h *hoseImpl) GetNozzle() Nozzle {
	return h.out
}

// flow reads an MQTT ControlPacket from the inbound io.Reader and forwards the result to the outbound Nozzle channel.
// flow does not return until an error occurs or its cancelled
func (h *hoseImpl) Flow(ctx context.Context) error {
	for {
		if in, ok := h.in.(net.Conn); ok {
			err := in.SetReadDeadline(time.Now().Add(h.config.ReadTimeout))
			if err != nil {
				return err
			}
		}

		cp, err := packets.ReadPacket(h.in)
		if err != nil {
			if err == io.EOF {
				return ErrClosed
			}

			return err
		}

		p := &packetImpl{
			cp:    cp,
			err:   err,
			route: Process,
		}

		select {
		case <-time.NewTimer(h.config.WriteTimeout).C:
			return ErrTimeout
		case <-ctx.Done():
			return ErrCanceled
		case h.out <- p:
		}
	}
}
