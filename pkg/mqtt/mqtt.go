package mqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client

type PublishFunc func(string, string)

func StartMqttManager() {
	clientOpts := mqtt.ClientOptions{}
	client = mqtt.NewClient(&clientOpts)
}

func Publish(payload string, topic string) {

}
