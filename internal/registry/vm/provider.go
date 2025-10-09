// vm_provider.go
package vm

import "context"

type Request struct {
	OsImage   string
	Ram       int
	Cpu       int
	RequestId string
	Namespace string
	VMName    string
}

type DeclaredVM struct {
	RequestInfo Request
	ID          string
}

// Provider vm_provider
type Provider interface {
	Name() string
	ProviderID() string
	Description() string
	CreateVM(ctx context.Context, request Request) (DeclaredVM, error)
	GetVM(ctx context.Context, vmID string) (DeclaredVM, error)
	DeleteVM(ctx context.Context, vmID string) (DeclaredVM, error)
	ListVMs(ctx context.Context) ([]DeclaredVM, error)
}
