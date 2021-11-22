package mqtt

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	connectionRetries = 5
	baseTopic         = "moody/device/#"
)

var (
	client mqtt.Client

	// this may be changed in some way not to be global state
	dataTable *DataTable
)

type PublishFunc func(string, string)

func StartMqttManager(brokerString string, dataTableRef *DataTable) error {
	clientOpts := mqtt.ClientOptions{}
	clientOpts.AddBroker(brokerString)
	clientOpts.SetClientID(fmt.Sprintf("Moody-Recv"))
	clientOpts.SetAutoReconnect(true)
	clientOpts.SetOnConnectHandler(subscribe)

	client = mqtt.NewClient(&clientOpts)
	if err := connect(); err != nil {
		log.Fatal(err)
	}

	dataTable = dataTableRef
	return nil
}

func Publish(payload string, topic string) {
	token := client.Publish(topic, 0, true, payload)
	if token.Wait() && token.Error() != nil {
		// TODO
	}
}

func connect() error {
	var token mqtt.Token
	opts := client.OptionsReader()
	for retries := 0; retries < connectionRetries; retries += 1 {
		log.Printf("attempting mqtt connection #%d to server %s\n", retries+1, opts.Servers()[0])
		token = client.Connect()
		if token.Wait() && token.Error() != nil {
			retries += 1
			continue
		}
		log.Printf("succesfully connected to the mqtt broker @%s!", opts.Servers()[0])
		return nil
	}
	return token.Error()
}

func subscribe(client mqtt.Client) {
	token := client.Subscribe(baseTopic, 0, dataCallback)
	for token.Wait() && token.Error() != nil {
	}
	log.Printf("succesfully subscribed to the %s topic\n", baseTopic)
}

func dataCallback(c mqtt.Client, m mqtt.Message) {
	if dataTable != nil {
		topic := m.Topic()
		payload := string(m.Payload())
		log.Printf("received MQTT message from topic %s, with payload: %s\n", topic, payload)
		dataTable.Add(topic, payload)
	}
}
