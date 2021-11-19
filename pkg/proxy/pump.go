package proxy

import (
	"context"
	"net"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Pump interface to Start the Pump and fetch the Nozzle (channel)
type Pump interface {
	Nozzle() Nozzle
	Send(cp packets.ControlPacket) error
	Start(ctx context.Context) error
}

// pumpImpl implements the Pump interface
type pumpImpl struct {
	// The network connection
	conn net.Conn

	// The hose to pump the MQTT packet Gas from the network connection to the channel
	hose Hose
}

// NewPump create a new bi-directional communication channel between the client and broker
func NewPump(conn net.Conn) Pump {
	return &pumpImpl{
		conn: conn,
		hose: NewHose(conn, nil),
	}
}

func (p *pumpImpl) Nozzle() Nozzle {
	return p.hose.GetNozzle()
}

// Send the control packet to the Pump
func (p *pumpImpl) Send(cp packets.ControlPacket) error {
	return cp.Write(p.conn)
}

// Start the Pump
func (p *pumpImpl) Start(ctx context.Context) error {
	return p.hose.Flow(ctx)
}
