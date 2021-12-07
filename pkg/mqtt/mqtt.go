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

type MqttManager struct {
	client    mqtt.Client
	dataTable *DataTable
}

func StartMqttManager(brokerString string, dataTableRef *DataTable) *MqttManager {
	mgr := &MqttManager{}

	clientOpts := mqtt.ClientOptions{}
	clientOpts.AddBroker(brokerString)
	clientOpts.SetClientID(fmt.Sprintf("Moody-Recv"))
	clientOpts.SetAutoReconnect(true)
	clientOpts.SetOnConnectHandler(mgr.subscribe)
	clientOpts.SetConnectionLostHandler(mgr.lostConnectionHandler)
	client := mqtt.NewClient(&clientOpts)

	mgr.client = client
	mgr.dataTable = dataTableRef
	if err := mgr.connect(client); err != nil {
		log.Fatal(err)
	}

	return mgr
}

func (mgr *MqttManager) StopMqttManager() {
	log.Println("stopping the mqtt service")
	mgr.client.Unsubscribe(baseTopic)
	mgr.client.Disconnect(100)
}

func (mgr *MqttManager) Publish(payload string, topic string) {
	// TODO
	// if the actuate function from a service returns with a
	// send flag, this should be called with a specific topic
	// and payload obtained from the actuate return value
	token := mgr.client.Publish(topic, 0, true, payload)
	if token.Wait() && token.Error() != nil {
	}
}

func (mgr *MqttManager) connect(client mqtt.Client) error {
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

func (mgr *MqttManager) subscribe(c mqtt.Client) {
	opts := c.OptionsReader()
	log.Printf("succesfully connected to the mqtt broker @%s!", opts.Servers()[0])
	token := c.Subscribe(baseTopic, 0, mgr.dataCallback)
	for token.Wait() && token.Error() != nil {
	}
	log.Printf("succesfully subscribed to the %s topic\n", baseTopic)
}

func (mgr *MqttManager) dataCallback(c mqtt.Client, m mqtt.Message) {
	if mgr.dataTable != nil {
		topic := m.Topic()
		payload := string(m.Payload())
		log.Printf("received MQTT message from topic %s, with payload: %s\n", topic, payload)
		mgr.dataTable.Add(topic, payload)
	}
}

func (mgr *MqttManager) lostConnectionHandler(c mqtt.Client, e error) {
	opts := c.OptionsReader()
	log.Printf("lost connection with the broker @%s, trying to reconnect", opts.Servers()[0])
}
