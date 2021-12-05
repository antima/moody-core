package api

import (
	"encoding/json"
	"net/http"

	"github.com/antima/moody-core/pkg/mqtt"
)

func getServices(services *mqtt.ServiceMap) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		serviceList := services.List()
		if err := json.NewEncoder(w).Encode(&serviceList); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
