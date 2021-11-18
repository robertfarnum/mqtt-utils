package proxy

import (
	"context"
)

type Proxy struct{}

func (p Proxy) Start(ctx context.Context) error {
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
