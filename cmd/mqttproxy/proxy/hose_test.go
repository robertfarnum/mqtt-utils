package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

type SleepReader struct {
	io.Reader
}

func (r SleepReader) Read(p []byte) (n int, err error) {
	time.Sleep(3 * time.Second)

	return 0, fmt.Errorf("timeout")
}

type ConnectPacketReader struct {
	io.Reader
}

func (r ConnectPacketReader) Read(p []byte) (n int, err error) {
	packet := packets.ConnectPacket{
		Username: "test",
	}

	b := bytes.NewBuffer(p)

	err = packet.Write(b)
	if err != nil {
		return 0, err
	}

	return b.Len(), nil
}

func Test_Hose(t *testing.T) {
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 2*time.Second)
	cancelCtx, cancelCancel := context.WithCancel(context.TODO())

	type args struct {
		ctx    context.Context
		cancel context.CancelFunc
		hose   Hose
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "Flow Timeout",
			args: args{
				ctx:    timeoutCtx,
				cancel: timeoutCancel,
				hose:   NewHose(SleepReader{}),
			},
			wantErr: fmt.Errorf("timeout"),
		},
		{
			name: "Connect Packet Test",
			args: args{
				ctx:    cancelCtx,
				cancel: cancelCancel,
				hose:   NewHose(ConnectPacketReader{}),
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := flow(tt.args.ctx, tt.args.hose)
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("flow() error = %v, wantErr %v", err, tt.wantErr)
			} else if err != nil && tt.wantErr == nil {
				t.Errorf("flow() error = %v not expected", err)
			} else if err == nil && tt.wantErr != nil {
				t.Errorf("flow() error = %v not returned", tt.wantErr)
			}
		})
	}
}
