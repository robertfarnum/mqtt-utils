package proxy_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/robertfarnum/mqtt-utils/pkg/proxy"
)

func printPacket(p proxy.Packet) {
	b, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("processed packet: %s\n", string(b))
}

func processNozzle(ctx context.Context, n proxy.Nozzle) (pkts []proxy.Packet) {
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case p := <-n:
			pkts = append(pkts, p)

			printPacket(p)
		}
	}

	return pkts
}

func TestHose(t *testing.T) {
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 2*time.Second)
	cancelCtx, cancelCancel := context.WithCancel(context.TODO())

	type args struct {
		ctx    context.Context
		cancel context.CancelFunc
		hose   proxy.Hose
	}
	tests := []struct {
		name     string
		args     args
		wantErr  error
		wantPkts []proxy.Packet
	}{
		{
			name: "Flow Timeout",
			args: args{
				ctx:    timeoutCtx,
				cancel: timeoutCancel,
				hose: proxy.NewHose(&SleepReader{
					Timeout: 3 * time.Second,
					Err:     fmt.Errorf("timeout"),
				}, nil),
			},
			wantErr:  fmt.Errorf("timeout"),
			wantPkts: nil,
		},
		{
			name: "Connect Packet Test",
			args: args{
				ctx:    cancelCtx,
				cancel: cancelCancel,
				hose: proxy.NewHose(&ControlPacketReader{
					ControlPacketList: []packets.ControlPacket{
						func() packets.ControlPacket {
							cp := packets.NewControlPacket(packets.Connect).(*packets.ConnectPacket)
							cp.Username = "test"
							cp.UsernameFlag = true
							return cp
						}(),
					},
				}, nil),
			},
			wantErr:  TerminateTest,
			wantPkts: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.args.cancel()

			var (
				pkts []proxy.Packet
				err  error
			)

			go func() {
				pkts = processNozzle(tt.args.ctx, tt.args.hose.GetNozzle())
			}()

			err = tt.args.hose.Flow(tt.args.ctx)

			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("flow() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err != nil && tt.wantErr == nil {
				t.Errorf("flow() error = %v not expected", err)
				return
			} else if err == nil && tt.wantErr != nil {
				t.Errorf("flow() error = %v not returned", tt.wantErr)
				return
			}

			if !reflect.DeepEqual(tt.wantPkts, pkts) {
				t.Errorf("process and flow do not match")
			}
		})
	}
}
