//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	providerclient "github.com/dcm-project/service-provider-manager/pkg/client/provider"
	. "github.com/onsi/gomega"
)

func wireMockURL() string {
	if url := os.Getenv("WIREMOCK_URL"); url != "" {
		return url
	}
	return "http://localhost:9090"
}

// providerEndpoint returns the WireMock URL as seen from inside the container network.
func providerEndpoint() string {
	if url := os.Getenv("PROVIDER_ENDPOINT"); url != "" {
		return url
	}
	return "http://provider-wiremock:8080"
}

func resetWireMock() {
	req, _ := http.NewRequest(http.MethodDelete, wireMockURL()+"/__admin/mappings", nil)
	http.DefaultClient.Do(req)
}

func stubProviderHealthEndpoint() {
	stub := map[string]interface{}{
		"request": map[string]interface{}{
			"method":  "GET",
			"urlPath": "/health",
		},
		"response": map[string]interface{}{
			"status": 200,
			"headers": map[string]string{
				"Content-Type": "application/json",
			},
			"jsonBody": map[string]interface{}{
				"status": "ok",
			},
		},
	}

	body, _ := json.Marshal(stub)
	http.Post(wireMockURL()+"/__admin/mappings", "application/json", bytes.NewReader(body))
}

// waitForProviderReady polls the provider API until the provider's health status is "ready".
// This ensures the health check monitor has confirmed the provider before tests proceed.
func waitForProviderReady(apiClient *providerclient.ClientWithResponses, ctx context.Context, providerID string) {
	Eventually(func() string {
		getResp, err := apiClient.GetProviderWithResponse(ctx, providerID)
		if err != nil || getResp.StatusCode() != http.StatusOK || getResp.JSON200 == nil {
			return ""
		}
		if getResp.JSON200.HealthStatus == nil {
			return ""
		}
		return *getResp.JSON200.HealthStatus
	}, 30*time.Second, 1*time.Second).Should(Equal("ready"), "Provider should become ready")
}

func stubProviderCreateInstance() {
	stub := map[string]interface{}{
		"request": map[string]interface{}{
			"method":  "POST",
			"urlPath": "/",
		},
		"response": map[string]interface{}{
			"status": 200,
			"headers": map[string]string{
				"Content-Type": "application/json",
			},
			"transformers": []string{"response-template"},
			"jsonBody": map[string]interface{}{
				"id":     "{{request.query.id}}",
				"status": "PROVISIONING",
			},
		},
	}

	body, _ := json.Marshal(stub)
	http.Post(wireMockURL()+"/__admin/mappings", "application/json", bytes.NewReader(body))
}

func stubProviderDeleteInstance() {
	stub := map[string]interface{}{
		"request": map[string]interface{}{
			"method":         "DELETE",
			"urlPathPattern": "/.*",
		},
		"response": map[string]interface{}{
			"status": 204,
		},
	}

	body, _ := json.Marshal(stub)
	http.Post(wireMockURL()+"/__admin/mappings", "application/json", bytes.NewReader(body))
}

func stubProviderDeleteInstanceFailure() {
	stub := map[string]interface{}{
		"request": map[string]interface{}{
			"method":         "DELETE",
			"urlPathPattern": "/.*",
		},
		"response": map[string]interface{}{
			"status": 500,
			"headers": map[string]string{
				"Content-Type": "application/json",
			},
			"jsonBody": map[string]interface{}{
				"error": "internal server error",
			},
		},
	}

	body, _ := json.Marshal(stub)
	http.Post(wireMockURL()+"/__admin/mappings", "application/json", bytes.NewReader(body))
}

func clearDeleteStubAndStubFailure() {
	// Remove all existing mappings for DELETE and re-stub with failure
	resetDeleteStubs()
	stubProviderDeleteInstanceFailure()
}

func resetDeleteStubs() {
	// Get all mappings and remove DELETE ones
	resp, err := http.Get(wireMockURL() + "/__admin/mappings")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result struct {
		Mappings []struct {
			ID      string `json:"id"`
			Request struct {
				Method string `json:"method"`
			} `json:"request"`
		} `json:"mappings"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	for _, m := range result.Mappings {
		if m.Request.Method == "DELETE" {
			req, _ := http.NewRequest(http.MethodDelete, wireMockURL()+"/__admin/mappings/"+m.ID, nil)
			http.DefaultClient.Do(req)
		}
	}
}
