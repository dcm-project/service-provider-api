# Deferred Deletion for Rehydration Flow

**Commit:** 01f743c
**Date:** 2026-03-27
**Context:** [Rehydration Flow Enhancement](https://github.com/dcm-project/enhancements/blob/1f357c1213ccfbb8638f9b5baed82ada86114c15/enhancements/rehydration-flow/rehydration-flow.md)

## Problem

During the rehydration flow, after a new resource is created with a new InstanceID, the old resource must be deleted. If the Service Provider is unavailable or returns an error, the deletion must not block the rehydration flow. The SP Resource Manager must return success and handle the failed deletion asynchronously.

## Solution

### Soft-Delete on Existing Table

Instead of a separate cleanup queue table, soft-delete columns were added to `ServiceTypeInstance`:

| Column | Type | Purpose |
|--------|------|---------|
| `deletion_status` | `*string` | `NULL` (active), `PENDING`, or `FAILED` |
| `retry_count` | `int` | Number of cleanup attempts |
| `last_deletion_attempt` | `*time.Time` | Timestamp of last retry |
| `deletion_requested_at` | `*time.Time` | When deferred deletion was first requested |

### API Changes

**DELETE `/service-type-instances/{id}?deferred=true`**
- `deferred=false` (default): existing behavior — returns error on SP failure
- `deferred=true`: marks instance for background cleanup on SP failure, returns 204

**LIST/GET `show_deleted=true`** (aligned to [AEP-164](https://aep.dev/164/))
- By default, LIST excludes soft-deleted instances and GET returns 404 for them
- `show_deleted=true` on LIST includes soft-deleted instances alongside active ones
- `show_deleted=true` on GET returns the soft-deleted instance instead of 404

**`deletion_status` field** added to `ServiceTypeInstance` response schema (read-only, enum: `PENDING`, `FAILED`)

### Delete Behavior Matrix

| Instance State | SP Deletion | deferred | Result |
|---------------|-------------|----------|--------|
| Active | Succeeds | any | Hard-delete, 204 |
| Active | Fails | `false` | Error returned |
| Active | Fails | `true` | Mark PENDING, 204 |
| PENDING/FAILED | Succeeds | any | Hard-delete, 204 |
| PENDING/FAILED | Fails | `false` | Reset retry count, error returned |
| PENDING/FAILED | Fails | `true` | Reset retry count, 204 |

### Background Cleanup Scheduler

`internal/cleanup/Scheduler` follows the `healthcheck.Monitor` pattern:

- Periodically queries for `deletion_status=PENDING` instances
- For each: calls `InstanceService.DeleteFromProvider` (shared with API delete flow)
- Success (2xx or 404) -> hard-deletes the DB record
- Failure -> increments retry count; marks `FAILED` after max retries

The scheduler depends on `InstanceService` rather than maintaining its own HTTP client and provider lookup, sharing `DeleteFromProvider` with the API delete handler.

**Configuration (env vars):**
- `CLEANUP_INTERVAL` — scheduler tick interval (default: `1m`)
- `CLEANUP_MAX_RETRIES` — max attempts before marking FAILED (default: `10`)
- `CLEANUP_TIMEOUT` — HTTP timeout for cleanup requests (default: `10s`)

## Files Changed

| File | Change |
|------|--------|
| `internal/store/model/service_type_instance.go` | Added soft-delete columns |
| `internal/store/resource_manager/service_instance.go` | Added `MarkForDeletion`, `ListPendingDeletions`, `IncrementDeletionRetry`, `MarkDeletionFailed`, `HardDelete` (with retry), `ResetRetryCount`; removed unused `Delete`; `Get` accepts `showDeleted bool`; `List` uses `ShowDeleted bool` instead of `DeletionStatus` filter |
| `api/v1alpha1/resource_manager/openapi.yaml` | Added `deferred` param on DELETE, `show_deleted` bool on LIST and GET (AEP-164), `deletion_status` field on schema |
| `internal/service/resource_manager/service_type_instance.go` | `DeleteInstance` with deferred mode; `DeleteFromProvider` shared by API and scheduler; `GetInstance` accepts `showDeleted`; `ListInstances` accepts `showDeleted` |
| `internal/service/resource_manager/convert.go` | Added `DeletionStatus` to model-to-API conversion |
| `internal/handlers/resource_manager/handler.go` | Passes `showDeleted` and `deferred` params to service |
| `internal/handlers/resource_manager/convert.go` | Added `DeletionStatus` to API-to-server conversion |
| `internal/cleanup/scheduler.go` | Background cleanup scheduler, uses `InstanceService.DeleteFromProvider` |
| `internal/config/config.go` | Added `CleanupConfig` |
| `cmd/service-provider-manager/main.go` | Wired cleanup scheduler with `InstanceService` dependency |
| `internal/service/resource_manager/service_type_instance_test.go` | Deferred deletion tests, updated for `showDeleted` API |
| `internal/store/resource_manager/service_instance_test.go` | Added tests for all soft-delete store methods and `showDeleted` filtering |
| `test/e2e/service_instance_test.go` | E2E tests for deferred deletion, `show_deleted` on GET/LIST |
| `test/e2e/setup_test.go` | Added WireMock helpers for SP delete failure stubbing |
| Generated files (`*.gen.go`, `go.mod`, `go.sum`) | Regenerated with updated OpenAPI spec and oapi-codegen |

## Test Coverage

### Service layer (`service_type_instance_test.go`)
- Deferred delete with SP failure -> instance marked PENDING, returns 204
- Non-deferred delete with SP failure -> returns error (regression)
- Deferred delete when provider returns error -> marked PENDING, returns 204
- DELETE on FAILED instance -> resets retry count
- DELETE on PENDING instance with SP available -> hard-deletes, returns 204
- LIST excludes soft-deleted instances by default
- LIST with `show_deleted=true` includes soft-deleted instances

### Store layer (`service_instance_test.go`)
- MarkForDeletion sets PENDING status and hides from default Get
- ListPendingDeletions returns only PENDING, excludes FAILED and active
- IncrementDeletionRetry increments count and sets last_deletion_attempt
- MarkDeletionFailed sets FAILED status
- ResetRetryCount resets to PENDING with zero count
- Get with showDeleted=true returns soft-deleted instances
- Get with showDeleted=false returns 404 for soft-deleted instances
- List with ShowDeleted=true includes soft-deleted instances
- List without ShowDeleted excludes soft-deleted instances

### E2E layer (`test/e2e/service_instance_test.go`)
- Non-deferred delete with SP failure -> returns error, instance stays active
- Deferred delete with SP failure -> returns 204, instance marked PENDING, hidden from default GET
- Deferred delete with SP success -> hard-deletes, instance fully removed
- Default LIST excludes soft-deleted instances
- LIST with `show_deleted=true` includes soft-deleted instances
- Re-delete soft-deleted instance when SP recovers -> hard-deletes
