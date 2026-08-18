# {{ project_title }}

A Go REST service using the chi router.

## API

Full CRUD over `{{ EntityName }}` (`{ id, displayName }`) at the p6m standard surface:

| Method | Path | Status |
|---|---|---|
| POST | `/api/v1/{{ entity-name }}s` | 201 |
| GET | `/api/v1/{{ entity-name }}s` | 200 |
| GET | `/api/v1/{{ entity-name }}s/{id}` | 200 / 404 |
| PUT | `/api/v1/{{ entity-name }}s/{id}` | 200 / 404 |
| DELETE | `/api/v1/{{ entity-name }}s/{id}` | 204 / 404 |

Health (`/health/readiness`, `/health/liveness`) and Prometheus `/metrics` answer on the
management port.

## Environment

The platform injects: `SERVER_PORT`, `MANAGEMENT_PORT`, `LOGGING_STRUCTURED`,
`DB_HOST`/`DB_PORT`/`DB_USERNAME`/`DB_PASSWORD`/`DB_DBNAME` (when persistence is selected),
`OTEL_SERVICE_NAME`, and `OTEL_EXPORTER_OTLP_ENDPOINT` (traces export iff set).

## Setup

```bash
make setup
```

## Development

```bash
go run ./cmd/server
```

## Build

```bash
make build
```

The production image builds from a clean checkout — no committed `go.sum` is required
(`go mod tidy` resolves it inside the builder stage; commit one via `make setup` to pin).

## Test

```bash
make test
```
