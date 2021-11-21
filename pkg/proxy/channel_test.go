package proxy

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestNewChannel(t *testing.T) {
	clientConn := &NetConn{}
	brokerConn := &NetConn{}

	type args struct {
		config ChannelConfig
	}
	tests := []struct {
		name string
		args args
		want Channel
	}{
		{
			name: "Happy Path",
			args: args{
				config: ChannelConfig{
					ClientEndpoint: NewEndpoint(EndpointConfig{
						Conn: clientConn,
					}),
					BrokerEndpoint: NewEndpoint(EndpointConfig{
						Conn: brokerConn,
					}),
				},
			},
			want: NewChannel(ChannelConfig{
				ClientEndpoint: NewEndpoint(EndpointConfig{
					Conn: clientConn,
				}),
				BrokerEndpoint: NewEndpoint(EndpointConfig{
					Conn: brokerConn,
				}),
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewChannel(tt.args.config)
			if got.(*channelImpl).config.ClientEndpoint.GetConn() != tt.want.(*channelImpl).config.ClientEndpoint.GetConn() {
				t.Errorf("client net.Conn do not match")
			}
			if got.(*channelImpl).config.BrokerEndpoint.GetConn() != tt.want.(*channelImpl).config.BrokerEndpoint.GetConn() {
				t.Errorf("broker net.Conn do not match")
			}
		})
	}
}

func Test_channelImpl_GetClientNozzle(t *testing.T) {
	clientPipe := &pipeImpl{}

	tests := []struct {
		name string
		s    channelImpl
		want Pipe
	}{
		{
			name: "Happy Path",
			s: channelImpl{
				config: ChannelConfig{
					ClientEndpoint: NewEndpoint(EndpointConfig{}),
				},
			},
			want: clientPipe,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.GetClientNozzle(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("channelImpl.GetClientPipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_channelImpl_GetBrokerNozzle(t *testing.T) {
	brokerPipe := &pipeImpl{}

	tests := []struct {
		name string
		s    channelImpl
		want Pipe
	}{
		{
			name: "Happy Path",
			s: channelImpl{
				config: ChannelConfig{
					ClientEndpoint: NewEndpoint(EndpointConfig{}),
				},
			},
			want: brokerPipe,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.GetBrokerNozzle(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("channelImpl.GetBrokerPipe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_channelImpl_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 1*time.Second)
	clientEndpoint := NewEndpoint(EndpointConfig{
		Conn: &NetConn{
			Reader: &ControlPacketReader{},
			Writer: &ControlPacketWriter{},
		},
	})
	brokerEndpoint := NewEndpoint(EndpointConfig{
		Conn: &NetConn{
			Reader: &ControlPacketReader{},
			Writer: &ControlPacketWriter{},
		},
	})
	ch := NewChannel(ChannelConfig{
		ClientEndpoint: clientEndpoint,
		BrokerEndpoint: brokerEndpoint,
	})

	type args struct {
		ctx    context.Context
		cancel context.CancelFunc
	}
	tests := []struct {
		name    string
		ch      *channelImpl
		args    args
		wantErr error
	}{
		{
			name: "Timeout - Terminate Test",
			ch:   ch.(*channelImpl),
			args: args{
				ctx:    timeoutCtx,
				cancel: timeoutCancel,
			},
			wantErr: &Errors{
				errs: []error{
					ErrTerminateTest,
					ErrTerminateTest,
				},
			},
		},
		{
			name: "Client Endpoint Not Initialized",
			ch:   &channelImpl{},
			args: args{
				ctx:    ctx,
				cancel: cancel,
			},
			wantErr: ErrClientEndpointIsNotSet,
		},
		{
			name: "Broker Endpoint Not Initialized",
			ch: &channelImpl{
				config: ChannelConfig{
					ClientEndpoint: NewEndpoint(EndpointConfig{
						Conn: &NetConn{},
					}),
				},
			},
			args: args{
				ctx:    ctx,
				cancel: cancel,
			},
			wantErr: ErrBrokerEndpointIsNotSet,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ch.Run(tt.args.ctx, tt.args.cancel)
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("channelImpl.Run() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
