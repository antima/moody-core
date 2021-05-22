package moody

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type DevicesResp struct {
	Devices []string `json:"devices"`
}

type DeviceResp struct {
}

func moodyApi(port string) {
	router := mux.NewRouter()

	router.HandleFunc("/api/device", getDevices).Methods("GET")
	router.HandleFunc("/api/device/{url}", getDevice).Methods("GET")
	log.Fatal(http.ListenAndServe(port, router))
}

func getDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	devs := DevicesResp{Devices: Devices.ConnectedIPs()}
	if err := json.NewEncoder(w).Encode(devs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func getDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	vars := mux.Vars(r)

	dev, exists := Devices.Get(vars["url"])
	if dev == nil || !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(&dev); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// get put per sensori ed attuatori
