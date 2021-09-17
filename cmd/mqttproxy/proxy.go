package main

import (
	"fmt"
	"io"
	"net"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

type proxyChanData struct {
	cp  packets.ControlPacket
	err error
}

type Proxy struct {
	// The client connection and interceptor.
	ClientConn        net.Conn
	ClientInterceptor func(cp packets.ControlPacket, err error) (packets.ControlPacket, error)

	// The broker connection and iterceptor.
	BrokerConn        net.Conn
	BrokerInterceptor func(cp packets.ControlPacket, err error) (packets.ControlPacket, error)
}

// processStream reads an MQTT ControlPacket, forwards it to the interceptor, and forwards the result to the proxy channel
func processStream(reader io.Reader, interceptor func(cp packets.ControlPacket, err error) (packets.ControlPacket, error)) chan proxyChanData {
	// Create the proxy channel
	proxyChan := make(chan proxyChanData, 1024)

	go func() {
		for {
			cp, err := packets.ReadPacket(reader)

			if interceptor != nil {
				cp, err = interceptor(cp, err)
			}

			proxyChan <- proxyChanData{
				cp:  cp,
				err: err,
			}
		}
	}()

	return proxyChan
}

func (proxy *Proxy) Start() error {
	if proxy.ClientConn == nil {
		return fmt.Errorf("client connection is nil")
	}

	if proxy.BrokerConn == nil {
		return fmt.Errorf("broker connection is nil")
	}

	clientChan := processStream(proxy.ClientConn, proxy.ClientInterceptor)
	brokerChan := processStream(proxy.BrokerConn, proxy.BrokerInterceptor)

	for {
		select {
		case clientData := <-clientChan:
			if clientData.err != nil {
				fmt.Println(clientData.err.Error())
				return clientData.err
			}

			clientData.cp.Write(proxy.BrokerConn)
		case brokerData := <-brokerChan:
			if brokerData.err != nil {
				fmt.Println(brokerData.err.Error())
				return brokerData.err
			}

			brokerData.cp.Write(proxy.ClientConn)
		}
	}
}
