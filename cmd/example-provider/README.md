# Example Provider

Reference implementation demonstrating the provider registration flow with DCM.

## What it does

This file storage provider shows how to:
- Register with DCM on startup for multiple resource types
- Implement resource lifecycle operations (CREATE/READ/DELETE)
- Handle graceful shutdown with automatic unregistration

## Running

**First time setup - Start PostgreSQL:**
```bash
make deploy-db
```

**Start service-provider-api server:**
```bash
make run
```

**Start provider (in another terminal):**
```bash
make run-example-provider
```

The provider automatically:
1. Registers for `file` and `container` resource types
2. Provides file storage API on `localhost:8081`
3. Unregisters when stopped with Ctrl+C

## Key Implementation Points

- **Service ID**: Same UUID for all resource type registrations
- **Endpoints**: Separate endpoints per resource type (`/api/file`, `/api/container`)
- **Registration**: Uses `pkg/registration/client` library
- **Idempotent**: Re-running updates existing registration

## API Endpoints

```bash
# Create file
curl -X POST http://localhost:8081/api/file \
  -H "Content-Type: application/json" \
  -d '{"name": "test.txt", "content": "Hello!"}'

# Get file
curl http://localhost:8081/api/file/{id}

# Delete file
curl -X DELETE http://localhost:8081/api/file/{id}
```

## Monitoring

```bash
# View all registered providers
curl http://localhost:9090/admin/registry | jq

# View service catalog  
curl http://localhost:9090/admin/catalog | jq
```

## Database Management

```bash
# Stop PostgreSQL
make kill-db

# Restart PostgreSQL
make deploy-db
```

