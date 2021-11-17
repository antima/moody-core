package http

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockOkConn() *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/api/conn", func(w http.ResponseWriter, r *http.Request) {
		conn := ConnectionPacket{
			DeviceType: "sensor",
			MacAddress: "aa:aa:aa:aa:aa:aa",
			Service:    "example",
		}
		_ = json.NewEncoder(w).Encode(&conn)
	})

	return httptest.NewServer(router)
}

func mockWrongDeviceTypeConn() *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/api/conn", func(w http.ResponseWriter, r *http.Request) {
		conn := ConnectionPacket{
			DeviceType: "notok",
			MacAddress: "aa:aa:aa:aa:aa:aa",
			Service:    "example",
		}
		_ = json.NewEncoder(w).Encode(&conn)
	})

	return httptest.NewServer(router)
}

func mockSensor() *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/api/conn", func(w http.ResponseWriter, r *http.Request) {
		conn := ConnectionPacket{
			DeviceType: "sensor",
			MacAddress: "aa:aa:aa:aa:aa:aa",
			Service:    "example",
		}
		_ = json.NewEncoder(w).Encode(&conn)
	})

	router.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		n := rand.Float64()*1000 + 10
		data := DataPacket{Payload: n}
		_ = json.NewEncoder(w).Encode(&data)
	})

	return httptest.NewServer(router)
}

func mockActuator() *httptest.Server {
	router := http.NewServeMux()
	router.HandleFunc("/api/conn", func(w http.ResponseWriter, r *http.Request) {
		conn := ConnectionPacket{
			DeviceType: "actuator",
			MacAddress: "aa:aa:aa:aa:aa:aa",
			Service:    "example",
		}
		_ = json.NewEncoder(w).Encode(&conn)
	})

	router.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		data := DataPacket{}
		_ = json.NewDecoder(r.Body).Decode(&data)
		defer func(body io.ReadCloser) { _ = body.Close() }(r.Body)
		_ = json.NewEncoder(w).Encode(&data)
	})

	return httptest.NewServer(router)
}

func nillableErrorString(err error) string {
	if err == nil {
		return "nil"
	}
	return err.Error()
}

func TestNewDevice(t *testing.T) {
	testCases := []struct {
		Ip         string
		ServerFunc func() *httptest.Server
		Expected   error
	}{
		{"192.168.1.1", nil, NodeConnectionError},
		{"", mockWrongDeviceTypeConn, UnsupportedNodeError},
		{"", mockOkConn, nil},
	}

	for _, test := range testCases {
		var ip string
		var server *httptest.Server
		var err error

		if test.ServerFunc != nil {
			server = test.ServerFunc()
			ip = server.URL
			ipStart := strings.Index(ip, "://") + 3
			ip = ip[ipStart:]
		} else {
			_, err = NewDevice(test.Ip)
			ip = test.Ip
		}

		_, err = NewDevice(ip)

		if err != test.Expected {
			t.Errorf("got %s, expected %s", nillableErrorString(err), nillableErrorString(test.Expected))
		}

		if server != nil {
			server.Close()
		}
	}
}

func TestSensor_Read(t *testing.T) {
	server := mockSensor()
	ipStart := strings.Index(server.URL, "://") + 3
	ip := server.URL[ipStart:]

	dev, _ := NewDevice(ip)
	sensor := dev.(*Sensor)
	val := sensor.Read()
	if val == 0 {
		t.Errorf("got 0, expected val != 0")
	}

	server.Close()
	newVal := sensor.Read()
	if newVal != val {
		t.Errorf("got %f, expected %f", newVal, val)
	}
}

func TestActuator_SetState(t *testing.T) {
	server := mockActuator()
	ipStart := strings.Index(server.URL, "://") + 3
	ip := server.URL[ipStart:]

	dev, _ := NewDevice(ip)
	actuator := dev.(*Actuator)
	val := 1500.0
	actuator.Actuate(val)

	if actuator.State() != val {
		t.Errorf("expected %f, got %f", val, actuator.State())
	}

}
