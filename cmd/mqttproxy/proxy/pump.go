package proxy

import (
	"context"
	"net"
	"sync"
)

// Pump
type Pump struct {
	// The bi-directional network connection
	conn net.Conn

	// The hose to pump the MQTT packet Gas from the network connection to the channel
	hose Hose
}

// NewPump create a new bi-directional communication channel between the client and broker
func NewPump(conn net.Conn) Pump {
	return Pump{
		conn: conn,
		hose: NewHose(conn),
	}
}

func (p Pump) pump(ctx context.Context) {

}

// Station
type Station struct {
	client Pump
	broker Pump
}

// NewStation create a new bi-directional communication channel between the client and broker
func NewStation(client net.Conn, broker net.Conn) Station {
	return Station{
		client: NewPump(client),
		broker: NewPump(broker),
	}
}

// GetClientChan returns the client output channel
func (s Station) GetClientChan() Nozzle {
	return s.client.hose.out
}

// GetBrokerChan returns the broker output channel
func (s Station) GetBrokerChan() Nozzle {
	return s.broker.hose.out
}

// Start the pump; stops only after the pipes are drained
func (s *Station) Open(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()
	defer s.client.conn.Close()
	defer s.broker.conn.Close()

	//errs := make([]error, 2)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	flow(ctx, s.client.hose)

	wg.Add(2)
	flow(ctx, s.broker.hose)

	// go func() {
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return nil
	// 		case clientData := <-clientChan:
	// 			if clientData.err != nil {
	// 				fmt.Println(clientData.err.Error())
	// 				return clientData.err
	// 			}

	// 			clientData.cp.Write(proxy.BrokerConn)
	// 		}
	// 	}
	// }()

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
