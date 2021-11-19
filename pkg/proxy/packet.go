package proxy

import (
	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Packet interface provides access to the control packet or error
type Packet interface {
	GetControlPacket() packets.ControlPacket
	GetErrStatus() error
	getRoute() Route
}

type Route int

const (
	// Forward the packet to the other endpoint
	Forward Route = iota
	// Loopback the packet to the same endpoint
	Loopback
	// Process the packet from the endpoint
	Process
)

// NewPacket creates a packet with a route
func NewPacket(cp packets.ControlPacket, route Route) Packet {
	return &packetImpl{
		cp:    cp,
		route: route,
	}
}

// packetImpl wraps an MQTT control packet or error
type packetImpl struct {
	cp    packets.ControlPacket
	err   error
	route Route
}

// GetControlPacket retrieves the control packet
func (p *packetImpl) GetControlPacket() packets.ControlPacket {
	return p.cp
}

// GetErrStatus retrieves the error from the packet
func (p *packetImpl) GetErrStatus() error {
	return p.err
}

// getRoute retrieves the packet route
func (p *packetImpl) getRoute() Route {
	return p.route
}

// Packets is an slice of Packet
type Packets struct {
	packets []Packet
}

func (pkts *Packets) Add(p Packet) {
	if pkts != nil {
		pkts.packets = append(pkts.packets, p)
	}
}
