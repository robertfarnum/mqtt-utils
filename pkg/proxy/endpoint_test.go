package proxy

import (
	"errors"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/robertfarnum/mqtt-utils/pkg/test"
)

var (
	ErrWriteError = errors.New("write error")
)

func TestNewEndpoint(t *testing.T) {
	type args struct {
		config EndpointConfig
	}
	tests := []struct {
		name string
		args args
		want Endpoint
	}{
		{
			name: "Config Timeout Provided",
			args: args{
				config: EndpointConfig{
					TimeoutConfig: &TimeoutConfig{
						ReadTimeout:  time.Second,
						WriteTimeout: time.Second,
					},
				},
			},
			want: NewEndpoint(
				EndpointConfig{
					TimeoutConfig: &TimeoutConfig{
						ReadTimeout:  time.Second,
						WriteTimeout: time.Second,
					},
				},
			),
		},
		{
			name: "Config Timeout Not Provided",
			args: args{
				config: EndpointConfig{},
			},
			want: NewEndpoint(
				EndpointConfig{
					TimeoutConfig: &TimeoutConfig{
						ReadTimeout:  DefaultReadTimeout,
						WriteTimeout: DefaultWriteTimeout,
					},
				},
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewEndpoint(tt.args.config)
			err := test.CheckTestResults(got, tt.want, nil, nil)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_endpointImpl_Send(t *testing.T) {
	type args struct {
		cp packets.ControlPacket
	}
	tests := []struct {
		name    string
		e       *endpointImpl
		args    args
		wantErr error
	}{
		{
			name: "Send Success",
			e: NewEndpoint(EndpointConfig{
				Conn: &NetConn{
					Writer: &ControlPacketWriter{},
				},
			}).(*endpointImpl),
			args: args{
				cp: packets.NewControlPacket(packets.Connect),
			},
			wantErr: nil,
		},
		{
			name: "Send Error",
			e: NewEndpoint(EndpointConfig{
				Conn: &NetConn{
					Writer: &ControlPacketWriter{
						err: ErrWriteError,
					},
				},
			}).(*endpointImpl),
			args: args{
				cp: packets.NewControlPacket(packets.Connect),
			},
			wantErr: endpointError(ErrWriteError),
		},
		{
			name: "Error Control Packet Is Not Set",
			e: NewEndpoint(EndpointConfig{
				Conn: &NetConn{
					Writer: &ControlPacketWriter{},
				},
			}).(*endpointImpl),
			wantErr: ErrEndpointControlPacketIsNotSet,
		},
		{
			name:    "Error Endpoint Is Not Set",
			e:       nil,
			wantErr: ErrEndpointIsNotSet,
		},
		{
			name:    "Error Endpoint Timeout Config Is Not Set",
			e:       &endpointImpl{},
			wantErr: ErrEndpointTimeoutConfigIsNotSet,
		},
		{
			name:    "Error Conn Is Not Set",
			e:       NewEndpoint(EndpointConfig{}).(*endpointImpl),
			wantErr: ErrEndpointConnIsNotSet,
		},
		{
			name: "Error Write Deadline Error Is Not Set",
			e: NewEndpoint(EndpointConfig{
				Conn: &NetConn{
					Writer:              &ControlPacketWriter{},
					SetWriteDeadlineErr: os.ErrDeadlineExceeded,
				},
			}).(*endpointImpl),
			wantErr: os.ErrDeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.e.Send(tt.args.cp)
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Error(err)
			}
		})
	}
}

func Test_endpointImpl_GetConn(t *testing.T) {
	nc := &NetConn{}

	tests := []struct {
		name string
		e    *endpointImpl
		want net.Conn
	}{
		{
			name: "Success",
			e: NewEndpoint(EndpointConfig{
				Conn: nc,
			}).(*endpointImpl),
			want: nc,
		},
		{
			name: "Endpoint nil",
			e:    nil,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.GetConn(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("endpointImpl.GetConn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_endpointImpl_GetTimeoutConfig(t *testing.T) {
	tests := []struct {
		name string
		e    *endpointImpl
		want TimeoutConfig
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.GetTimeoutConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("endpointImpl.GetTimeoutConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
