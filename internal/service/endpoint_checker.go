package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// HTTPEndpointChecker checks if provider endpoints are reachable
type HTTPEndpointChecker struct {
	client  *resty.Client
	timeout time.Duration
}

// NewHTTPEndpointChecker creates a new endpoint checker
func NewHTTPEndpointChecker(timeout time.Duration) *HTTPEndpointChecker {
	client := resty.New().
		SetTimeout(timeout).
		SetRetryCount(2).
		SetRetryWaitTime(1 * time.Second)

	return &HTTPEndpointChecker{
		client:  client,
		timeout: timeout,
	}
}

// CheckEndpoint verifies the provider endpoint is reachable and healthy
func (c *HTTPEndpointChecker) CheckEndpoint(ctx context.Context, endpoint string) error {
	logger := zap.S().Named("endpoint_checker")
	
	logger.Infow("Checking endpoint reachability", "endpoint", endpoint)

	// Try to reach the health endpoint
	// Following common health check patterns: /health, /healthz, /ready
	healthPaths := []string{"/health", "/healthz", "/ready", "/"}
	
	for _, path := range healthPaths {
		url := endpoint + path
		
		resp, err := c.client.R().
			SetContext(ctx).
			Get(url)
		
		if err == nil && (resp.StatusCode() == http.StatusOK || resp.StatusCode() == http.StatusNoContent) {
			logger.Infow("Endpoint is reachable", "endpoint", endpoint, "path", path, "status", resp.StatusCode())
			return nil
		}
	}

	// If we get here, none of the health paths worked
	logger.Warnw("Endpoint is not reachable", "endpoint", endpoint)
	return fmt.Errorf("endpoint %s is not reachable on any standard health path", endpoint)
}

