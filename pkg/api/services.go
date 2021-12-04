package api

import (
	"encoding/json"
	"net/http"

	"github.com/antima/moody-core/pkg/mqtt"
)

func getServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	services := mqtt.GetActiveServices()
	if err := json.NewEncoder(w).Encode(&services); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
