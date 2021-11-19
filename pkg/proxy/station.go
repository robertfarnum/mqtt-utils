package proxy

import (
	"context"
	"net"
	"strings"
	"sync"
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

// StationErrors is the set of errors return after starting the Station
type StationErrors struct {
	errs []error
}

// Add and error to the StationErrors
func (ses *StationErrors) Add(err error) {
	ses.errs = append(ses.errs, err)
}

// Error return the string format for the StationErrors
func (ses *StationErrors) Error() string {
	acc := make([]string, 0, len(ses.errs))
	for _, ce := range ses.errs {
		if ce != nil {
			acc = append(acc, ce.Error())
		}
	}
	if 0 < len(acc) {
		return strings.Join(acc, ",")
	}

	return "N/A"
}

// Start the pump; stops only after the hoses are disconnected or cancelled
func (s *Station) Run(ctx context.Context, cancel context.CancelFunc) error {
	ses := &StationErrors{}

	var wg = &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		err := s.client.Run(ctx)
		ses.Add(err)
		cancel()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		err := s.broker.Run(ctx)
		ses.Add(err)
		cancel()
		wg.Done()
	}()

	wg.Wait()

	return ses
}
