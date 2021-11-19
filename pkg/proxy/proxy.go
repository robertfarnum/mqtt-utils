package proxy

import (
	"context"
	"fmt"
	"net"
)

var (
	// ErrBadClientEndpoint is a bad client endpoint error
	ErrBadClientEndpoint = fmt.Errorf("bad client endpoint")
	// ErrBadBrokerEndpoint is a bad broker endpoint error
	ErrBadBrokerEndpoint = fmt.Errorf("bad broker endpoint")
	// ErrForwardEndpointNotSet is a forward endpoint not set error
	ErrForwardEndpointNotSet = fmt.Errorf("forward endpoint not set")
	// ErrLoopbackEndpointNotSet is a loopback endpoint not set error
	ErrLoopbackEndpointNotSet = fmt.Errorf("loopback endpoint not set")
	// ErrProcessorNotSet is a endpoint processor not set error
	ErrProcessorNotSet = fmt.Errorf("endpoint processor not set")
	// ErrInvalidPacketRoute is a invalid packet route error
	ErrInvalidPacketRoute = fmt.Errorf("invalid packet route")
)

// Processor interface used to process Packet(s)
type Processor interface {
	Process(ctx context.Context, p Packet) (pkts *Packets, err error)
}

// Endpoint interface for the client or broker
type Endpoint interface {
	getConn() net.Conn
	getProcessor() Processor
}

// endpointImpl used to store endpoint (client or broker) connection and processor information
type endpointImpl struct {
	conn      net.Conn
	processor Processor
}

// NewEndpoint creates a new Endpoint with a network connection and processor
func NewEndpoint(conn net.Conn, processor Processor) Endpoint {
	return &endpointImpl{
		conn:      conn,
		processor: processor,
	}
}

// getConn return the network connection of an endpoint
func (e *endpointImpl) getConn() net.Conn {
	return e.conn
}

// getProcessor returns the processor for and endpoint
func (e *endpointImpl) getProcessor() Processor {
	return e.processor
}

// Proxy for client and broker
type Proxy interface {
	Start(ctx context.Context) error
}

// proxyImpl implements the Proxy interface
type proxyImpl struct {
	client Endpoint
	broker Endpoint
}

// NewProxy creates a new proxy between the client and broker
func NewProxy(client Endpoint, broker Endpoint) Proxy {
	return &proxyImpl{
		client: client,
		broker: broker,
	}
}

// processPacket will process a incoming Packet and then forward or loopback output Packet(s)
func processPacket(ctx context.Context, p Packet, loopback Endpoint, forward Endpoint) (err error) {
	if loopback == nil {
		return ErrLoopbackEndpointNotSet
	}

	if forward == nil {
		return ErrForwardEndpointNotSet
	}

	processor := loopback.getProcessor()
	if processor == nil {
		return ErrProcessorNotSet
	}

	if p.getRoute() != Process {
		return ErrInvalidPacketRoute
	}

	pkts, err := processor.Process(ctx, p)
	if err != nil {
		return err
	}

	for _, pkt := range pkts.packets {
		route := pkt.getRoute()
		cp := pkt.GetControlPacket()
		switch route {
		case Process:
			return ErrInvalidPacketRoute
		case Forward:
			cp.Write(forward.getConn())
		case Loopback:
			cp.Write(loopback.getConn())
		}
	}

	return nil
}

// Start the proxy
func (p *proxyImpl) Start(ctx context.Context) error {

	if p.client == nil || p.client.getConn() == nil {
		return ErrBadClientEndpoint
	}

	if p.client == nil || p.client.getConn() == nil {
		return ErrBadBrokerEndpoint
	}

	s := NewStation(p.client.getConn(), p.broker.getConn())
	ctx, cancel := context.WithCancel(ctx)

	errs := &Errors{}

	go func() {
		n := s.GetClientPump().Nozzle()

	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case pkt := <-n:
				err := processPacket(ctx, pkt, p.client, p.broker)
				if err != nil {
					errs.Add(err)
					cancel()
					return
				}
			}
		}
	}()

	go func() {
		n := s.GetBrokerPump().Nozzle()

	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case pkt := <-n:
				err := processPacket(ctx, pkt, p.broker, p.client)
				if err != nil {
					errs.Add(err)
					cancel()
					return
				}
			}
		}
	}()

	err := s.Run(ctx, cancel)
	errs.Add(err)

	return errs
}
