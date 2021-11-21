package main

import (
	"context"
	"fmt"
	"time"

	"github.com/robertfarnum/mqtt-utils/pkg/proxy"

	"github.com/fatih/color"
)

func printPacketInfo(action string, p proxy.Packet) {
	now := time.Now().UTC()

	color.Blue("%s at %s:\n", action, now.String())

	cp := p.GetControlPacket()
	es := p.GetErrStatus()

	if es != nil {
		color.Red("	Error: %v\n", es)
	} else if cp != nil {
		color.Green("	Packet: %s\n", cp.String())
		color.Green("	Details: %v\n", cp.Details())
	}

	fmt.Println()
}

type brokerProcessor struct{}

func (bp *brokerProcessor) processControlPacket(ctx context.Context, p *proxy.ProcessControlPacket) (*proxy.Packets, error) {
	switch p.GetPacketType() {
	case proxy.ReadyPacketType:
		color.Green("READY")
	}

	return nil, nil
}

func (bp *brokerProcessor) Process(ctx context.Context, p proxy.Packet) (*proxy.Packets, error) {
	printPacketInfo("RCVD", p)

	es := p.GetErrStatus()
	if es != nil {
		return nil, es
	}

	pkts := &proxy.Packets{}

	if pp, ok := p.GetControlPacket().(*proxy.ProcessControlPacket); ok {
		pps, err := bp.processControlPacket(ctx, pp)
		if err != nil {
			return pkts, err
		}
		pkts.AddPackets(pps)
	} else {
		np := proxy.NewPacket(p.GetControlPacket(), proxy.Forward)

		pkts.Add(np)
	}

	return pkts, nil
}

type clientProcessor struct{}

func (cp *clientProcessor) Process(ctx context.Context, p proxy.Packet) (*proxy.Packets, error) {
	printPacketInfo("SENT", p)

	es := p.GetErrStatus()
	if es != nil {
		return nil, es
	}

	np := proxy.NewPacket(p.GetControlPacket(), proxy.Forward)

	pkts := &proxy.Packets{}
	pkts.Add(np)

	return pkts, nil
}
