package handler

import (
	"net/http"

	"github.com/otoritech/chatat/pkg/response"
)

type healthResponse struct {
	Status string `json:"status"`
}

// HealthCheck handles the health check endpoint.
func HealthCheck(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, healthResponse{Status: "ok"})
}
