package proxy

import (
	"context"
	"errors"
	"io"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/robertfarnum/mqtt-utils/pkg/test"
)

var (
	ErrSetReadDeadline = pipeStatus(errors.New("SetReadDeadline"))
)

type CancelContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewCancelContext() CancelContext {
	ctx, cancel := context.WithCancel(context.TODO())
	return CancelContext{
		ctx:    ctx,
		cancel: cancel,
	}
}

func NewTimeoutContext(timeout time.Duration) CancelContext {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return CancelContext{
		ctx:    ctx,
		cancel: cancel,
	}
}

func TestPipeRun(t *testing.T) {
	var pipeImplPtr *pipeImpl
	var pipe Pipe = pipeImplPtr

	type args struct {
		cc   CancelContext
		pipe Pipe
	}
	tests := []struct {
		name         string
		args         args
		isTimeout    bool
		clearChannel bool
		wantErr      error
		wantPkts     Packets
	}{
		{
			name: "Pipe Context Timeout",
			args: args{
				cc: NewTimeoutContext(1 * time.Second),
				pipe: NewPipe(PipeConfig{
					Reader: &SleepReader{
						Timeout: 2 * time.Second,
						Err:     ErrPipeReadTimeout,
					},
				}),
			},
			isTimeout: true,
			wantErr:   ErrPipeReadTimeout,
		},
		{
			name: "Pipe Read Timeout",
			args: args{
				cc: NewCancelContext(),
				pipe: NewPipe(PipeConfig{
					Reader: NewDeadlineReader(&SleepReader{
						Timeout: 2 * time.Second,
						Err:     nil,
					}, 1*time.Second),
				}),
			},
			isTimeout: true,
			wantErr:   ErrPipeReadTimeout,
		},
		{
			name: "Pipe NetConn Read Timeout",
			args: args{
				cc: NewCancelContext(),
				pipe: NewPipe(PipeConfig{
					Reader: &NetConn{
						Reader: &SleepReader{
							Timeout: 2 * time.Second,
							Err:     ErrPipeReadTimeout,
						},
					},
					TimeoutConfig: &TimeoutConfig{
						ReadTimeout:  1 * time.Second,
						WriteTimeout: DefaultWriteTimeout,
					},
				}),
			},
			isTimeout: true,
			wantErr:   ErrPipeReadTimeout,
		},
		{
			name: "Pipe NetConn Read Deadline Error",
			args: args{
				cc: NewCancelContext(),
				pipe: NewPipe(PipeConfig{
					Reader: &NetConn{
						Reader: &SleepReader{
							Timeout: 2 * time.Second,
							Err:     ErrPipeReadTimeout,
						},
						SetReadDeadlineErr: ErrSetReadDeadline,
					},
				}),
			},
			wantErr: ErrSetReadDeadline,
		},
		{
			name: "Pipe Channel Write Timeout",
			args: args{
				cc: NewCancelContext(),
				pipe: NewPipe(PipeConfig{
					Reader: &NetConn{
						Reader: &SleepReader{
							Timeout: 2 * time.Second,
							Err:     ErrPipeWriteTimeout,
						},
					},
					TimeoutConfig: &TimeoutConfig{
						ReadTimeout: 1 * time.Second,
					},
				}),
			},
			isTimeout: true,
			wantErr:   ErrPipeWriteTimeout,
		},
		{
			name: "Pipe Closed",
			args: args{
				cc: NewCancelContext(),
				pipe: NewPipe(PipeConfig{
					Reader: &ErrorReader{
						Err: io.EOF,
					},
				}),
			},
			wantErr: ErrPipeClosed,
		},
		{
			name: "Connect Packet Test",
			args: args{
				cc: NewCancelContext(),
				pipe: NewPipe(PipeConfig{
					Reader: &ControlPacketReader{
						ControlPacketList: Packets{
							packets: []Packet{
								NewConnectPacket(),
							},
						},
					},
				}),
			},
			wantPkts: Packets{
				packets: []Packet{
					NewConnectPacket(),
				},
			},
			wantErr: ErrTerminateTest,
		},
		{
			name: "Pipe Is Not Set",
			args: args{
				pipe: pipe,
			},
			wantErr: ErrPipeIsNotSet,
		},
		{
			name: "Ctx Is Not Set",
			args: args{
				pipe: NewPipe(PipeConfig{}),
			},
			wantErr: ErrPipeCtxIsNotSet,
		},
		{
			name: "Reader Is Not Set",
			args: args{
				cc:   NewTimeoutContext(time.Second),
				pipe: NewPipe(PipeConfig{}),
			},
			wantErr: ErrPipeReaderIsNotSet,
		},
		{
			name: "Channel Is Not Set",
			args: args{
				cc: NewTimeoutContext(time.Second),
				pipe: NewPipe(PipeConfig{
					Reader: &ControlPacketReader{
						ControlPacketList: Packets{
							packets: []Packet{
								NewConnectPacket(),
							},
						},
					},
				}),
			},
			clearChannel: true,
			wantErr:      ErrPipeChannelIsNotSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.cc.cancel != nil {
				defer tt.args.cc.cancel()
			}

			var (
				pkts Packets
				err  error
			)

			wg := &sync.WaitGroup{}
			cc := tt.args.cc
			if p, ok := tt.args.pipe.(*pipeImpl); p != nil && ok {
				if cc.ctx != nil {
					l := tt.wantPkts.Len()

					pp := NewNozzlePacketProcessor(tt.args.pipe.Nozzle(), l)
					wg.Add(1)
					go func() {
						pp.Run(tt.args.cc.ctx)
						wg.Done()
						if !tt.isTimeout {
							cc.cancel()
						}
					}()
				}

				if tt.clearChannel {
					p.out = nil
				}
			}

			err = tt.args.pipe.Run(tt.args.cc.ctx)

			wg.Wait()

			err = test.CheckTestResults(pkts, tt.wantPkts, err, tt.wantErr)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestNewPipe(t *testing.T) {
	type args struct {
		config PipeConfig
	}
	tests := []struct {
		name string
		args args
		want *pipeImpl
	}{
		{
			name: "Config Timeout Provided",
			args: args{
				config: PipeConfig{
					TimeoutConfig: &TimeoutConfig{
						ReadTimeout:  time.Second,
						WriteTimeout: time.Second,
					},
				},
			},
			want: NewPipe(
				PipeConfig{
					TimeoutConfig: &TimeoutConfig{
						ReadTimeout:  time.Second,
						WriteTimeout: time.Second,
					},
				},
			).(*pipeImpl),
		},
		{
			name: "Config Timeout Not Provided",
			args: args{
				config: PipeConfig{},
			},
			want: NewPipe(
				PipeConfig{
					TimeoutConfig: &TimeoutConfig{
						ReadTimeout:  DefaultReadTimeout,
						WriteTimeout: DefaultWriteTimeout,
					},
				},
			).(*pipeImpl),
		},
	}
	for _, tt := range tests {
		got := NewPipe(tt.args.config)
		t.Run(tt.name, func(t *testing.T) {
			err := test.CheckTestResults(got, tt.want, nil, nil)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_pipeImpl_Nozzle(t *testing.T) {
	nozzle := make(Nozzle)

	tests := []struct {
		name string
		p    *pipeImpl
		want Nozzle
	}{
		{
			name: "Happy Path",
			p: &pipeImpl{
				out: nozzle,
			},
			want: nozzle,
		},
		{
			name: "Nozzle is nil",
			p: &pipeImpl{
				out: nil,
			},
			want: nil,
		},
		{
			name: "Pipe is nil",
			p:    nil,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Nozzle(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pipeImpl.Nozzle() = %v, want %v", got, tt.want)
			}
		})
	}
}
