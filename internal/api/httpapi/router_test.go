package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubReadinessChecker struct {
	err error
}

func (stub stubReadinessChecker) Ping(context.Context) error {
	return stub.err
}

func TestLiveness(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	response := httptest.NewRecorder()

	NewRouter(stubReadinessChecker{}).ServeHTTP(response, request)

	assertHealthResponse(
		t,
		response,
		http.StatusOK,
		"alive",
	)
}

func TestReadiness(t *testing.T) {
	tests := []struct {
		name           string
		databaseError  error
		expectedCode   int
		expectedStatus string
	}{
		{
			name:           "database available",
			expectedCode:   http.StatusOK,
			expectedStatus: "ready",
		},
		{
			name:           "database unavailable",
			databaseError:  errors.New("database unavailable"),
			expectedCode:   http.StatusServiceUnavailable,
			expectedStatus: "unavailable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodGet,
				"/health/ready",
				nil,
			)
			response := httptest.NewRecorder()

			NewRouter(stubReadinessChecker{
				err: test.databaseError,
			}).ServeHTTP(response, request)

			assertHealthResponse(
				t,
				response,
				test.expectedCode,
				test.expectedStatus,
			)
		})
	}
}

func assertHealthResponse(
	t *testing.T,
	response *httptest.ResponseRecorder,
	expectedCode int,
	expectedStatus string,
) {
	t.Helper()

	if response.Code != expectedCode {
		t.Fatalf(
			"status code = %d, want %d",
			response.Code,
			expectedCode,
		)
	}

	var body healthResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Service != "api" {
		t.Errorf(
			"service = %q, want %q",
			body.Service,
			"api",
		)
	}

	if body.Status != expectedStatus {
		t.Errorf(
			"status = %q, want %q",
			body.Status,
			expectedStatus,
		)
	}
}
