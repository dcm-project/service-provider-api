# Service Provider API Client

This package provides a Go client for the Service Provider API, automatically generated from the OpenAPI specification.

## Installation

```bash
go get github.com/dcm-project/service-provider-api/pkg/client
```

## Usage

```go
import (
    "context"
    "github.com/dcm-project/service-provider-api/pkg/client"
)

func main() {
    // Create a new client
    c, err := client.NewClient("http://localhost:8080")
    if err != nil {
        panic(err)
    }

    // Use the client with responses wrapper for automatic response parsing
    clientWithResponses, err := client.NewClientWithResponses("http://localhost:8080")
    if err != nil {
        panic(err)
    }

    // List providers
    resp, err := clientWithResponses.ListProvidersWithResponse(context.Background())
    if err != nil {
        panic(err)
    }

    if resp.JSON200 != nil {
        // Handle successful response
        for _, provider := range resp.JSON200.Providers {
            println(provider.Name)
        }
    }
}
```

## Features

- Full type-safe API client generated from OpenAPI spec
- Support for all API endpoints
- Request editor functions for customizing requests (auth, headers, etc.)
- ClientWithResponses for automatic response parsing
- Customizable HTTP client

## Customization

### Using a custom HTTP client

```go
import "net/http"

httpClient := &http.Client{
    Timeout: 30 * time.Second,
}

c, err := client.NewClient(
    "http://localhost:8080",
    client.WithHTTPClient(httpClient),
)
```

### Adding authentication

```go
func authEditor(ctx context.Context, req *http.Request) error {
    req.Header.Set("Authorization", "Bearer YOUR_TOKEN")
    return nil
}

c, err := client.NewClient(
    "http://localhost:8080",
    client.WithRequestEditorFn(authEditor),
)
```

## Code Generation

This client is automatically generated using [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen).

To regenerate the client:

```bash
cd pkg/client
go generate
```

