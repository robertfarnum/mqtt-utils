package proxy

import (
	"io"

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
	return NewPacketWithError(cp, route, nil)
}

// NewPacketWithError creates a packet with a route and error
func NewPacketWithError(cp packets.ControlPacket, route Route, err error) Packet {
	return &packetImpl{
		cp:    cp,
		route: route,
		err:   err,
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

func (pkts *Packets) Add(p ...Packet) {
	if pkts != nil {
		pkts.packets = append(pkts.packets, p...)
	}
}

func (pkts *Packets) AddPackets(ps *Packets) {
	if pkts == nil || ps == nil || ps.packets == nil {
		return
	}

	pkts.packets = append(pkts.packets, ps.packets...)
}

func (pkts *Packets) Len() int {
	return len(pkts.packets)
}

type ProcessControlPacketType string

const (
	ReadyPacketType ProcessControlPacketType = "ready"
)

func NewProcessControlPacket(pt ProcessControlPacketType) packets.ControlPacket {
	return &ProcessControlPacket{
		packetType: pt,
	}
}

type ProcessControlPacket struct {
	packets.ControlPacket
	packetType ProcessControlPacketType
}

func (p *ProcessControlPacket) Write(w io.Writer) error {
	return nil
}

func (p *ProcessControlPacket) Unpack(io.Reader) error {
	return nil
}
func (p *ProcessControlPacket) String() string {
	return string(p.packetType)
}

func (p *ProcessControlPacket) Details() packets.Details {
	return packets.Details{}
}

func (p *ProcessControlPacket) GetPacketType() ProcessControlPacketType {
	return p.packetType
}
