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

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

type Server struct {
	TCPListen string
	WSListen  string
	Broker    string
	IsDebug   bool
	IsTrace   bool
}

func printPacketInfo(action string, cp packets.ControlPacket, err error) {
	now := time.Now().UTC()

	color.Blue("%s at %s:\n", action, now.String())

	if err != nil {
		color.Red("	Error: %v\n", err)
	} else if cp != nil {
		color.Green("	Packet: %s\n", cp.String())
		color.Green("	Details: %v\n", cp.Details())
	}

	fmt.Println()
}

func clientInterceptor(ctx context.Context, cp packets.ControlPacket, err error) (packets.ControlPacket, error) {
	printPacketInfo("RCVD", cp, err)

	return cp, err
}

func brokerInterceptor(ctx context.Context, cp packets.ControlPacket, err error) (packets.ControlPacket, error) {
	printPacketInfo("SENT", cp, err)

	return cp, err
}

func (server *Server) serve(ctx context.Context, clientConn net.Conn, w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if server.IsDebug {
		fmt.Printf("New connection: %v\n", clientConn.RemoteAddr())
		fmt.Printf("Connecting to: %s\n", server.Broker)
	}

	brokerConn, err := server.getBrokerConn(ctx, w, r)
	if err != nil {
		return err
	}

	defer brokerConn.Close()
	defer clientConn.Close()

	// proxy := &Proxy{
	// 	ClientConn:        clientConn,
	// 	ClientInterceptor: clientInterceptor,
	// 	BrokerConn:        brokerConn,
	// 	BrokerInterceptor: brokerInterceptor,
	// }

	//return proxy.Start(ctx)

	return nil
}

func (server *Server) getBrokerConn(_ context.Context, w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	// first open a connection to the remote broker
	uri, err := url.Parse(server.Broker)
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

	conn, err := mqtt.OpenConnection(uri, tlsConfig, time.Duration(time.Second*10), header, nil)
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
func (server *Server) serveWebsocket(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

	connector := &mqtt.WebsocketConnector{
		Conn: conn,
	}

	connector.SetDeadline(time.Time{})

	go server.serve(ctx, connector, w, r)

	return nil
}

func (server *Server) startWebsocketServer(ctx context.Context) {
	http.HandleFunc("/mqtt", func(w http.ResponseWriter, r *http.Request) {
		err := server.serveWebsocket(ctx, w, r)
		if err != nil {
			log.Println(err)
		}
	})

	err := http.ListenAndServe(server.WSListen, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (server *Server) startTCPServer(ctx context.Context) {
	if server.TCPListen != "" {
		listener, err := net.Listen("tcp", server.TCPListen)
		if err != nil {
			panic(err)
		}

		for {
			conn, err := listener.Accept()
			if err != nil {
				panic(err)
			}

			go server.serve(ctx, conn, nil, nil)
		}
	}
}

func (server *Server) Start(ctx context.Context) {
	if server.IsDebug {
		fmt.Println("verbose mode enabled")
	}

	if server.IsTrace {
		fmt.Println("trace is enabled")
	}

	go server.startTCPServer(ctx)

	server.startWebsocketServer(ctx)
}
