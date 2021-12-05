package api

import (
	"context"
	"log"
	"net/http"

	httpIfc "github.com/antima/moody-core/pkg/http"
	"github.com/gorilla/mux"
)

type DevicesResp struct {
	Devices []string `json:"devices"`
}

type DeviceResp struct {
	httpIfc.Node
	Type string `json:"type"`
}

func StartMoodyApi(deviceList *httpIfc.DeviceList, port string) *http.Server {
	if deviceList == nil {
		panic("MoodyApi: device list can't be nil")
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/device", getDevices(deviceList)).Methods("GET")
	router.HandleFunc("/api/device/{url}", getDevice(deviceList)).Methods("GET")
	router.HandleFunc("/api/sensor/{url}", getSensorData(deviceList)).Methods("GET")
	router.HandleFunc("/api/actuator/{url}", getActuatorData(deviceList)).Methods("GET")
	router.HandleFunc("/api/actuator/{url}", putActuatorData(deviceList)).Methods("PUT")
	router.HandleFunc("/api/service", getServices).Methods("GET")

	log.Printf("starting the API server on port %s\n", port)
	server := &http.Server{Addr: port, Handler: router}
	go func(server *http.Server) {
		log.Fatal(server.ListenAndServe())
	}(server)
	return server
}

func StopMoodyApi(server *http.Server) {
	log.Println("stopping the API server")
	if err := server.Shutdown(context.TODO()); err != nil {
	}
}
