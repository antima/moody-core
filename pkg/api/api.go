package api

import (
	"context"
	"fmt"
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

func StartMoodyApi(deviceList *httpIfc.DeviceList, port string) *http.Server {
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
	fmt.Println("aaa")
}
