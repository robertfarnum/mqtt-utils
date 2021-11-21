package proxy

import (
	"context"
	"errors"
	"sync"
)

// Channel interface for Run(ing) the Channel and Start(ing) the Pipe(s)
type Channel interface {
	GetClientNozzle() Nozzle
	GetBrokerNozzle() Nozzle
	Run(ctx context.Context, cancel context.CancelFunc) error
}

var (
	// ErrChannelIsNotSet indicates the channel is not set
	ErrChannelIsNotSet = errors.New("channel is not set")
	// ErrClientEndpointIsNotSet indicate the channels client endpoint is not set
	ErrClientEndpointIsNotSet = errors.New("client endpint is not set")
	// ErrBrokerEndpointIsNotSet indicate the channels client endpoint is not set
	ErrBrokerEndpointIsNotSet = errors.New("broker endpint is not set")
)

// ChannelConfig
type ChannelConfig struct {
	ClientEndpoint Endpoint
	BrokerEndpoint Endpoint
}

// channelImpl implements the Channel interface
type channelImpl struct {
	config ChannelConfig
}

// NewChannel creates a new bi-directional communication channel between the client and broker
func NewChannel(channelConfig ChannelConfig) Channel {
	return &channelImpl{
		config: channelConfig,
	}
}

// GetClientNozzle return the client nozzle (channel) for consuming packets
func (c *channelImpl) GetClientNozzle() Nozzle {
	if c.config.ClientEndpoint == nil {
		return nil
	}

	return c.config.ClientEndpoint.GetPipe().Nozzle()
}

// GetClientNozzle return the broker nozzle (channel) for consuming packets
func (c *channelImpl) GetBrokerNozzle() Nozzle {
	if c.config.BrokerEndpoint == nil {
		return nil
	}

	return c.config.BrokerEndpoint.GetPipe().Nozzle()
}

// Run the pump; stops only after the hoses are disconnected or cancelled
func (c *channelImpl) Run(ctx context.Context, cancel context.CancelFunc) error {
	if c == nil {
		return ErrChannelIsNotSet
	}

	ses := &Errors{}
	wg := &sync.WaitGroup{}

	if c.config.ClientEndpoint == nil {
		return ErrClientEndpointIsNotSet
	}

	// Run the client pipe
	wg.Add(1)
	go func() {
		err := c.config.ClientEndpoint.GetPipe().Run(ctx)
		ses.Add(err)
		cancel()
		wg.Done()
	}()

	if c.config.BrokerEndpoint == nil {
		return ErrBrokerEndpointIsNotSet
	}

	// Run the broker pipe
	wg.Add(1)
	go func() {
		err := c.config.BrokerEndpoint.GetPipe().Run(ctx)
		ses.Add(err)
		cancel()
		wg.Done()
	}()

	wg.Wait()

	return ses
}
