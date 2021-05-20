package moody

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

// A ConnectionPacket represents a packet returned by a conn node endpoint
type ConnectionPacket struct {
	DeviceType string `json:"type"`
	MacAddress string `json:"mac"`
	Service    string `json:"service"`
}

// A DataPacket represents a packet returned by a data node endpoint
type DataPacket struct {
	Payload float64 `json:"payload"`
}

// A Device is a virtualization of a remote machine that can be synced
type Device interface {
	sync() bool
}

// A Node is a generic remote device in the WSAN that implements the basic moody protocol
// exposing tha /api/conn endpoint
type Node struct {
	isUp       bool
	ipAddress  string
	macAddress string
	service    string
}

func (mDev *Node) sync() bool {
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

	mDev.isUp = true
	return true
}

// NewDevice initializes a device for the first time from an ip string, returning an error
// if the ip is unreachable, returns a badly formatted response or an unrecognized node type.
func NewDevice(ip string) (Device, error) {
	resp, err := http.Get("http://" + ip + "/api/conn")
	if err != nil {
		return nil, err
	}

	defer func(body io.ReadCloser) { _ = body.Close() }(resp.Body)
	connPkt := ConnectionPacket{}
	if err := json.NewDecoder(resp.Body).Decode(&connPkt); err != nil {
		return nil, err
	}

	baseDev := Node{
		isUp:       true,
		ipAddress:  ip,
		macAddress: connPkt.MacAddress,
		service:    connPkt.Service,
	}

	switch connPkt.DeviceType {
	case "sensor":
		return &Sensor{
			Node:        baseDev,
			lastReading: 0,
		}, nil
	case "actuator":
		return &Actuator{
			Node:  baseDev,
			state: 0,
		}, nil
	default:
		return nil, errors.New("unsupported node type")
	}
}

// A Sensor is a particular type of Node that can be queried for sensed data
type Sensor struct {
	Node
	lastReading float64
}

// Read attempts to get a new reading from the remote Sensor and either returns the new
// data if the Sensor responds, or returns the last successful reading
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
	Node
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
