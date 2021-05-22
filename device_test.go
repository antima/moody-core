package moody

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func randomString(length uint) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	randB := make([]byte, length)
	for idx := range randB {
		randB[idx] = letters[rand.Intn(len(letters))]
	}
	return string(randB)
}

func openPort() string {
	mockSocket, _ := net.Listen("tcp", ":0")
	port := fmt.Sprintf(":%s", strings.Split(mockSocket.Addr().String(), "]:")[1])
	_ = mockSocket.Close()
	return port
}

func mockOkConn(port *string) {
	router := mux.NewRouter()
	router.HandleFunc("/api/conn", func(w http.ResponseWriter, r *http.Request) {
		conn := ConnectionPacket{
			DeviceType: "sensor",
			MacAddress: "aa:aa:aa:aa:aa:aa",
			Service:    "example",
		}
		_ = json.NewEncoder(w).Encode(&conn)
	})

	*port = openPort()
	_ = http.ListenAndServe(*port, router)
}

func mockWrongFormatConn(port *string) {
	router := mux.NewRouter()
	router.HandleFunc("/api/conn", func(w http.ResponseWriter, r *http.Request) {
		conn := struct {
			DeviceType string `json:"type"`
			Service    string `json:"service"`
		}{
			DeviceType: "sensor",
			Service:    "example",
		}
		_ = json.NewEncoder(w).Encode(&conn)
	})

	*port = openPort()
	_ = http.ListenAndServe(*port, router)
}

func mockWrongDeviceTypeConn(port *string) {
	router := mux.NewRouter()
	router.HandleFunc("/api/conn", func(w http.ResponseWriter, r *http.Request) {
		conn := ConnectionPacket{
			DeviceType: "notok",
			MacAddress: "aa:aa:aa:aa:aa:aa",
			Service:    "example",
		}
		_ = json.NewEncoder(w).Encode(&conn)
	})

	*port = openPort()
	_ = http.ListenAndServe(*port, router)
}

func nillableErrorString(err error) string {
	if err == nil {
		return "nil"
	}
	return err.Error()
}

func TestNewDevice(t *testing.T) {
	var okPort string
	var okWrongFmtPort string
	var okWrongTypePort string

	go mockOkConn(&okPort)
	go mockWrongFormatConn(&okWrongFmtPort)
	go mockWrongDeviceTypeConn(&okWrongTypePort)

	for okPort == "" || okWrongFmtPort == "" || okWrongTypePort == "" {
	}

	testCases := []struct {
		Ip       string
		Expected error
	}{
		{randomString(10), InvalidIPError},
		{"192.168.1.1", NodeConnectionError},
		{"192.168.1.1" + okWrongFmtPort, NodeConnectionError},
		{"192.168.1.1" + okWrongTypePort, UnsupportedNodeError},
		{"192.168.1.1" + okPort, nil},
	}

	for _, test := range testCases {
		_, err := NewDevice(test.Ip)
		if err != test.Expected {
			t.Errorf("got %s, expected %s", nillableErrorString(err), nillableErrorString(test.Expected))
		}
	}
}
