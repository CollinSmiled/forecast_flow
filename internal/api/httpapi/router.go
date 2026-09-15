package httpapi

import (
	"encoding/json"
	"net/http"
)

type healthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func NewRouter() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /health/live", liveness)

	return router
}

func liveness(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(response).Encode(healthResponse{
		Service: "api",
		Status:  "alive",
	}); err != nil {
		http.Error(
			response,
			"failed to encode response",
			http.StatusInternalServerError,
		)
	}
}
