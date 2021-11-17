package http

import (
	"sync"
)

type DeviceEvent uint

const (
	EventAdded DeviceEvent = iota
	EventRemoved
)

type DeviceMsg struct {
	Device Device
	Event  DeviceEvent
}

type DeviceList struct {
	changed    bool
	namesCache []string
	devices    map[string]Device
	observers  []chan<- DeviceMsg
	mutex      sync.Mutex
}

func NewDeviceList() *DeviceList {
	return &DeviceList{
		devices:    make(map[string]Device),
		namesCache: []string{},
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

	if _, exists := list.devices[ip]; exists {
		return
	}

	list.devices[ip] = device
	list.namesCache = append(list.namesCache, ip)

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
	list.changed = true
	for _, observer := range list.observers {
		observer <- devMsg
	}
}

func (list *DeviceList) Get(ip string) (Device, bool) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	dev, exists := list.devices[ip]
	return dev, exists
}

func (list *DeviceList) ConnectedIPs() []string {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	if list.changed {
		list.namesCache = list.namesCache[:0]
		for ip := range list.devices {
			list.namesCache = append(list.namesCache, ip)
		}
	}
	return list.namesCache
}
