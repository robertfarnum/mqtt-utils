package proxy

import (
	"context"
	"net"
)

// Pump
type Pump struct {
	// The bi-directional network connection
	conn net.Conn

	// The hose to pump the MQTT packet Gas from the network connection to the channel
	hose Hose
}

// NewPump create a new bi-directional communication channel between the client and broker
func NewPump(conn net.Conn) Pump {
	return Pump{
		conn: conn,
		hose: NewHose(conn),
	}
}

func (p Pump) pump(ctx context.Context) error {
	return nil
}
