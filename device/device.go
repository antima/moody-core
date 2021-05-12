package device

import (
	"encoding/json"
	"net/http"
)

type Device interface {
	sync() bool
}

type MoodyDevice struct {
	isUp       bool
	ipAddress  string
	macAddress string
}

type Actuator struct {
	MoodyDevice
	state float64
}

type ConnectionPacket struct {
	DeviceType string `json:"type"`
	MacAddress string `json:"mac"`
	Service    string `json:"service"`
}

type DataPacket struct {
	Payload float64 `json:"payload"`
}

func (mDev *MoodyDevice) sync() bool {
	resp, err := http.Get(mDev.ipAddress)
	if err != nil {
		mDev.isUp = false
		return false
	}

	connPkt := ConnectionPacket{}
	if err := json.NewDecoder(resp.Body).Decode(&connPkt); err != nil {
		mDev.isUp = false
		return false
	}

	mDev.macAddress = connPkt.MacAddress
	return true
}

type Sensor struct {
	MoodyDevice
	service     string
	lastReading float64
}

func (s *Sensor) Read() float64 {
	// if timeout elapses just return the last reading
	return s.lastReading
}
