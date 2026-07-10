# Switchyard Go Backend

Go backend for the Switchyard dispatch and logistics platform.

## Prerequisites

- Go 1.25+
- PostgreSQL 16 (via Docker recommended — see below)
- [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI for schema migrations

## First-time setup

```bash
go mod tidy
```

> **Note — viper/gotenv incompatibility:** `github.com/subosito/gotenv` v1.6.0 reorganized
> its package structure and breaks viper's internal dotenv encoder. `go.mod` pins gotenv to
> v1.4.2 as an indirect dependency. If you ever see the error
> `module gotenv@latest found but does not contain package github.com/subosito/gotenv`
> after updating dependencies, run `go get github.com/subosito/gotenv@v1.4.2` before
> re-running `go mod tidy`.

## Database

### Start PostgreSQL

Port 5433 avoids conflicting with other local Postgres instances running on the default port.

```bash
docker run -d \
  --name switchyard-pg \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=switchyard \
  -p 5433:5432 \
  postgres:16
```

Set `DATABASE_URL` in your `.env`:

```
DATABASE_URL=postgres://postgres:password@localhost:5433/switchyard?sslmode=disable
```

### Read replica (CQRS, disaster recovery)

Read queries (currently `internal/repository/postgres/analytics.go`) hit a separate, read-only Postgres instance kept current via native logical replication — not application code. This is scoped for DR against a *smaller* disaster (a bad write, a botched migration, single-table corruption on the primary), not full instance/region loss, so a logically independent replica is the right tool, not full physical replication.

**Must be a genuinely separate Postgres instance/container** — a publisher and subscriber on the same instance deadlock reliably on `CREATE SUBSCRIPTION` (the subscription's own transaction blocks its own replication-slot worker). `docker-compose.yml` runs it as `switchyard-pg-read`, a distinct container with its own volume, alongside `switchyard-pg`.

Setup, run once against a fresh replica:

```bash
# 1. Bring up the replica container
docker compose up -d switchyard-pg-read

# 2. Apply Go's schema migrations to the replica's database
migrate -path internal/migrations -database "postgres://postgres:password@localhost:5434/switchyard_read?sslmode=disable" up

# 3. If any migrations seed data that also exists on the primary (e.g. seeded
#    lookup tables), truncate those tables on the replica first — the
#    subscription's initial sync copies rows from the primary and will
#    collide on primary keys otherwise.
docker exec switchyard-pg-read psql -U postgres -d switchyard_read -c "TRUNCATE hos_limit, warehouse CASCADE;"

# 4. Publish on the primary (switchyard-pg)
docker exec switchyard-pg psql -U postgres -d switchyard -c "CREATE PUBLICATION switchyard_go_pub FOR ALL TABLES;"

# 5. Subscribe from the replica, connecting cross-instance over the docker network
docker exec switchyard-pg-read psql -U postgres -d switchyard_read -c "CREATE SUBSCRIPTION switchyard_go_sub CONNECTION 'host=switchyard-pg port=5432 dbname=switchyard user=postgres password=password' PUBLICATION switchyard_go_pub;"
```

Set `READ_DATABASE_URL` in your `.env`:

```
READ_DATABASE_URL=postgres://postgres:password@localhost:5434/switchyard_read?sslmode=disable
```

### Run migrations

Run from the `Switchyard-Go/` directory — the relative path `file://internal/migrations` resolves from there.

```bash
migrate -path internal/migrations -database "$DATABASE_URL" up
```

On PowerShell, reference the env var explicitly:

```powershell
migrate -path internal/migrations -database $env:DATABASE_URL up
```

### Seed dev data (optional)

Loads a Monday-morning demo board state: 10 drivers, 6 trucks, 7 trailers across WH001–WH005, plus a handful of BOLs across the pipeline. All HOS windows are fresh. v1.4 additions: one BOL near the dead-head cutoff window (pairing-at-risk warn border) and one completed run with no pairing secured (Empty Return + Delivered close-out demo).

```bash
psql "$DATABASE_URL" -f internal/migrations/seed_dev_data.sql
```

## Build & Run

Viper reads environment variables from the shell — it does **not** auto-load `.env` files. Load them first.

**Bash / Git Bash:**
```bash
set -a && source .env && set +a
go run ./cmd/main.go
```

**PowerShell:**
```powershell
Get-Content .env | Where-Object { $_ -notmatch '^\s*#' -and $_ -match '=' } | ForEach-Object { $k,$v = $_ -split '=',2; Set-Item "Env:$($k.Trim())" $v.Trim() }
go run ./cmd/main.go
```

Other commands (no env vars needed):

```bash
go build ./...          # compile all packages
go test ./...           # run all tests
```

## Environment

Copy `.env.example` to `.env` and fill in values before running locally.

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | — | PostgreSQL connection string (required) |
| `READ_DATABASE_URL` | — | Read-replica connection string, kept current via native logical replication (required, see [Read replica](#read-replica-cqrs-disaster-recovery)) |
| `LOGISTICS_BASE_URL` | — | Switchyard .NET Logistics API base URL |
| `INVENTORY_BASE_URL` | — | Switchyard .NET Inventory API base URL |
| `AUTH0_DOMAIN` | — | Auth0 tenant domain |
| `AUTH0_CLIENT_ID` | — | M2M client ID (event handler only) |
| `AUTH0_CLIENT_SECRET` | — | M2M client secret |
| `AUTH0_AUDIENCE` | — | Auth0 API audience identifier |
| `SMTP_HOST` | — | Outbound SMTP host for notifications |
| `SMTP_PORT` | `587` | SMTP port |
| `SMTP_USER` | — | SMTP login user |
| `SMTP_PASS` | — | SMTP password |
| `DISPATCH_EMAIL` | — | Recipient address for dispatch alerts |
| `HOS_WARNING_THRESHOLD_HOURS` | `2.0` | Hours remaining before HOS warning fires |
| `DEADHEAD_WINDOW_HOURS` | `4.0` | Minimum hours between run fulfillment and dead-head pairing |
| `DEADHEAD_CUTOFF_MINUTES` | `30` | Hard pre-arrival deadline — a pairing must be secured this many minutes before the driver's estimated fulfillment, checked both at pairing creation and on the board (pairing-at-risk warn border) |
| `DEADHEAD_SEARCH_WINDOW_HOURS` | `2.0` | Look-ahead window when finding eligible dead-head pairings; also how long an Empty Return driver's pairing window stays open before it's flagged as expiring |
| `LOADING_AGE_THRESHOLD_HOURS` | `4.0` | Hours before a BOL stuck in `loading` triggers a dispatch alert |
| `DEFAULT_CYCLE_LABEL` | `60h/7d` | HOS cycle used when none is specified on a plan |
| `EMPTY_RETURN_TRANSIT_HOURS` | `2.0` | Fixed transit-time estimate for a driver/equipment on empty return with no dead-head pairing secured (v1.4 — see `WhiteboardService.emptyReturnETA` for the future distance-based upgrade path) |
| `WAREHOUSE_IDS` | `WH001,...,WH009` | Comma-separated list of warehouse IDs the regional inventory endpoint fans out to. Add a new warehouse here whenever a warehouse is added to the network. *(v1.2: this flat list will be replaced by a `WAREHOUSE_REGIONS` param once the `region` field is added to the Warehouse model — new warehouses will be picked up automatically via the DB rather than requiring a config change.)* |
| `PORT` | `8080` | HTTP listen port |
| `LOG_LEVEL` | `info` | Log verbosity |

## Key Constraints

These apply when modifying Go backend code. They are hard rules, not guidelines.

1. **CUD authority boundary.** `PlanBOLRecord`, `PlanBOLStop`, and `TruckInventorySnapshot` are exclusively Go's domain — Go creates, updates, and deletes them. Committed BOL stops and inventory writes are .NET's domain. The Go backend reads .NET data through the integrations adapter only; it never writes .NET entities. This boundary never crosses.

2. **M2M token lives in the event handler only.** `internal/events` holds the sole M2M token used to call the .NET system. No other package requests, caches, or refreshes a token.

3. **Integrations adapter is the only .NET caller.** `internal/integrations` is the only package permitted to call the Switchyard .NET APIs. No handler, service, or repository may import or call the .NET system directly.

4. **Business rules are rejections, not warnings.** The empty truck rule, the 4-hour dead-head window, and state-level HOS limits are enforced at the service layer and return errors. They are not advisory.

5. **Whiteboard column transitions are derived, not set.** The Kanban board columns follow from `PlanBOLStatus` values and assignment timestamps — not from dispatcher input. Read `internal/services/whiteboard_service.go` before touching board code.

## API Endpoints

All `/api/*` routes return JSON. `/` and `/driver/:id` return server-rendered HTML.

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/events` | Receive workflow event, route to service |
| `POST` | `/api/plan-bol` | Create BOL from invoice (status: `draft`) |
| `GET` | `/api/plan-bol/:id` | Get PlanBOLRecord with full stop sequence |
| `POST` | `/api/plan-bol/:id/begin-planning` | Claim for route planning (`draft → plan-progress`) |
| `POST` | `/api/plan-bol/:id/validate` | Run constraint solver, return violations |
| `POST` | `/api/plan-bol/:id/commit` | Commit plan, call .NET CreateBOL (`plan-progress → loading`) |
| `POST` | `/api/plan-bol/:id/mark-loaded` | Dock confirms trailer loaded (`loading → validated`) |
| `GET` | `/api/plan-bol/:id/truck-state` | Truck inventory snapshot at each stop |
| `GET` | `/api/plan-bol/:id/history` | Status transition audit trail |
| `GET` | `/api/driver` | All drivers with current HOS state |
| `GET` | `/api/driver/:id/runsheet` | Current run — stops + live inventory state |
| `POST` | `/api/driver/:id/stop/:stopId/log` | Log stop completion, notify dispatch |
| `GET` | `/api/driver/:id/active-bol` | Current active PlanBOLRecord for this driver |
| `GET` | `/api/driver/:id/hos` | Current HOS window state |
| `POST` | `/api/assignment` | Create driver–BOL–equipment assignment |
| `GET` | `/api/assignment/:id` | Get assignment with all linked entities |
| `PATCH` | `/api/assignment/:id/depart` | Mark departed (moves card to In Transit) |
| `PATCH` | `/api/assignment/:id/fulfill` | Mark fulfilled (starts dead-head timer) |
| `PATCH` | `/api/assignment/:id/deadhead` | Confirm dead-head return (clears from board) |
| `GET` | `/api/equipment` | All equipment with current status |
| `POST` | `/api/equipment` | Register new truck or tractor |
| `PATCH` | `/api/equipment/:id/maintenance` | Report scheduled maintenance |
| `PATCH` | `/api/equipment/:id/breakdown` | Report breakdown (depot or roadside) |
| `PATCH` | `/api/equipment/:id/resolve` | Resolve maintenance or breakdown record |
| `GET` | `/api/deadhead/eligible` | Eligible return BOLs given location + time |
| `POST` | `/api/deadhead/pair` | Pair two BOLs (enforces 4-hour rule) |
| `DELETE` | `/api/deadhead/:pairingId` | Cancel a confirmed pairing |
| `GET` | `/api/dispatch/board` | Full Kanban board state — all columns |
| `GET` | `/api/dispatch/alerts` | HOS warnings, breakdown alerts, expiring timers |
| `GET` | `/api/analytics/summary` | BOL counts by status, stop completion rate, 7/30-day throughput |
| `GET` | `/api/inventory/region` | Regional inventory fan-out; optional `?sku=` filter |
| `GET` | `/api/invoice/:id` | Get invoice record |
| `GET` | `/api/invoice/store/:storeId` | All invoices for a given store |
| `GET` | `/` | Dispatch Kanban whiteboard (HTML) |
| `GET` | `/driver/:id` | Driver run sheet (HTML) |
