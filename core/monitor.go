package core

import (
	"log"
	"strings"

	"github.com/koron/go-ssdp"
)

var (
	monitorChan chan bool
	Devices     DeviceList
	Discovered  []string
)

func onAlive(m *ssdp.AliveMessage) {
	ip := strings.Split(m.From.String(), ":")[0]
	server := m.Server

	if strings.Contains(server, "Arduino") {
		Discovered = append(Discovered, ip)
	}
}

func ssdpMonitorStart() {
	monitor := ssdp.Monitor{
		Alive: onAlive,
	}
	err := monitor.Start()

	if err != nil {
		log.Fatal(err)
	}

	monitorChan = make(chan bool)
	<-monitorChan

	_ = monitor.Close()
}

func ssdpStop() {
	if monitorChan != nil {
		monitorChan <- true
	}
}
