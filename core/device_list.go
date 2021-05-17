package core

import (
	"sync"

	"github.com/antima/moody-core/device"
)

type DeviceEvent uint

const (
	EventAdded DeviceEvent = iota
	EventRemoved
)

type Observer interface {
	ListenForUpdates()
}

type DeviceMsg struct {
	Device device.Device
	Event  DeviceEvent
}

type DeviceList struct {
	devices   map[string]device.Device
	observers []chan<- DeviceMsg
	mutex     sync.Mutex
}

func (list *DeviceList) Attach(obsChan chan<- DeviceMsg) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	list.observers = append(list.observers, obsChan)
}

func (list *DeviceList) Add(ip string, device device.Device) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	list.devices[ip] = device

	devMsg := DeviceMsg{
		Device: device,
		Event:  EventAdded,
	}
	for _, observer := range list.observers {
		observer <- devMsg
	}
}

func (list *DeviceList) Remove(ip string) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	dev, exists := list.devices[ip]
	if !exists {
		return
	}

	devMsg := DeviceMsg{
		Device: dev,
		Event:  EventRemoved,
	}

	delete(list.devices, ip)
	for _, observer := range list.observers {
		observer <- devMsg
	}
}
