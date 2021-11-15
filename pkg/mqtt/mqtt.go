package mqtt

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	connectionRetries   = 5
	baseTopic           = "moody/device/#"
	brokerConnectionErr = "could not connect to the mqtt broker"
)

var (
	client mqtt.Client
)

type PublishFunc func(string, string)

func StartMqttManager(brokerString string, dataTable *DataTable) error {
	clientOpts := mqtt.ClientOptions{}
	clientOpts.AddBroker(brokerString)
	clientOpts.SetAutoReconnect(true)
	clientOpts.SetOnConnectHandler(subscribe)

	client = mqtt.NewClient(&clientOpts)
	if err := connect(); err != nil {
		log.Fatal(brokerConnectionErr)
	}

	return nil
}

func Publish(payload string, topic string) {
	token := client.Publish(topic, 0, true, payload)
	if token.Wait() && token.Error() != nil {
		// TODO
	}
}

func connect() error {
	retries := 0
	for retries < connectionRetries {
		token := client.Connect()
		if token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

func subscribe(client mqtt.Client) {
	subscribed := false
	for subscribed {
		token := client.Subscribe(baseTopic, 0, dataCallback)
		if token.Wait() && token.Error() != nil {
			continue
		}
	}
}

func dataCallback(c mqtt.Client, m mqtt.Message) {

}
