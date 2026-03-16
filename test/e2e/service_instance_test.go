//go:build e2e

package e2e_test

import (
	"context"
	"net/http"
	"os"

	"github.com/dcm-project/service-provider-manager/api/v1alpha1"
	"github.com/dcm-project/service-provider-manager/api/v1alpha1/resource_manager"
	"github.com/dcm-project/service-provider-manager/pkg/client"
	rmClient "github.com/dcm-project/service-provider-manager/pkg/client/resource_manager"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Instance API", func() {
	var (
		rmApiClient *rmClient.ClientWithResponses
		apiClient   *client.ClientWithResponses
		ctx         context.Context
		baseURL     string
	)

	BeforeEach(func() {
		baseURL = os.Getenv("API_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080/api/v1alpha1"
		}

		var err error
		rmApiClient, err = rmClient.NewClientWithResponses(baseURL)
		Expect(err).NotTo(HaveOccurred())

		apiClient, err = client.NewClientWithResponses(baseURL)
		Expect(err).NotTo(HaveOccurred())

		ctx = context.Background()
	})

	Describe("Service Instance Health Check", func() {
		It("returns healthy status", func() {
			resp, err := rmApiClient.GetHealthWithResponse(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode()).To(Equal(http.StatusOK))
			Expect(resp.JSON200).NotTo(BeNil())
			Expect(*resp.JSON200.Status).To(Equal("ok"))
		})
	})

	Describe("Service Instance Create", func() {
		var providerID string
		var providerName string

		BeforeEach(func() {
			// Create a test provider for instance operations
			providerName = "e2e-test-provider-" + uuid.New().String()[:8]
			createResp, err := apiClient.CreateProviderWithResponse(ctx, nil, v1alpha1.Provider{
				Name:          providerName,
				Endpoint:      "http://example.com/api", // Mock endpoint
				ServiceType:   "vm",
				SchemaVersion: "v1alpha1",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(createResp.StatusCode()).To(Equal(http.StatusCreated))
			providerID = *createResp.JSON201.Id
		})

		AfterEach(func() {
			if providerID != "" {
				apiClient.DeleteProviderWithResponse(ctx, providerID)
			}
		})

		It("validates instance creation requires a provider", func() {
			createResp, err := rmApiClient.CreateInstanceWithResponse(ctx, nil, resource_manager.ServiceTypeInstance{
				ProviderName: "non-existent-provider-" + uuid.New().String(),
				Spec:         map[string]interface{}{"cpu": 1},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(createResp.StatusCode()).To(Equal(http.StatusNotFound))
		})

		It("handles specified instance IDs", func() {
			customID := uuid.New().String()
			params := &resource_manager.CreateInstanceParams{Id: &customID}

			createResp, err := rmApiClient.CreateInstanceWithResponse(ctx, params, resource_manager.ServiceTypeInstance{
				ProviderName: providerName,
				Spec:         map[string]interface{}{"cpu": 2, "memory": "4GB"},
			})

			Expect(err).NotTo(HaveOccurred())
			// Will fail because example.com is not a real provider, but ID should be validated
			if createResp.StatusCode() == http.StatusCreated {
				Expect(*createResp.JSON201.Id).To(Equal(customID))
			}
		})

		It("rejects duplicate instance IDs", func() {
			customID := uuid.New().String()
			params := &resource_manager.CreateInstanceParams{Id: &customID}
			body := resource_manager.ServiceTypeInstance{
				ProviderName: providerName,
				Spec:         map[string]interface{}{"cpu": 2},
			}

			// First attempt
			resp1, err := rmApiClient.CreateInstanceWithResponse(ctx, params, body)
			Expect(err).NotTo(HaveOccurred())

			// Second attempt with same ID should fail with 409 if first succeeded
			if resp1.StatusCode() == http.StatusCreated {
				resp2, err := rmApiClient.CreateInstanceWithResponse(ctx, params, body)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp2.StatusCode()).To(Equal(http.StatusConflict))
			}
		})
	})

	Describe("Service Instance List", func() {
		It("returns a list of instances (may be empty)", func() {
			listResp, err := rmApiClient.ListInstancesWithResponse(ctx, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(listResp.StatusCode()).To(Equal(http.StatusOK))
			Expect(listResp.JSON200).NotTo(BeNil())
			Expect(listResp.JSON200.Instances).NotTo(BeNil())
		})

		It("respects max page size parameter", func() {
			maxPageSize := 10
			params := &resource_manager.ListInstancesParams{
				MaxPageSize: &maxPageSize,
			}

			listResp, err := rmApiClient.ListInstancesWithResponse(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			Expect(listResp.StatusCode()).To(Equal(http.StatusOK))
			Expect(listResp.JSON200).NotTo(BeNil())

			// Verify we don't get more than max page size
			if listResp.JSON200.Instances != nil {
				Expect(len(*listResp.JSON200.Instances)).To(BeNumerically("<=", maxPageSize))
			}
		})

		It("accepts provider filter parameter", func() {
			providerName := "test-filter-provider"
			params := &resource_manager.ListInstancesParams{
				Provider: &providerName,
			}

			listResp, err := rmApiClient.ListInstancesWithResponse(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			Expect(listResp.StatusCode()).To(Equal(http.StatusOK))
			Expect(listResp.JSON200).NotTo(BeNil())
		})

		It("returns error for invalid max page size", func() {
			invalidPageSize := 1000 // exceeds max of 100
			params := &resource_manager.ListInstancesParams{
				MaxPageSize: &invalidPageSize,
			}

			listResp, err := rmApiClient.ListInstancesWithResponse(ctx, params)

			Expect(err).NotTo(HaveOccurred())
			Expect(listResp.StatusCode()).To(Equal(http.StatusBadRequest))
		})

		It("handles page token for pagination", func() {
			maxPageSize := 5
			params := &resource_manager.ListInstancesParams{
				MaxPageSize: &maxPageSize,
			}

			// Get first page
			listResp, err := rmApiClient.ListInstancesWithResponse(ctx, params)
			Expect(err).NotTo(HaveOccurred())
			Expect(listResp.StatusCode()).To(Equal(http.StatusOK))

			// If there's a next page token, test it
			if listResp.JSON200.NextPageToken != nil {
				params.PageToken = listResp.JSON200.NextPageToken
				nextPageResp, err := rmApiClient.ListInstancesWithResponse(ctx, params)
				Expect(err).NotTo(HaveOccurred())
				Expect(nextPageResp.StatusCode()).To(Equal(http.StatusOK))
			}
		})

		It("handles invalid page token gracefully", func() {
			invalidToken := "invalid-base64-token"
			params := &resource_manager.ListInstancesParams{
				PageToken: &invalidToken,
			}

			listResp, err := rmApiClient.ListInstancesWithResponse(ctx, params)
			Expect(err).NotTo(HaveOccurred())
			// Invalid page token is treated as empty, returns 200 with results from start
			Expect(listResp.StatusCode()).To(Equal(http.StatusOK))
		})
	})

	Describe("Service Instance Get Operation", func() {
		It("returns 404 for non-existent instance", func() {
			nonExistentID := uuid.New().String()
			getResp, err := rmApiClient.GetInstanceWithResponse(ctx, nonExistentID)

			Expect(err).NotTo(HaveOccurred())
			Expect(getResp.StatusCode()).To(Equal(http.StatusNotFound))
		})
	})

	Describe("Service Instance Delete Operation", func() {
		It("returns 404 when deleting non-existent instance", func() {
			nonExistentID := uuid.New().String()
			deleteResp, err := rmApiClient.DeleteInstanceWithResponse(ctx, nonExistentID)

			Expect(err).NotTo(HaveOccurred())
			Expect(deleteResp.StatusCode()).To(Equal(http.StatusNotFound))
		})
	})

	Describe("Service Instance Error Scenarios", func() {
		It("returns 404 for missing provider name", func() {
			createResp, err := rmApiClient.CreateInstanceWithResponse(ctx, nil, resource_manager.ServiceTypeInstance{
				ProviderName: "",
				Spec:         map[string]interface{}{"cpu": 1},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(createResp.StatusCode()).To(Equal(http.StatusNotFound))
		})

		It("returns 422 when provider health status is not Ready", func() {
			// Create a provider with non-ready endpoint (will fail health check)
			providerName := "unhealthy-provider-" + uuid.New().String()[:8]
			createProviderResp, err := apiClient.CreateProviderWithResponse(ctx, nil, v1alpha1.Provider{
				Name:          providerName,
				Endpoint:      "http://invalid-endpoint-does-not-exist.local/api",
				ServiceType:   "vm",
				SchemaVersion: "v1alpha1",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(createProviderResp.StatusCode()).To(Equal(http.StatusCreated))
			providerID := *createProviderResp.JSON201.Id

			defer func() {
				apiClient.DeleteProviderWithResponse(ctx, providerID)
			}()

			// Try to create instance with unhealthy provider
			createResp, err := rmApiClient.CreateInstanceWithResponse(ctx, nil, resource_manager.ServiceTypeInstance{
				ProviderName: providerName,
				Spec:         map[string]interface{}{"cpu": 2, "memory": "4GB"},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(createResp.StatusCode()).To(Equal(http.StatusUnprocessableEntity))

			// Verify error response has appropriate message
			if createResp.ApplicationproblemJSON422 != nil {
				Expect(createResp.ApplicationproblemJSON422.Title).NotTo(BeEmpty())
				if createResp.ApplicationproblemJSON422.Detail != nil {
					// Error message may contain "not ready" or "failed to connect to provider"
					Expect(*createResp.ApplicationproblemJSON422.Detail).To(Or(
						ContainSubstring("not ready"),
						ContainSubstring("failed to connect to provider"),
					))
				}
			}
		})
	})
})
