package proxy

import (
	"context"
	"io"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Packet wraps an MQTT control packet or error
type Packet struct {
	cp  packets.ControlPacket
	err error
}

// Nozzle is the channel type for the Packet output channel
type Nozzle chan Packet

// Hose represents the MQTT network connection to packet channel conduit
type Hose struct {
	in  io.Reader
	out Nozzle
}

// NewHose creates a Hose from an input network connection and build a packet output channel (Nozzle)
func NewHose(in io.Reader) Hose {
	return Hose{
		in:  in,
		out: make(Nozzle, 1024),
	}
}

func (hose Hose) GetNozzle() Nozzle {
	return hose.out
}

// flow reads an MQTT ControlPacket from the inbound io.Reader and forwards the result to the outbound Nozzle channel.
// flow does not return until an error occurs or its cancelled
func flow(ctx context.Context, Hose Hose) error {
	for {
		cp, err := packets.ReadPacket(Hose.in)
		if err != nil {
			return err
		}

		packet := Packet{
			cp:  cp,
			err: err,
		}

		select {
		case <-ctx.Done():
			return nil
		case Hose.out <- packet:
		}
	}
}
