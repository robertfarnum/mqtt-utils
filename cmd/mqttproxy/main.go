package main

import (
	"context"
	"flag"
	"fmt"
)

func main() {
	server := &Server{}

	fmt.Println("mqttproxy: An MQTT proxy")

	flag.StringVar(&server.TCPListen, "ts", "", "the MQTT TCP address and port to listen for MQTT proxy clients")
	flag.StringVar(&server.WSListen, "ws", "0.0.0.0:443", "the MQTT WSS address and port to listen for MQTT proxy clients")

	flag.StringVar(&server.Broker, "b", "", "the MQTT address and port to connect to the MQTT server")
	flag.BoolVar(&server.IsDebug, "v", false, "dumps verbose debug information")
	flag.BoolVar(&server.IsTrace, "t", false, "trace every communication to JSON files")

	flag.Parse()

	fmt.Printf("%v\n", server)

	server.Start(context.Background())
}
