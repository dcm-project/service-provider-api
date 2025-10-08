package registry

import (
	"errors"

	"github.com/dcm-project/service-provider-api/internal/registry/vm"
)

type ProviderRegistry struct {
	providers map[string]vm.Provider
}

func NewRegistry() *ProviderRegistry {
	return &ProviderRegistry{providers: map[string]vm.Provider{}}
}

func (r *ProviderRegistry) RegisterProvider(p vm.Provider) {
	r.providers[p.ProviderID()] = p
}

func (r *ProviderRegistry) GetProvider(providerID string) (vm.Provider, error) {
	p, ok := r.providers[providerID]
	if !ok {
		return nil, errors.New("provider not found")
	}
	return p, nil
}
