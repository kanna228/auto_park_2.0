# Auto Park Modular Monolith

This archive is a refactored version of the original microservice project into a modular monolith.

## Structure

- `cmd/auto_park/main.go` — single entrypoint
- `router.go` — unified Gin router on port `8080`
- `pkg/postgres` — shared Postgres connection helper
- `internal/config` — unified config loader
- `middleware` — shared JWT and role middleware
- `user_module` — users/auth/drivers
- `vehicle_module` — vehicles/tires/tire places
- `tripsheet_module` — tripsheets/trips
- `database/migrations` — merged SQL migrations from all previous services
- `.env` — single environment file
- `air.toml` — live reload config
- `docker-compose.yaml` — postgres + migrate + monolith app

## Run locally

```bash
go mod tidy
go run ./cmd/auto_park
```

## Run with Docker

```bash
docker compose up --build
```

## Main routes

- `POST /api/users/login`
- `GET|POST|PUT|DELETE /api/users...`
- `GET|POST|PUT|DELETE /api/users/drivers...`
- `GET|POST|PUT|DELETE /api/vehicles...`
- `GET|POST|PUT|DELETE /api/vehicles/tires...`
- `GET|POST|PUT|DELETE /api/vehicles/tire-places...`
- `GET /api/vehicles/vehicle-tires/:vehicle_id`
- `GET|POST|PUT|DELETE /api/tripsheet...`
- `GET|POST|PUT|DELETE /api/tripsheet/trips...`
- `GET /healthz`
- `GET /api/tripsheet/health`

## Notes

- All tables are created in the default PostgreSQL schema (`public`).
- Old service-specific schemas were removed from migrations and SQL queries.
- Swagger UI is available at `/swagger/index.html` and loads its spec from `/swagger/doc.json`.
