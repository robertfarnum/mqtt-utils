package main

import (
	"context"
	"fmt"
	"net"

	"github.com/robertfarnum/mqtt-utils/pkg/proxy"
)

// Proxy for client and broker
type Proxy interface {
	Start(ctx context.Context) error
}

// proxyImpl implements the Proxy interface
type proxyImpl struct {
	client net.Conn
	broker net.Conn
}

// NewProxy creates a new proxy between the client and broker
func NewProxy(client net.Conn, broker net.Conn) Proxy {
	return &proxyImpl{
		client: client,
		broker: broker,
	}
}

func (p *proxyImpl) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	s := proxy.NewStation(p.client, p.broker)

	go func() {
		n := s.GetClientPump().Nozzle()

	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case pkt := <-n:
				es := pkt.GetErrStatus()
				if es != nil {
					fmt.Println(es.Error())
					return
				}

				cp := pkt.GetControlPacket()
				err := pkt.GetErrStatus()
				printPacketInfo("RCVD", cp, err)

				cp.Write(p.broker)
			}
		}
	}()

	go func() {
		n := s.GetBrokerPump().Nozzle()

	LOOP:
		for {
			select {
			case <-ctx.Done():
				break LOOP
			case pkt := <-n:
				es := pkt.GetErrStatus()
				if es != nil {
					fmt.Println(es.Error())
					return
				}

				cp := pkt.GetControlPacket()
				err := pkt.GetErrStatus()
				printPacketInfo("SENT", cp, err)

				cp.Write(p.client)
			}
		}
	}()

	err := s.Run(ctx, cancel)

	return err
}
