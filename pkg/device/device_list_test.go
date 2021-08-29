package device

import (
	"sync"
	"testing"
	"time"
)

func TestNewDeviceList(t *testing.T) {
	list := NewDeviceList()
	if list == nil || list.devices == nil {
		t.Errorf("got nil device list")
	}
}

func TestDeviceList_Attach(t *testing.T) {
	obsChan := make(chan DeviceMsg)
	list := NewDeviceList()
	list.Attach(obsChan)

	for _, observer := range list.observers {
		if observer == obsChan {
			return
		}
	}
	t.Errorf("expected chan into device list observers, got not found")
}

func TestDeviceList_Add(t *testing.T) {
	ip := "127.0.0.1"
	dev := &Sensor{Node: Node{
		IpAddress: ip,
	}}
	obsChan := make(chan DeviceMsg, 1)
	list := NewDeviceList()
	list.Attach(obsChan)
	list.Add(ip, dev)

	var wg sync.WaitGroup
	wg.Add(3)

	go func(device *Sensor, devices map[string]Device, wg *sync.WaitGroup) {
		defer wg.Done()
		for ipAddr := range devices {
			if ipAddr == device.IpAddress {
				return
			}
		}
		t.Errorf("expected ip into device list, got not found")
	}(dev, list.devices, &wg)

	go func(device *Sensor, ips []string, wg *sync.WaitGroup) {
		defer wg.Done()
		for _, ipAddr := range ips {
			if ipAddr == device.IpAddress {
				return
			}
		}
		t.Errorf("expected ip into device list cache, got not found")
	}(dev, list.namesCache, &wg)

	go func(ch chan DeviceMsg, wg *sync.WaitGroup) {
		defer wg.Done()
		select {
		case <-time.After(2 * time.Second):
			t.Errorf("expected device message, got nothing")
		case <-obsChan:
		}
	}(obsChan, &wg)
	wg.Wait()
}
