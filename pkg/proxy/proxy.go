package proxy

import (
	"context"
	"errors"
)

var (
	// ErrProxyForwardEndpointNotSet is a forward endpoint not set error
	ErrProxyForwardEndpointNotSet = errors.New("forward endpoint not set")
	// ErrProxyLoopbackEndpointNotSet is a loopback endpoint not set error
	ErrProxyLoopbackEndpointNotSet = errors.New("loopback endpoint not set")
	// ErrProcessorNotSet is a endpoint processor not set error
	ErrProcessorNotSet = errors.New("endpoint processor not set")
	// ErrProxyInvalidPacketRoute is a invalid packet route error
	ErrProxyInvalidPacketRoute = errors.New("invalid packet route")
	// ErrProxyClientDisconnect indicates the client has disconnected
	ErrProxyClientDisconnect = errors.New("client disconnect")
	// ErrProxyBrokerDisconnect indicates the broker has disconnected
	ErrProxyBrokerDisconnect = errors.New("broker disconnect")
	// ErrProxyPacketNotSet indicates the packet is not set
	ErrProxyPacketNotSet = errors.New("packet is not set")
)

// Proxy for client and broker
type Proxy interface {
	Run(ctx context.Context) error
}

// proxyImpl implements the Proxy interface
type proxyImpl struct {
	channelConfig ChannelConfig
}

// NewProxy creates a new proxy between the client and broker
func NewProxy(channelConfig ChannelConfig) Proxy {
	return &proxyImpl{
		channelConfig: channelConfig,
	}
}

// processPacket will process a incoming Packet and then forward or loopback output Packet(s)
func processPacket(ctx context.Context, p Packet, loopback Endpoint, forward Endpoint) (err error) {
	if p == nil {
		return ErrProxyPacketNotSet
	}

	if loopback == nil {
		return ErrProxyLoopbackEndpointNotSet
	}

	if forward == nil {
		return ErrProxyForwardEndpointNotSet
	}

	processor := loopback.GetProcessor()
	if processor == nil {
		return ErrProcessorNotSet
	}

	if p.getRoute() != Process {
		return ErrProxyInvalidPacketRoute
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
			return ErrProxyInvalidPacketRoute
		case Forward:
			cp.Write(forward.GetConn())
		case Loopback:
			cp.Write(loopback.GetConn())
		}
	}

	return nil
}

// Run the proxy
func (p *proxyImpl) Run(ctx context.Context) error {
	ch := NewChannel(p.channelConfig)
	ctx, cancel := context.WithCancel(ctx)
	errs := &Errors{}

	// The client packet processing loop
	go func() {
		n := ch.GetClientNozzle()

	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case pkt := <-n:
				if pkt == nil {
					errs.Add(ErrProxyClientDisconnect)
					break LOOP
				}
				err := processPacket(ctx, pkt, p.channelConfig.ClientEndpoint, p.channelConfig.BrokerEndpoint)
				if err != nil {
					errs.Add(err)
					break LOOP
				}
			}
		}
		cancel()
	}()

	// The broker packet processing loop
	go func() {
		n := ch.GetBrokerNozzle()

	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case pkt := <-n:
				if pkt == nil {
					errs.Add(ErrProxyBrokerDisconnect)
					break LOOP
				}
				err := processPacket(ctx, pkt, p.channelConfig.BrokerEndpoint, p.channelConfig.ClientEndpoint)
				if err != nil {
					errs.Add(err)
					break LOOP
				}
			}
		}
		cancel()
	}()

	err := ch.Run(ctx, cancel)
	errs.Add(err)

	return errs
}
