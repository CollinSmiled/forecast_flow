package openmeteo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout  = 10 * time.Second
	maximumResponseSize = 2 << 20
)

type openMeteoErrorResponse struct {
	Reason string `json:"reason"`
}

func requestJSON(
	ctx context.Context,
	httpClient *http.Client,
	endpoint *url.URL,
	service string,
	target any,
) error {
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint.String(),
		nil,
	)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	request.Header.Set("Accept", "application/json")

	response, err := httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(
		response.Body,
		maximumResponseSize+1,
	))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if len(body) > maximumResponseSize {
		return errors.New("response is too large")
	}

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		return decodeOpenMeteoError(
			service,
			response.StatusCode,
			body,
		)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func decodeOpenMeteoError(
	service string,
	statusCode int,
	body []byte,
) error {
	var payload openMeteoErrorResponse

	service = strings.TrimSpace(service)
	if service == "" {
		service = "API"
	}

	if err := json.Unmarshal(body, &payload); err == nil &&
		payload.Reason != "" {
		return fmt.Errorf(
			"Open-Meteo %s returned status %d: %s",
			service,
			statusCode,
			payload.Reason,
		)
	}

	return fmt.Errorf(
		"Open-Meteo %s returned status %d",
		service,
		statusCode,
	)
}
