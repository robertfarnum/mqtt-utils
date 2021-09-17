package main

import (
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

func clientInterceptor(cp packets.ControlPacket, err error) (packets.ControlPacket, error) {
	printPacketInfo("RCVD", cp, err)

	return cp, err
}

func brokerInterceptor(cp packets.ControlPacket, err error) (packets.ControlPacket, error) {
	printPacketInfo("SENT", cp, err)

	return cp, err
}

func (server *Server) serve(clientConn net.Conn, w http.ResponseWriter, r *http.Request) error {
	if server.IsDebug {
		fmt.Printf("New connection: %v\n", clientConn.RemoteAddr())
		fmt.Printf("Connecting to: %s\n", server.Broker)
	}

	brokerConn, err := server.getBrokerConn(w, r)
	if err != nil {
		return err
	}

	defer brokerConn.Close()
	defer clientConn.Close()

	proxy := &Proxy{
		ClientConn:        clientConn,
		ClientInterceptor: clientInterceptor,
		BrokerConn:        brokerConn,
		BrokerInterceptor: brokerInterceptor,
	}

	return proxy.Start()
}

func (server *Server) getBrokerConn(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
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
func (server *Server) serveWebsocket(w http.ResponseWriter, r *http.Request) error {
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

	go server.serve(connector, w, r)

	return nil
}

func (server *Server) startWebsocketServer() {
	http.HandleFunc("/mqtt", func(w http.ResponseWriter, r *http.Request) {
		err := server.serveWebsocket(w, r)
		if err != nil {
			log.Println(err)
		}
	})

	err := http.ListenAndServe(server.WSListen, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (server *Server) startTCPServer() {
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

			go server.serve(conn, nil, nil)
		}
	}
}

func (server *Server) Start() {
	if server.IsDebug {
		fmt.Println("verbose mode enabled")
	}

	if server.IsTrace {
		fmt.Println("trace is enabled")
	}

	go server.startTCPServer()

	server.startWebsocketServer()
}
