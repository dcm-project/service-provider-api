package provider

import (
	"encoding/json"

	providerserver "github.com/dcm-project/service-provider-manager/internal/api/server/provider"
	"github.com/dcm-project/service-provider-manager/internal/service"
	"github.com/dcm-project/service-provider-manager/internal/store/model"
	"gorm.io/datatypes"
)

// ModelToProvider converts a database model to an API response type.
func ModelToProvider(m *model.Provider) *providerserver.Provider {
	p := &providerserver.Provider{
		Id:            &m.ID,
		Name:          m.Name,
		ServiceType:   m.ServiceType,
		SchemaVersion: m.SchemaVersion,
		Endpoint:      m.Endpoint,
		DisplayName:   m.DisplayName,
		HealthStatus:  m.HealthStatus.StringPtr(),
		CreateTime:    service.PtrTime(m.CreateTime),
		UpdateTime:    service.PtrTime(m.UpdateTime),
	}
	if len(m.MetadataJSON) > 0 {
		var meta providerserver.ProviderMetadata
		if err := json.Unmarshal(m.MetadataJSON, &meta); err == nil {
			p.Metadata = &meta
		}
	}
	if len(m.OperationsJSON) > 0 {
		var ops []string
		if err := json.Unmarshal(m.OperationsJSON, &ops); err == nil {
			p.Operations = &ops
		}
	}
	return p
}

// ModelToProviderWithStatus converts a database model to an API response with status.
func ModelToProviderWithStatus(m *model.Provider, status providerserver.ProviderStatus) *providerserver.Provider {
	p := ModelToProvider(m)
	p.Status = &status
	return p
}

// ProviderToModel converts an API request to a database model.
func ProviderToModel(req *providerserver.Provider, id string) model.Provider {
	m := model.Provider{
		ID:            id,
		Name:          req.Name,
		ServiceType:   req.ServiceType,
		SchemaVersion: req.SchemaVersion,
		Endpoint:      req.Endpoint,
		DisplayName:   req.DisplayName,
	}
	if req.Metadata != nil {
		if b, err := json.Marshal(req.Metadata); err == nil {
			m.MetadataJSON = datatypes.JSON(b)
		}
	}
	if req.Operations != nil {
		if b, err := json.Marshal(*req.Operations); err == nil {
			m.OperationsJSON = datatypes.JSON(b)
		}
	}
	return m
}

func applyProviderRequestToModel(dest *model.Provider, req *providerserver.Provider) {
	dest.Name = req.Name
	dest.ServiceType = req.ServiceType
	dest.SchemaVersion = req.SchemaVersion
	dest.Endpoint = req.Endpoint
	dest.DisplayName = req.DisplayName
	dest.MetadataJSON = nil
	if req.Metadata != nil {
		if b, err := json.Marshal(req.Metadata); err == nil {
			dest.MetadataJSON = datatypes.JSON(b)
		}
	}
	dest.OperationsJSON = nil
	if req.Operations != nil {
		if b, err := json.Marshal(*req.Operations); err == nil {
			dest.OperationsJSON = datatypes.JSON(b)
		}
	}
}
