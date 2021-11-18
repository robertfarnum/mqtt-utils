package main

import (
	"context"
	"fmt"
	"net"

	"github.com/robertfarnum/mqtt-utils/cmd/mqttproxy/proxy"
)

type Proxy struct {
	client net.Conn
	broker net.Conn
}

func (p Proxy) Start(ctx context.Context) error {
	s := proxy.NewStation(p.client, p.broker)
	clientNozzle := s.GetClientPump().GetNozzle()
	brokerNozzle := s.GetBrokerPump().GetNozzle()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case packet := <-clientNozzle:
				if packet.GetError() != nil {
					fmt.Println(clientData.err.Error())
					return clientData.err
				}

				clientData.cp.Write(proxy.BrokerConn)
			}
		}
	}()

	// go func() {
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return nil
	// 		case brokerData := <-brokerChan:
	// 			if brokerData.err != nil {
	// 				fmt.Println(brokerData.err.Error())
	// 				return brokerData.err
	// 			}

	// 			brokerData.cp.Write(proxy.ClientConn)
	// 		}
	// 	}
	// }()

	return nil
}
