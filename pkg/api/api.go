package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/antima/moody-core/pkg/device"

	"github.com/antima/moody-core"

	"github.com/gorilla/mux"
)

var (
	devices *device.DeviceList
)

type DevicesResp struct {
	Devices []string `json:"devices"`
}

type DeviceResp struct {
	device.Node
	Type string `json:"type"`
}

func moodyApi(deviceList *device.DeviceList, port string) {
	if deviceList == nil {
		panic("Device list can't be nil")
	}
	router := mux.NewRouter()
	router.HandleFunc("/api/device", getDevices).Methods("GET")
	router.HandleFunc("/api/device/{url}", getDevice).Methods("GET")
	router.HandleFunc("/api/sensor/{url}", getSensorData).Methods("GET")
	router.HandleFunc("/api/actuator/{url}", getActuatorData).Methods("GET")
	router.HandleFunc("/api/actuator/{url}", putActuatorData).Methods("PUT")
	log.Fatal(http.ListenAndServe(port, router))
}

func getDevices(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-type", "application/json")
	devs := DevicesResp{Devices: devices.ConnectedIPs()}
	if err := json.NewEncoder(w).Encode(devs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	vars := mux.Vars(r)
	dev, exists := moody.Devices.Get(vars["url"])
	if dev == nil || !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var devResp DeviceResp
	switch dev.(type) {
	case *device.Sensor:
		devResp = DeviceResp{Node: dev.(*device.Sensor).Node}
		devResp.Type = "sensor"
	case *device.Actuator:
		devResp = DeviceResp{Node: dev.(*device.Actuator).Node}
		devResp.Type = "actuator"
	default:
		break
	}

	if err := json.NewEncoder(w).Encode(&devResp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getSensorData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	vars := mux.Vars(r)
	dev, exists := moody.Devices.Get(vars["url"])
	if dev == nil || !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sensor, isSensor := dev.(*device.Sensor)
	if !isSensor {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := sensor.Read()
	dataResp := device.DataPacket{Payload: data}
	if err := json.NewEncoder(w).Encode(&dataResp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getActuatorData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	vars := mux.Vars(r)
	dev, exists := moody.Devices.Get(vars["url"])
	if dev == nil || !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	actuator, isActuator := dev.(*device.Actuator)
	if !isActuator {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	state := actuator.State()
	dataResp := device.DataPacket{Payload: state}
	if err := json.NewEncoder(w).Encode(&dataResp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func putActuatorData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	vars := mux.Vars(r)
	dev, exists := moody.Devices.Get(vars["url"])
	if dev == nil || !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := device.DataPacket{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	actuator, isActuator := dev.(*device.Actuator)
	if !isActuator {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	actuator.Actuate(data.Payload)
	if err := json.NewEncoder(w).Encode(&data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
