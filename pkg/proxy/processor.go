package proxy

import "context"

// Processor interface used to process Packet(s)
type Processor interface {
	Process(ctx context.Context, p Packet) (pkts *Packets, err error)
}
