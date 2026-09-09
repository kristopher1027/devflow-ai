package http

import (
	"encoding/json"
	nethttp "net/http"
)

func HealthHandler(w nethttp.ResponseWriter, r *nethttp.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{
		"service": "devflow-api",
		"status":  "ok",
	}

	json.NewEncoder(w).Encode(response)
}
