package proxy

import (
	"context"
	"net"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Pump
type Pump struct {
	// The network connection
	conn net.Conn

	// The hose to pump the MQTT packet Gas from the network connection to the channel
	hose Hose
}

// NewPump create a new bi-directional communication channel between the client and broker
func NewPump(conn net.Conn) *Pump {
	return &Pump{
		conn: conn,
		hose: NewHose(conn, nil),
	}
}

func (p *Pump) Nozzle() Nozzle {
	return p.hose.GetNozzle()
}

// Write to the
func (p *Pump) Write(cp packets.ControlPacket) {
	cp.Write(p.conn)
}

func (p *Pump) Run(ctx context.Context) error {
	return p.hose.Flow(ctx)
}
