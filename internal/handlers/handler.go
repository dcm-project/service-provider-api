// Package handlers implements HTTP handlers for the Provider API.
package handlers

import (
	"context"

	"github.com/dcm-project/service-provider-manager/internal/api/server"
	"github.com/dcm-project/service-provider-manager/internal/logging"
	"github.com/dcm-project/service-provider-manager/internal/service"
)

// Handler implements the generated StrictServerInterface for the Provider API.
type Handler struct {
	providerService *service.ProviderService
}

// NewHandler creates a new Handler with the given provider service.
func NewHandler(providerService *service.ProviderService) *Handler {
	return &Handler{providerService: providerService}
}

// Ensure Handler implements StrictServerInterface
var _ server.StrictServerInterface = (*Handler)(nil)

func (h *Handler) GetHealth(_ context.Context, _ server.GetHealthRequestObject) (server.GetHealthResponseObject, error) {
	status := "ok"
	path := "health"
	return server.GetHealth200JSONResponse{Status: &status, Path: &path}, nil
}

func (h *Handler) ListProviders(ctx context.Context, request server.ListProvidersRequestObject) (server.ListProvidersResponseObject, error) {
	log := logging.FromContext(ctx)
	log.Debug("ListProviders request received",
		"type", request.Params.Type,
		"page_size", request.Params.MaxPageSize,
	)

	var serviceType string
	var maxPageSize int
	var pageToken string

	if request.Params.Type != nil {
		serviceType = *request.Params.Type
	}
	if request.Params.MaxPageSize != nil {
		maxPageSize = *request.Params.MaxPageSize
	}
	if request.Params.PageToken != nil {
		pageToken = *request.Params.PageToken
	}

	result, err := h.providerService.ListProviders(ctx, serviceType, maxPageSize, pageToken)
	if err != nil {
		logServiceError(ctx, "ListProviders failed", err)
		if svcErr, ok := err.(*service.ServiceError); ok && svcErr.Code == service.ErrCodeValidation {
			return server.ListProviders400ApplicationProblemPlusJSONResponse(newError("validation-error", "Invalid request", svcErr.Message, 400)), nil
		}
		return server.ListProviders400ApplicationProblemPlusJSONResponse(newError("list-error", "Failed to list providers", err.Error(), 400)), nil
	}

	response := server.ListProviders200JSONResponse{Providers: &result.Providers}
	if result.NextPageToken != "" {
		response.NextPageToken = &result.NextPageToken
	}

	log.Debug("ListProviders completed", "count", len(result.Providers))
	return response, nil
}

func (h *Handler) CreateProvider(ctx context.Context, request server.CreateProviderRequestObject) (server.CreateProviderResponseObject, error) {
	log := logging.FromContext(ctx)
	log.Debug("CreateProvider request received",
		"client_id", request.Params.Id,
		"name", request.Body.Name,
	)

	response, err := h.providerService.RegisterOrUpdateProvider(ctx, request.Body, request.Params.Id)
	if err != nil {
		logServiceError(ctx, "CreateProvider failed", err)
		if svcErr, ok := err.(*service.ServiceError); ok {
			switch svcErr.Code {
			case service.ErrCodeValidation:
				return server.CreateProvider400ApplicationProblemPlusJSONResponse(newError("validation-error", "Validation failed", svcErr.Message, 400)), nil
			case service.ErrCodeConflict:
				return server.CreateProvider409ApplicationProblemPlusJSONResponse(newError("conflict", "Resource conflict", svcErr.Message, 409)), nil
			}
		}
		return server.CreateProvider400ApplicationProblemPlusJSONResponse(newError("create-error", "Failed to create provider", err.Error(), 400)), nil
	}

	if response.Status != nil && *response.Status == server.Updated {
		log.Info("Provider updated", "provider_id", *response.Id)
		return server.CreateProvider200JSONResponse(*response), nil
	}
	log.Info("Provider created", "provider_id", *response.Id)
	return server.CreateProvider201JSONResponse(*response), nil
}

func (h *Handler) GetProvider(ctx context.Context, request server.GetProviderRequestObject) (server.GetProviderResponseObject, error) {
	log := logging.FromContext(ctx)
	log.Debug("GetProvider request received", "provider_id", request.ProviderId)

	provider, err := h.providerService.GetProvider(ctx, request.ProviderId)
	if err != nil {
		logServiceError(ctx, "GetProvider failed", err, "provider_id", request.ProviderId)
		if svcErr, ok := err.(*service.ServiceError); ok && svcErr.Code == service.ErrCodeNotFound {
			return server.GetProvider404ApplicationProblemPlusJSONResponse(newError("not-found", "Provider not found", svcErr.Message, 404)), nil
		}
		return server.GetProvider400ApplicationProblemPlusJSONResponse(newError("get-error", "Failed to get provider", err.Error(), 400)), nil
	}

	log.Debug("GetProvider completed", "provider_id", request.ProviderId)
	return server.GetProvider200JSONResponse(*provider), nil
}

func (h *Handler) ApplyProvider(ctx context.Context, request server.ApplyProviderRequestObject) (server.ApplyProviderResponseObject, error) {
	log := logging.FromContext(ctx)
	log.Debug("ApplyProvider request received", "provider_id", request.ProviderId)

	provider, err := h.providerService.UpdateProvider(ctx, request.ProviderId, request.Body)
	if err != nil {
		logServiceError(ctx, "ApplyProvider failed", err, "provider_id", request.ProviderId)
		if svcErr, ok := err.(*service.ServiceError); ok {
			switch svcErr.Code {
			case service.ErrCodeNotFound:
				return server.ApplyProvider404ApplicationProblemPlusJSONResponse(newError("not-found", "Provider not found", svcErr.Message, 404)), nil
			case service.ErrCodeConflict:
				return server.ApplyProvider409ApplicationProblemPlusJSONResponse(newError("conflict", "Name conflict", svcErr.Message, 409)), nil
			}
		}
		return server.ApplyProvider400ApplicationProblemPlusJSONResponse(newError("update-error", "Failed to update provider", err.Error(), 400)), nil
	}

	log.Info("Provider updated", "provider_id", request.ProviderId)
	return server.ApplyProvider200JSONResponse(*provider), nil
}

func (h *Handler) DeleteProvider(ctx context.Context, request server.DeleteProviderRequestObject) (server.DeleteProviderResponseObject, error) {
	log := logging.FromContext(ctx)
	log.Debug("DeleteProvider request received", "provider_id", request.ProviderId)

	err := h.providerService.DeleteProvider(ctx, request.ProviderId)
	if err != nil {
		logServiceError(ctx, "DeleteProvider failed", err, "provider_id", request.ProviderId)
		if svcErr, ok := err.(*service.ServiceError); ok && svcErr.Code == service.ErrCodeNotFound {
			return server.DeleteProvider404ApplicationProblemPlusJSONResponse(newError("not-found", "Provider not found", svcErr.Message, 404)), nil
		}
		return server.DeleteProvider400ApplicationProblemPlusJSONResponse(newError("delete-error", "Failed to delete provider", err.Error(), 400)), nil
	}

	log.Info("Provider deleted", "provider_id", request.ProviderId)
	return server.DeleteProvider204Response{}, nil
}

// logServiceError logs at Warn level for client errors (4xx) and Error level
// for internal failures (5xx), so log severity matches HTTP response semantics.
func logServiceError(ctx context.Context, msg string, err error, attrs ...any) {
	log := logging.FromContext(ctx)
	args := append([]any{"error", err}, attrs...)
	var svcErr *service.ServiceError
	if service.IsClientError(err, &svcErr) {
		log.Warn(msg, args...)
	} else {
		log.Error(msg, args...)
	}
}

func newError(errType, title, detail string, status int) server.Error {
	return server.Error{
		Type:   errType,
		Title:  title,
		Detail: &detail,
		Status: &status,
	}
}
