package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

type Status error

const (
	// DefaultReadTimeout is the default read timeout for DefaultHostConfig
	DefaultReadTimeout = 10 * time.Second
	// DefaultWriteTimeout is the default read timeout for DefaultHostConfig
	DefaultWriteTimeout = 10 * time.Second
)

var (
	ErrClosed   Status = fmt.Errorf("closed")
	ErrCanceled Status = fmt.Errorf("cancled")
	ErrTimeout  Status = fmt.Errorf("timeout")
)

// Packet wraps an MQTT control packet or error
type Packet struct {
	cp  packets.ControlPacket
	err error
}

// GetControlPacket retrieves the control packet
func (p *Packet) GetControlPacket() packets.ControlPacket {
	return p.cp
}

// GetError retrieves the packet error
func (p *Packet) GetError() error {
	return p.err
}

// Nozzle is the channel type for the Packet output channel
type Nozzle chan Packet

// Hose represents the MQTT network connection to packet channel (Nozzle)
type Hose struct {
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

	return Hose{
		in:     in,
		out:    make(Nozzle),
		config: *config,
	}
}

func (hose *Hose) GetNozzle() Nozzle {
	return hose.out
}

// flow reads an MQTT ControlPacket from the inbound io.Reader and forwards the result to the outbound Nozzle channel.
// flow does not return until an error occurs or its cancelled
func (h *Hose) Flow(ctx context.Context) error {
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

		p := Packet{
			cp:  cp,
			err: err,
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
