package proxy

import (
	"context"
	"io"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Gas wraps an MQTT control packet or error
type Gas struct {
	cp  packets.ControlPacket
	err error
}

// Nozzle is the channel type for the gas output channel
type Nozzle chan Gas

// Hose represents the MQTT network connection to channel Hose
type Hose struct {
	in  io.Reader
	out chan Gas
}

// NewHose creates a Hose from an input network connection to an output channel
func NewHose(in io.Reader) Hose {
	return Hose{
		in:  in,
		out: make(Nozzle, 1024),
	}
}

// flow reads an MQTT ControlPacket from the inbound io.Reader and forwards the result to the outbound Nozzle channel.
// flow does not return until an error occurs or its cancelled
func flow(ctx context.Context, Hose Hose) error {
	for {
		cp, err := packets.ReadPacket(Hose.in)
		if err != nil {
			return err
		}

		gas := Gas{
			cp:  cp,
			err: err,
		}

		select {
		case <-ctx.Done():
			return nil
		case Hose.out <- gas:
		}
	}
}
