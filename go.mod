module github.com/robertfarnum/mqtt-utils

go 1.16

require (
	github.com/aws/aws-sdk-go v1.38.55
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/fatih/color v1.12.0
	github.com/gorilla/websocket v1.4.2
)

replace github.com/eclipse/paho.mqtt.golang => ../paho.mqtt.golang
