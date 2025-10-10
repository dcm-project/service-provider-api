// vm_provider.go
package vm

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
type Provider struct {
	Name        string
	Endpoint    string
	ProviderID  string
	Description string
	Type        string
}
