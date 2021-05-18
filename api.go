package moody

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type DevicesResp struct {
	Devices map[string]Device `json:"devices"`
}

func moodyApi() {
	router := mux.NewRouter()

	router.HandleFunc("/api/device", getDevices).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func getDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	devs := DevicesResp{Devices: Devices.devices}
	if err := json.NewEncoder(w).Encode(devs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// get put per sensori ed attuatori
