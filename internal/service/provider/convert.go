package provider

import (
	"time"

	providerserver "github.com/dcm-project/service-provider-manager/internal/api/server/provider"
	"github.com/dcm-project/service-provider-manager/internal/service"
	"github.com/dcm-project/service-provider-manager/internal/store/model"
)

// ModelToProvider converts a database model to an API response type.
func ModelToProvider(m *model.Provider) *providerserver.Provider {
	return &providerserver.Provider{
		Id:            &m.ID,
		Name:          m.Name,
		ServiceType:   m.ServiceType,
		SchemaVersion: m.SchemaVersion,
		Endpoint:      m.Endpoint,
		HealthStatus:  m.HealthStatus.StringPtr(),
		CreateTime:    service.PtrTime(m.CreateTime),
		UpdateTime:    service.PtrTime(m.UpdateTime),
	}
}

// ModelToProviderWithStatus converts a database model to an API response with status.
func ModelToProviderWithStatus(m *model.Provider, status providerserver.ProviderStatus) *providerserver.Provider {
	p := ModelToProvider(m)
	p.Status = &status
	return p
}

// ProviderToModel converts an API request to a database model.
func ProviderToModel(req *providerserver.Provider, id string) model.Provider {
	now := time.Now()
	return model.Provider{
		ID:            id,
		Name:          req.Name,
		ServiceType:   req.ServiceType,
		SchemaVersion: req.SchemaVersion,
		Endpoint:      req.Endpoint,
		CreateTime:    now,
		UpdateTime:    now,
	}
}
