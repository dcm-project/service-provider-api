package registration

// DefaultCatalogDefinitions returns the predefined catalog items that should be available
// These are defined by administrators and exist before any providers register
var DefaultCatalogDefinitions = []CatalogDefinition{
	{
		Name:         "file",
		DisplayName:  "File Storage",
		Description:  "Basic file storage service",
		ResourceKind: "file",
	},
	{
		Name:         "vm",
		DisplayName:  "Virtual Machine",
		Description:  "Standard virtual machine",
		ResourceKind: "vm",
	},
	{
		Name:         "container",
		DisplayName:  "Container",
		Description:  "Container runtime service",
		ResourceKind: "container",
	},
	{
		Name:         "postgresql",
		DisplayName:  "PostgreSQL Database",
		Description:  "PostgreSQL database service",
		ResourceKind: "postgresql",
	},
}

// CatalogDefinition represents a catalog item definition
type CatalogDefinition struct {
	Name         string
	DisplayName  string
	Description  string
	ResourceKind string
}
