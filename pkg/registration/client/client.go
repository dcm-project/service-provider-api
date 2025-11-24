package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client for Resource Provider to register with DCM
type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// Config for creating a registration client
type Config struct {
	// BaseURL of the DCM Service Provider API (e.g., "https://dcm.local")
	BaseURL string

	// Timeout for registration requests
	Timeout time.Duration

	// HTTPClient custom HTTP client (optional)
	HTTPClient *http.Client
}

// New creates a new registration client
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	return &Client{
		baseURL:    cfg.BaseURL,
		httpClient: httpClient,
		timeout:    cfg.Timeout,
	}
}

// RegistrationRequest represents a registration request
type RegistrationRequest struct {
	ServiceID  string   `json:"service_id"`
	Endpoint   string   `json:"endpoint"`
	Metadata   Metadata `json:"metadata"`
	Operations []string `json:"operations"`
}

// Metadata about the provider
type Metadata struct {
	Zone                string            `json:"zone"`
	Region              string            `json:"region"`
	ResourceConstraints map[string]string `json:"resource_constraints,omitempty"`
}

// RegistrationResponse from DCM
type RegistrationResponse struct {
	ServiceID    string    `json:"service_id"`
	ResourceKind string    `json:"resource_kind"`
	Status       string    `json:"status"`
	RegisteredAt time.Time `json:"registered_at"`
	Message      string    `json:"message,omitempty"`
}

// Register sends a registration request to DCM
func (c *Client) Register(ctx context.Context, resourceKind string, req *RegistrationRequest) (*RegistrationResponse, error) {
	url := fmt.Sprintf("%s/resource/%s/provider", c.baseURL, resourceKind)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var regResp RegistrationResponse
	if err := json.Unmarshal(respBody, &regResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &regResp, nil
}

// Unregister removes the provider registration
func (c *Client) Unregister(ctx context.Context, resourceKind, providerID string) error {
	url := fmt.Sprintf("%s/resource/%s/provider/%s", c.baseURL, resourceKind, providerID)

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("unregister request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unregister failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetRegistration retrieves current registration info
func (c *Client) GetRegistration(ctx context.Context, resourceKind, serviceID string) (*RegistrationResponse, error) {
	url := fmt.Sprintf("%s/resource/%s/provider/%s", c.baseURL, resourceKind, serviceID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("get request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get failed with status %d: %s", resp.StatusCode, string(body))
	}

	var regResp RegistrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &regResp, nil
}
