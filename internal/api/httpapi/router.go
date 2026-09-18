package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const readinessTimeout = 2 * time.Second

type ReadinessChecker interface {
	Ping(context.Context) error
}

type healthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func NewRouter(readinessChecker ReadinessChecker) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /health/live", liveness)
	router.HandleFunc(
		"GET /health/ready",
		readiness(readinessChecker),
	)

	return router
}

func liveness(response http.ResponseWriter, _ *http.Request) {
	writeHealthResponse(
		response,
		http.StatusOK,
		healthResponse{
			Service: "api",
			Status:  "alive",
		},
	)
}

func readiness(checker ReadinessChecker) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		ctx, cancel := context.WithTimeout(
			request.Context(),
			readinessTimeout,
		)
		defer cancel()

		if err := checker.Ping(ctx); err != nil {
			writeHealthResponse(
				response,
				http.StatusServiceUnavailable,
				healthResponse{
					Service: "api",
					Status:  "unavailable",
				},
			)

			return
		}

		writeHealthResponse(
			response,
			http.StatusOK,
			healthResponse{
				Service: "api",
				Status:  "ready",
			},
		)
	}
}

func writeHealthResponse(
	response http.ResponseWriter,
	statusCode int,
	body healthResponse,
) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(statusCode)

	if err := json.NewEncoder(response).Encode(body); err != nil {
		return
	}
}
