package moody

import (
	"log"
	"strings"
	"sync"

	"github.com/koron/go-ssdp"
)

var (
	Devices = NewDeviceList()

	notSyncedMutex sync.Mutex
	NotSynced      []string

	monitor = &ssdp.Monitor{
		Alive: onAlive,
	}
)

func onAlive(m *ssdp.AliveMessage) {
	log.Printf("Alive: From=%s Type=%s USN=%s Location=%s Server=%s MaxAge=%d",
		m.From.String(), m.Type, m.USN, m.Location, m.Server, m.MaxAge())
	ip := strings.Split(m.From.String(), ":")[0]
	server := m.Server

	if strings.Contains(server, "Arduino") {
		dev, err := NewDevice(ip)
		if err != nil {
			notSyncedMutex.Lock()
			defer notSyncedMutex.Unlock()
			NotSynced = append(NotSynced, ip)
			return
		}
		Devices.Add(ip, dev)
	}
}

func ssdpMonitorStart() {
	err := monitor.Start()

	if err != nil {
		log.Fatal(err)
	}
}

func ssdpStop() {
	_ = monitor.Close()
}
