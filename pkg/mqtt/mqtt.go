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

func StartMqttManager(brokerString string, dataTableRef *DataTable) {
	clientOpts := mqtt.ClientOptions{}
	clientOpts.AddBroker(brokerString)
	clientOpts.SetClientID(fmt.Sprintf("Moody-Recv"))
	clientOpts.SetAutoReconnect(true)
	clientOpts.SetOnConnectHandler(subscribe)
	clientOpts.SetConnectionLostHandler(lostConnectionHandler)
	client = mqtt.NewClient(&clientOpts)
	if err := connect(); err != nil {
		log.Fatal(err)
	}

	dataTable = dataTableRef
}

func StopMqttManager() {
	log.Println("stopping the mqtt service")
	client.Unsubscribe(baseTopic)
	client.Disconnect(100)
}

func Publish(payload string, topic string) {
	// TODO
	// if the actuate function from a service returns with a
	// send flag, this should be called with a specific topic
	// and payload obtained from the actuate return value
	token := client.Publish(topic, 0, true, payload)
	if token.Wait() && token.Error() != nil {
	}
}

func connect() error {
	var token mqtt.Token
	opts := client.OptionsReader()
	for retries := 0; retries < connectionRetries; retries += 1 {
		log.Printf("attempting a connection #%d to the mqtt broker @%s\n", retries+1, opts.Servers()[0])
		token = client.Connect()
		if token.Wait() && token.Error() != nil {
			continue
		}
		return nil
	}
	return token.Error()
}

func subscribe(client mqtt.Client) {
	opts := client.OptionsReader()
	log.Printf("succesfully connected to the mqtt broker @%s!", opts.Servers()[0])
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

func lostConnectionHandler(c mqtt.Client, e error) {
	opts := client.OptionsReader()
	log.Printf("lost connection with the broker @%s, trying to reconnect", opts.Servers()[0])
}
