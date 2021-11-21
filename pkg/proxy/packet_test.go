package proxy

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

func NewConnectPacket() Packet {
	cp := packets.NewControlPacket(packets.Connect).(*packets.ConnectPacket)
	cp.Username = "test"
	cp.UsernameFlag = true
	return NewPacket(cp, Forward)
}

type NozzlePacketProcessor struct {
	done   bool
	pkts   Packets
	nozzle Nozzle
	len    int
}

func NewNozzlePacketProcessor(nozzle Nozzle, len int) *NozzlePacketProcessor {
	return &NozzlePacketProcessor{
		nozzle: nozzle,
		len:    len,
	}
}

func (pp *NozzlePacketProcessor) Process(ctx context.Context, pkt Packet) (err error) {
	switch cp := pkt.GetControlPacket().(type) {
	case *ProcessControlPacket:
		fmt.Printf("Got Process Control Packet for %s\n", cp.GetPacketType())
	case *packets.ConnectPacket:
		fmt.Printf("Got Connect Packet\n")
	default:
		err = fmt.Errorf("Unknown control packet type")
	}

	return err
}

func (pp *NozzlePacketProcessor) Run(ctx context.Context) (pkts Packets, err error) {
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case p, more := <-pp.nozzle:
			if !more {
				err = ErrTerminateTest
				break LOOP
			}

			err := pp.Process(ctx, p)
			if err != nil {
				break LOOP
			}

			pkts.Add(p)
		}
	}

	return pkts, err
}

func TestNewPacket(t *testing.T) {
	type args struct {
		cp    packets.ControlPacket
		route Route
	}
	tests := []struct {
		name string
		args args
		want Packet
	}{
		{
			name: "Happy Path",
			args: args{
				cp:    packets.NewControlPacket(packets.Connect),
				route: Process,
			},
			want: NewPacket(packets.NewControlPacket(packets.Connect), Process),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPacket(tt.args.cp, tt.args.route); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPacket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_packetImpl_GetControlPacket(t *testing.T) {
	tests := []struct {
		name string
		p    *packetImpl
		want packets.ControlPacket
	}{
		{
			name: "Happy Path",
			p: &packetImpl{
				cp: packets.NewControlPacket(packets.Connect),
			},
			want: packets.NewControlPacket(packets.Connect),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.GetControlPacket(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("packetImpl.GetControlPacket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_packetImpl_GetErrStatus(t *testing.T) {
	tests := []struct {
		name    string
		p       *packetImpl
		wantErr error
	}{
		{
			name: "Happy Path",
			p: &packetImpl{
				err: errors.New("error"),
			},
			wantErr: errors.New("error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.GetErrStatus(); err.Error() != tt.wantErr.Error() {
				t.Errorf("packetImpl.GetErrStatus() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func Test_packetImpl_getRoute(t *testing.T) {
	tests := []struct {
		name string
		p    *packetImpl
		want Route
	}{
		{
			name: "Happy Path",
			p: &packetImpl{
				route: Forward,
			},
			want: Forward,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.getRoute(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("packetImpl.getRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPackets_Add(t *testing.T) {
	type args struct {
		p Packet
	}
	tests := []struct {
		name string
		pkts *Packets
		args args
		len  int
	}{
		{
			name: "Happy Path",
			pkts: &Packets{
				packets: []Packet{
					&packetImpl{
						err:   nil,
						route: Forward,
					},
				},
			},
			args: args{
				p: NewPacket(nil, Loopback),
			},
			len: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.pkts.Add(tt.args.p)
			if len(tt.pkts.packets) != tt.len {
				t.Errorf("len(tt.pkts.packets) = %d != tt.len = %d", len(tt.pkts.packets), tt.len)
			}
			if tt.pkts.packets[tt.len-1].getRoute() != tt.args.p.getRoute() {
				t.Errorf("routes do not match")
			}
		})
	}
}

func TestNewPacketWithError(t *testing.T) {
	type args struct {
		cp    packets.ControlPacket
		route Route
		err   error
	}
	tests := []struct {
		name string
		args args
		want Packet
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPacketWithError(tt.args.cp, tt.args.route, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPacketWithError() = %v, want %v", got, tt.want)
			}
		})
	}
}
