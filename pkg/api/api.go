package api

import (
	"encoding/json"
	"log"
	"net/http"

	httpIfc "github.com/antima/moody-core/pkg/http"
	"github.com/gorilla/mux"
)

var (
	devices *httpIfc.DeviceList
)

type DevicesResp struct {
	Devices []string `json:"devices"`
}

type DeviceResp struct {
	httpIfc.Node
	Type string `json:"type"`
}

func MoodyApi(deviceList *httpIfc.DeviceList, port string) {
	if deviceList == nil {
		panic("MoodyApi: device list can't be nil")
	}

	devices = deviceList

	router := mux.NewRouter()
	router.HandleFunc("/api/device", getDevices).Methods("GET")
	router.HandleFunc("/api/device/{url}", getDevice).Methods("GET")
	router.HandleFunc("/api/sensor/{url}", getSensorData).Methods("GET")
	router.HandleFunc("/api/actuator/{url}", getActuatorData).Methods("GET")
	router.HandleFunc("/api/actuator/{url}", putActuatorData).Methods("PUT")

	log.Printf("starting the API server on port %s\n", port)
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
	dev, exists := devices.Get(vars["url"])
	if dev == nil || !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var devResp DeviceResp
	switch dev.(type) {
	case *httpIfc.Sensor:
		devResp = DeviceResp{Node: dev.(*httpIfc.Sensor).Node}
		devResp.Type = "sensor"
	case *httpIfc.Actuator:
		devResp = DeviceResp{Node: dev.(*httpIfc.Actuator).Node}
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
	dev, exists := devices.Get(vars["url"])
	if dev == nil || !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sensor, isSensor := dev.(*httpIfc.Sensor)
	if !isSensor {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := sensor.Read()
	dataResp := httpIfc.DataPacket{Payload: data}
	if err := json.NewEncoder(w).Encode(&dataResp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getActuatorData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	vars := mux.Vars(r)
	dev, exists := devices.Get(vars["url"])
	if dev == nil || !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	actuator, isActuator := dev.(*httpIfc.Actuator)
	if !isActuator {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	state := actuator.State()
	dataResp := httpIfc.DataPacket{Payload: state}
	if err := json.NewEncoder(w).Encode(&dataResp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func putActuatorData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	vars := mux.Vars(r)
	dev, exists := devices.Get(vars["url"])
	if dev == nil || !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := httpIfc.DataPacket{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	actuator, isActuator := dev.(*httpIfc.Actuator)
	if !isActuator {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	actuator.Actuate(data.Payload)
	if err := json.NewEncoder(w).Encode(&data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
