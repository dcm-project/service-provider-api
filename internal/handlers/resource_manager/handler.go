// Package resource_manager implements HTTP handlers for the Resource Manager API.
package resource_manager

import (
	"context"

	server "github.com/dcm-project/service-provider-manager/internal/api/server/resource_manager"
	"github.com/dcm-project/service-provider-manager/internal/logging"
	rmsvc "github.com/dcm-project/service-provider-manager/internal/service/resource_manager"
)

// Handler implements the generated StrictServerInterface for the Resource Manager API.
type Handler struct {
	instanceService *rmsvc.InstanceService
}

// NewHandler creates a new Handler with the given instance service.
func NewHandler(instanceService *rmsvc.InstanceService) *Handler {
	return &Handler{instanceService: instanceService}
}

// Ensure Handler implements StrictServerInterface
var _ server.StrictServerInterface = (*Handler)(nil)

// GetHealth returns the health status of the service.
func (h *Handler) GetHealth(_ context.Context, _ server.GetHealthRequestObject) (server.GetHealthResponseObject, error) {
	status := "ok"
	path := "health"
	return server.GetHealth200JSONResponse{Status: &status, Path: &path}, nil
}

// ListInstances returns a paginated list of service type instances.
func (h *Handler) ListInstances(ctx context.Context, request server.ListInstancesRequestObject) (server.ListInstancesResponseObject, error) {
	log := logging.FromContext(ctx)
	log.Debug("ListInstances request received",
		"provider", request.Params.Provider,
		"page_size", request.Params.MaxPageSize,
	)

	result, err := h.instanceService.ListInstances(
		ctx,
		request.Params.Provider,
		request.Params.MaxPageSize,
		request.Params.PageToken,
	)
	if err != nil {
		logServiceError(ctx, "ListInstances failed", err)
		return handleListInstancesError(err), nil
	}

	instances := convertAPIListToServer(result.Instances)
	response := server.ListInstances200JSONResponse{
		Instances:     &instances,
		NextPageToken: result.NextPageToken,
	}

	log.Debug("ListInstances completed", "count", len(instances))
	return response, nil
}

// CreateInstance creates a new service type instance.
func (h *Handler) CreateInstance(ctx context.Context, request server.CreateInstanceRequestObject) (server.CreateInstanceResponseObject, error) {
	log := logging.FromContext(ctx)
	log.Debug("CreateInstance request received",
		"client_id", request.Params.Id,
		"provider_name", request.Body.ProviderName,
	)

	instance := convertServerToAPI(request.Body)

	result, err := h.instanceService.CreateInstance(ctx, instance, request.Params.Id)
	if err != nil {
		logServiceError(ctx, "CreateInstance failed", err)
		return handleCreateInstanceError(err), nil
	}

	log.Info("Instance created", "instance_id", *result.Id)
	return server.CreateInstance201JSONResponse(convertAPIToServer(result)), nil
}

// GetInstance retrieves a service type instance by ID.
func (h *Handler) GetInstance(ctx context.Context, request server.GetInstanceRequestObject) (server.GetInstanceResponseObject, error) {
	log := logging.FromContext(ctx)
	log.Debug("GetInstance request received", "instance_id", request.InstanceId)

	result, err := h.instanceService.GetInstance(ctx, request.InstanceId)
	if err != nil {
		logServiceError(ctx, "GetInstance failed", err, "instance_id", request.InstanceId)
		return handleGetInstanceError(err), nil
	}

	log.Debug("GetInstance completed", "instance_id", request.InstanceId)
	return server.GetInstance200JSONResponse(convertAPIToServer(result)), nil
}

// DeleteInstance deletes a service type instance by ID.
func (h *Handler) DeleteInstance(ctx context.Context, request server.DeleteInstanceRequestObject) (server.DeleteInstanceResponseObject, error) {
	log := logging.FromContext(ctx)
	log.Debug("DeleteInstance request received", "instance_id", request.InstanceId)

	err := h.instanceService.DeleteInstance(ctx, request.InstanceId)
	if err != nil {
		logServiceError(ctx, "DeleteInstance failed", err, "instance_id", request.InstanceId)
		return handleDeleteInstanceError(err), nil
	}

	log.Info("Instance deleted", "instance_id", request.InstanceId)
	return server.DeleteInstance204Response{}, nil
}
