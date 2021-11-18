package proxy

import (
	"context"
	"net"
)

// Station
type Station struct {
	client *Pump
	broker *Pump
	cancel context.CancelFunc
}

// NewStation create a new bi-directional communication channel between the client and broker
func NewStation(client net.Conn, broker net.Conn) Station {
	return Station{
		client: NewPump(client),
		broker: NewPump(broker),
	}
}

// GetClientChan returns the client output channel
func (s Station) GetClientPump() *Pump {
	return s.client
}

// GetBrokerChan returns the broker output channel
func (s Station) GetBrokerPump() *Pump {
	return s.broker
}

// Start the pump; stops only after the pipes are drained
func (s *Station) Run(ctx context.Context, cancel context.CancelFunc) error {
	go func() {
		s.client.Run(ctx)
		cancel()
	}()

	go func() {
		s.broker.Run(ctx)
		cancel()
	}()

	return nil
}
