package device

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Device interface {
	sync() bool
}

type MoodyDevice struct {
	isUp       bool
	ipAddress  string
	macAddress string
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

	defer func(body io.ReadCloser) { _ = body.Close() }(resp.Body)

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
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(s.ipAddress + "/api/data")
	if err != nil {
		return s.lastReading
	}

	defer func(body io.ReadCloser) { _ = body.Close() }(resp.Body)

	if err := json.NewDecoder(resp.Body).Decode(&s.lastReading); err != nil {
		return s.lastReading
	}

	return s.lastReading

}

type Actuator struct {
	MoodyDevice
	state float64
}

func (a *Actuator) Actuate(action float64) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	actionPacket := DataPacket{
		Payload: action,
	}
	actionBytes, err := json.Marshal(&actionPacket)
	if err != nil {
		return
	}

	req, err := http.NewRequest("PUT", a.ipAddress+"/api/data", bytes.NewReader(actionBytes))
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer func(body io.ReadCloser) { _ = body.Close() }(resp.Body)
	if err := json.NewDecoder(resp.Body).Decode(&actionPacket); err != nil {
		return
	}

	a.state = actionPacket.Payload
}
