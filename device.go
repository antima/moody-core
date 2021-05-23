package moody

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

var (
	NodeConnectionError  = errors.New("could not establish a connection with the device")
	UnsupportedNodeError = errors.New("unsupported node type")
)

type Endpoint string

const (
	ConnectionEndpoint Endpoint = "/api/conn"
	DataEndpoint       Endpoint = "/api/data"
)

func getEndpointData(ip string, remote Endpoint, dest interface{}) bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	// select protocol maybe
	resp, err := client.Get("http://" + ip + string(remote))
	if err != nil {
		return false
	}

	defer func(body io.ReadCloser) { _ = body.Close() }(resp.Body)

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return false
	}
	return true
}

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
	Sync() bool
}

// A Node is a generic remote device in the WSAN that implements the basic moody protocol
// exposing tha /api/conn endpoint
type Node struct {
	isUp       bool
	ipAddress  string
	macAddress string
	service    string
}

// NewDevice initializes a device for the first time from an ip string, returning an error
// if the ip is unreachable, returns a badly formatted response or an unrecognized node type.
func NewDevice(ip string) (Device, error) {
	connPkt := ConnectionPacket{}
	res := getEndpointData(ip, ConnectionEndpoint, &connPkt)
	if !res {
		return nil, NodeConnectionError
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
		return nil, UnsupportedNodeError
	}
}

// A Sensor is a particular type of Node that can be queried for sensed data
type Sensor struct {
	Node
	lastReading float64
}

func (s *Sensor) Read() float64 {
	s.Sync()
	return s.lastReading
}

// Sync attempts to get a new reading from the remote Sensor and either returns the new
// data if the Sensor responds, or returns the last successful reading
func (s *Sensor) Sync() bool {
	dataPkt := DataPacket{}
	res := getEndpointData(s.ipAddress, DataEndpoint, &dataPkt)
	if res {
		s.lastReading = dataPkt.Payload
	}
	s.isUp = res
	return res
}

// An Actuator describes a node that is using the Moody Actuator object as its
// fw on a remote device
type Actuator struct {
	Node
	stateSynced bool
	state       float64
}

func (a *Actuator) SetState(state float64) {
	if state != a.state {
		a.state = state
		a.Sync()
	}
}

func (a *Actuator) Sync() bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	actionPacket := DataPacket{
		Payload: a.state,
	}
	actionBytes, err := json.Marshal(&actionPacket)
	if err != nil {
		return false
	}

	req, err := http.NewRequest("PUT", a.ipAddress+"/api/data", bytes.NewReader(actionBytes))
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	defer func(body io.ReadCloser) { _ = body.Close() }(resp.Body)
	if err := json.NewDecoder(resp.Body).Decode(&actionPacket); err != nil {
		return false
	}

	a.state = actionPacket.Payload
	return true
}
