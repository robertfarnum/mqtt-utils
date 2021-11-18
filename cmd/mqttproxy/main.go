package main

import (
	"context"
	"flag"
	"fmt"
)

func main() {
	service := &Service{}

	fmt.Println("mqttproxy: An MQTT proxy")

	flag.StringVar(&service.TCPListen, "ts", "", "the MQTT TCP address and port to listen for MQTT proxy clients")
	flag.StringVar(&service.WSListen, "ws", "0.0.0.0:443", "the MQTT WSS address and port to listen for MQTT proxy clients")

	flag.StringVar(&service.Broker, "b", "", "the MQTT address and port to connect to the MQTT service")
	flag.BoolVar(&service.IsDebug, "v", false, "dumps verbose debug information")
	flag.BoolVar(&service.IsTrace, "t", false, "trace every communication to JSON files")

	flag.Parse()

	fmt.Printf("%v\n", service)

	service.Start(context.Background())
}
