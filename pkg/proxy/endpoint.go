package proxy

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// EndpointError of the Endpoint
type EndpointError error

var (
	// ErrEndpointIsNotSet indicates that the endpoint is not set
	ErrEndpointIsNotSet EndpointError = errors.New("endpoint is not set")
	// ErrEndpointConnIsNotSet indicates that the endpoint conn is not set
	ErrEndpointConnIsNotSet EndpointError = errors.New("endpoint conn is not set")
	// ErrEndpointTimeoutConfig indicates that the endpoint timeout config is not set
	ErrEndpointTimeoutConfigIsNotSet EndpointError = errors.New("endpoint timeout config is not set")
	// ErrEndpointControlPacketIsNotSet
	ErrEndpointControlPacketIsNotSet EndpointError = errors.New("endpoint control packet is not set")
)

// endpointError wraps and error as and Endpoint Error
func endpointError(err error) EndpointError {
	return fmt.Errorf("endpoint error %w", err)
}

type Endpoint interface {
	GetConn() net.Conn
	Send(cp packets.ControlPacket) error
	GetTimeoutConfig() *TimeoutConfig
	GetProcessor() Processor
	GetPipe() Pipe
}

// EndpointConfig used to create and Endpoint with:
//    - a network connection
//    - pipe configuration
//    - a processor
type EndpointConfig struct {
	// Conn is the network connection
	Conn net.Conn
	// TimeoutConfig is the pipe configuration
	TimeoutConfig *TimeoutConfig
	// Processor is the network package processor
	Processor Processor
}

// endpointImpl used to store endpoint (client or broker) connection and processor information
type endpointImpl struct {
	config EndpointConfig
	pipe   Pipe
}

// NewEndpoint creates a new Endpoint with the provided endpoint configuration
func NewEndpoint(config EndpointConfig) Endpoint {
	if config.TimeoutConfig == nil {
		config.TimeoutConfig = DefaultTimeoutConfig
	}

	pipeConfig := PipeConfig{
		Reader:        config.Conn,
		TimeoutConfig: config.TimeoutConfig,
	}

	return &endpointImpl{
		config: config,
		pipe:   NewPipe(pipeConfig),
	}
}

// getConn return the network connection of an endpoint
func (e *endpointImpl) GetConn() net.Conn {
	if e != nil {
		return e.config.Conn
	}

	return nil
}

// Send the control packet to the Endpoint network connection
func (e *endpointImpl) Send(cp packets.ControlPacket) (err error) {
	if e == nil {
		return ErrEndpointIsNotSet
	}

	if e.config.TimeoutConfig == nil {
		return ErrEndpointTimeoutConfigIsNotSet
	}

	if e.config.Conn == nil {
		return ErrEndpointConnIsNotSet
	}

	err = e.config.Conn.SetWriteDeadline(time.Now().Add(e.config.TimeoutConfig.WriteTimeout))
	if err != nil {
		return endpointError(err)
	}

	if cp == nil {
		return ErrEndpointControlPacketIsNotSet
	}

	err = cp.Write(e.config.Conn)
	if err != nil {
		return endpointError(err)
	}

	return nil
}

// GetTimeoutConfig returns the timeout configuration for and endpoint
func (e *endpointImpl) GetTimeoutConfig() *TimeoutConfig {
	return e.config.TimeoutConfig
}

// getProcessor returns the processor for and endpoint
func (e *endpointImpl) GetProcessor() Processor {
	return e.config.Processor
}

// getPipe returns the pipe for and endpoint
func (e *endpointImpl) GetPipe() Pipe {
	return e.pipe
}
