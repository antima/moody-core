package http

import (
	"log"
	"strings"
	"sync"

	"github.com/koron/go-ssdp"
)

type SsdpMonitor struct {
	DeviceList     *DeviceList
	notSyncedMutex sync.Mutex
	NotSynced      []string
	monitor        *ssdp.Monitor
}

// TODO removal, management, fn this only adds nodes
func NewMonitor(list *DeviceList) *SsdpMonitor {
	monitor := &SsdpMonitor{
		DeviceList: list,
		monitor:    &ssdp.Monitor{},
	}

	monitor.monitor.Alive = func(m *ssdp.AliveMessage) {
		log.Printf("Alive: From=%s Type=%s USN=%s Location=%s Server=%s MaxAge=%d",
			m.From.String(), m.Type, m.USN, m.Location, m.Server, m.MaxAge())
		ip := strings.Split(m.From.String(), ":")[0]
		server := m.Server

		if strings.Contains(server, "Arduino") {
			dev, err := NewDevice(ip)
			if err != nil {
				monitor.notSyncedMutex.Lock()
				defer monitor.notSyncedMutex.Unlock()
				monitor.NotSynced = append(monitor.NotSynced, ip)
				return
			}
			monitor.DeviceList.Add(ip, dev)
		}
	}

	return monitor

}

func (m *SsdpMonitor) Start() {
	err := m.monitor.Start()

	if err != nil {
		log.Fatal(err)
	}
}

func (m *SsdpMonitor) Stop() {
	_ = m.monitor.Close()
}
