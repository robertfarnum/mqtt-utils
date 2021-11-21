package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/robertfarnum/mqtt-utils/pkg/network"
	"github.com/robertfarnum/mqtt-utils/pkg/proxy"

	"github.com/gorilla/websocket"
)

// Service listens for the incoming MQTT client connections and start the proxy
type Service struct {
	TCPListen string
	WSListen  string
	Broker    string
	IsDebug   bool
	IsTrace   bool
}

func (service *Service) serve(ctx context.Context, clientConn net.Conn, w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if service.IsDebug {
		fmt.Printf("New connection: %v\n", clientConn.RemoteAddr())
		fmt.Printf("Connecting to: %s\n", service.Broker)
	}

	brokerConn, err := service.getBrokerConn(ctx, w, r)
	if err != nil {
		return err
	}
	brokerEndpoint := proxy.NewEndpoint(proxy.EndpointConfig{
		Conn:      brokerConn,
		Processor: &brokerProcessor{},
	})
	defer brokerConn.Close()

	clientEndpoint := proxy.NewEndpoint(proxy.EndpointConfig{
		Conn:      clientConn,
		Processor: &clientProcessor{},
	})
	defer clientConn.Close()

	proxy := proxy.NewProxy(proxy.ChannelConfig{
		ClientEndpoint: clientEndpoint,
		BrokerEndpoint: brokerEndpoint,
	})

	return proxy.Run(ctx)
}

func (service *Service) getBrokerConn(_ context.Context, w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	// first open a connection to the remote broker
	uri, err := url.Parse(service.Broker)
	if err != nil {
		return nil, err
	}

	header := http.Header{}

	if r != nil {
		uri.RawQuery = r.URL.RawQuery
		//header = r.Header
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := network.OpenConnection(uri, tlsConfig, time.Duration(time.Second*10), header, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

const (
	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
)

// serve a connected MQTT Websocket client
func (service *Service) serveWebsocket(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 5 * time.Second,
		Subprotocols: []string{
			"mqtt",
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	connector := &network.WebsocketConnector{
		Conn: conn,
	}

	connector.SetDeadline(time.Time{})

	go func() {
		err := service.serve(ctx, connector, w, r)
		fmt.Printf("Finished serving client: %v\n", err)
	}()

	return nil
}

func (service *Service) startWebsocketService(ctx context.Context) {
	http.HandleFunc("/mqtt", func(w http.ResponseWriter, r *http.Request) {
		err := service.serveWebsocket(ctx, w, r)
		if err != nil {
			log.Println(err)
		}
	})

	err := http.ListenAndServe(service.WSListen, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (service *Service) startTCPService(ctx context.Context) {
	if service.TCPListen != "" {
		listener, err := net.Listen("tcp", service.TCPListen)
		if err != nil {
			panic(err)
		}

		for {
			conn, err := listener.Accept()
			if err != nil {
				panic(err)
			}

			go func() {
				err := service.serve(ctx, conn, nil, nil)
				fmt.Printf("Finished serving client: %v\n", err)
			}()
		}
	}
}

func (service *Service) Run(ctx context.Context) {
	if service.IsDebug {
		fmt.Println("verbose mode enabled")
	}

	if service.IsTrace {
		fmt.Println("trace is enabled")
	}

	go service.startTCPService(ctx)

	service.startWebsocketService(ctx)
}
