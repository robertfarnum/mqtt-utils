package proxy

import (
	"context"
	"net"
	"sync"
)

// Station interface for Run(ing) the Station and Start(ing) the Pump(s)
type Station interface {
	GetClientPump() Pump
	GetBrokerPump() Pump
	Run(ctx context.Context, cancel context.CancelFunc) error
}

// stationImpl implements the Station interface
type stationImpl struct {
	client Pump
	broker Pump
}

// NewStation create a new bi-directional communication channel between the client and broker
func NewStation(client net.Conn, broker net.Conn) Station {
	return &stationImpl{
		client: NewPump(client),
		broker: NewPump(broker),
	}
}

// GetClientChan returns the client output channel
func (s stationImpl) GetClientPump() Pump {
	return s.client
}

// GetBrokerChan returns the broker output channel
func (s stationImpl) GetBrokerPump() Pump {
	return s.broker
}

// Start the pump; stops only after the hoses are disconnected or cancelled
func (s *stationImpl) Run(ctx context.Context, cancel context.CancelFunc) error {
	ses := &Errors{}

	var wg = &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		err := s.client.Start(ctx)
		ses.Add(err)
		cancel()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err := s.broker.Start(ctx)
		ses.Add(err)
		cancel()
		wg.Done()
	}()

	wg.Wait()

	return ses
}
