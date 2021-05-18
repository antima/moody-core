package moody

import (
	"sync"
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
	Device Device
	Event  DeviceEvent
}

type DeviceList struct {
	devices   map[string]Device
	observers []chan<- DeviceMsg
	mutex     sync.Mutex
}

func NewDeviceList() *DeviceList {
	return &DeviceList{
		devices: make(map[string]Device),
	}
}

func (list *DeviceList) Attach(obsChan chan<- DeviceMsg) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	list.observers = append(list.observers, obsChan)
}

func (list *DeviceList) Add(ip string, device Device) {
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
