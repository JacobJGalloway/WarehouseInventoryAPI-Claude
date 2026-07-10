<div align="center">
<img src="Switchyard.UI/src/assets/logo-full-name-light.png" />
</div>

# Switchyard 1.3

Switchyard is an inventory, driver, and equipment tracking and management system which coordinates logistics operations across a network of warehouses and stores. Inventory is tracked per location; Bills of Lading govern movement between any combination of stops — from same-day local transfers to multi-stop OTR runs with partial loads. Authenticated via Auth0.

## Solution Structure

| Project | Role | Port |
|---|---|---|
| `Switchyard.InventoryAPI` | Inventory API — Clothing, PPE, Tools | 7000 |
| `Switchyard.LogisticsAPI` | Logistics API — Bills of Lading, Stores, Warehouses, Users | 7001 |
| `Switchyard.Domain` | Shared class library — domain models for both .NET APIs | — |
| `Switchyard.UI` | React/TypeScript client UI | 5173 |
| `Switchyard-Go` | Go backend — PlanBOL, Dispatch Whiteboard, HOS, Equipment | 8080 |

**Go backend documentation:** [`Switchyard-Go/README.md`](Switchyard-Go/README.md) — setup, environment variables, key architectural constraints, and API reference.

## Prerequisites

| Requirement | Version | Notes |
|---|---|---|
| [.NET SDK](https://dotnet.microsoft.com/download) | 10.0 | InventoryAPI and LogisticsAPI |
| [Go](https://go.dev/dl/) | 1.25+ | Switchyard-Go backend |
| [Node.js](https://nodejs.org/) | 24+ | Switchyard.UI |
| [PostgreSQL](https://www.postgresql.org/download/) | 16 | All backends — Docker or local install |
| [Auth0 account](https://auth0.com/) | — | Tenant + API resource + two M2M applications |

**PostgreSQL:** Used by both the Go backend and the .NET APIs. Easiest to run via Docker (`postgres:16` image). Default dev port is **5433** — if another Postgres instance is already on 5432, use 5433 to avoid the conflict. See `Switchyard-Go/.env.example` for the full connection string format. The .NET APIs connect to separate databases (`switchyard_inventory`, `switchyard_logistics`) on the same instance.

**Go Service Initialization:** Due to how the environmental variables are read in Go, the initial setup for Docker will need to be different if the image is not up and running on a container. Subsequent restarts with the container already running are a single line restart. See the README.md under Switchyard-Go for more details.

**Auth0 M2M applications (free tier: 2 slots):** Switchyard uses both — one for the Scalar UI on the .NET APIs, one for the Go event handler. Confirm available M2M slots before setting up a new tenant. See [Auth0 Setup](#auth0-setup) for full configuration steps.

**SMTP:** Required for email notifications (HOS warnings, breakdown alerts, dead-head expiry timer). Any SMTP-accessible mail account works in dev. Notifications can be left unconfigured early on — fill in `SMTP_*` env vars when you need them.

**Go `.env` loading — known gotcha:** Viper does not auto-load `.env` files. Before running the Go backend, source your `.env` from `Switchyard-Go/` using the one-liner in `Switchyard-Go/README.md`, or set the vars directly in your shell session.

## Architecture

### Agentic AI Development

Switchyard was built with [Claude Code](https://claude.ai/code) as an active pair programming and mentoring partner throughout every sprint. The model always knew the answer — but frequently chose not to give it directly. Instead it would surface the right question, flag the constraint that hadn't been named yet, or ask what the invariant was before suggesting the fix. The design emerged from understanding, not from prescription.

That dynamic shaped the architecture as much as the implementation. The CUD authority boundary between Go and .NET, the service-layer transport agnosticism, and the CQRS read-replica pattern were all reasoned out through guided back-and-forth rather than handed down as decisions. The goal was to own every piece of it — not just ship it.

### CQRS Read Replica
Both .NET APIs maintain a read replica synced asynchronously after every write:
- Write operations target the primary PostgreSQL database
- Read operations target the read replica database (all `AsNoTracking`)
- `SaveChangesInterceptor` → `Channel<SyncJob>` → `BackgroundService` (full table resync per changed entity type)
- `GET /api/Audit` on each API reports write vs read row counts with an `InSync` flag

### Data Layer Pattern
- **Unit of Work** over repositories — services depend on `IUnitOfWork`
- **Repositories** — separate write context (CUD) and read context (queries)
- **EF Core** with PostgreSQL (Npgsql); migrations applied on startup for write context; `EnsureCreated` for read replica
- **Switchyard.Domain** — shared class library containing all entity models and interfaces; neither API project owns domain models directly

### Auth
Both APIs use Auth0 JWT bearer authentication. Permissions are claim-based:

| Permission | Used by |
|---|---|
| `read:inventory` | Inventory read endpoints |
| `read:bol` | Logistics read endpoints |
| `create:bol` | BOL creation |
| `modify:bol` | ProcessStop, ReplaceStop |
| `manage:users` | User management |

## API Endpoints

### Inventory API (`/api`) — port 7000

| Method | Path | Description |
|---|---|---|
| GET | `/Clothing` | All clothing items |
| GET | `/Clothing/{skuId}` | By SKU |
| GET | `/Clothing/location/{locationId}` | By location |
| GET | `/Clothing/filter?locationId=&skuId=` | By location + SKU |
| POST | `/Clothing` | Add item |
| PUT | `/Clothing/{skuId}` | Full update by SKU |
| PATCH | `/Clothing/item/{partitionKey}` | Patch projected/unloadedDate |
| DELETE | `/Clothing/item/{partitionKey}` | Delete item |
| _(same shape for `/PPE` and `/Tool`)_ | | |
| GET | `/Audit` | Write vs read row counts |

### Logistics API (`/api`) — port 7001

| Method | Path | Description |
|---|---|---|
| GET | `/BillOfLading` | All BOLs |
| GET | `/BillOfLading/{transactionId}` | BOL + line entries |
| GET | `/BillOfLading/{transactionId}/line-entry` | Line entries only |
| POST | `/BillOfLading` | Create BOL, persist line entries, write `.txt` to Downloads |
| POST | `/BillOfLading/{transactionId}/process/{locationId}` | Mark location stop as processed |
| POST | `/BillOfLading/{transactionId}/replace-stop` | Move unprocessed stop to a new location |
| GET | `/Store` | All stores |
| GET | `/Warehouse` | All warehouses |
| GET | `/User` | All Auth0 users |
| POST | `/User` | Create Auth0 user + assign role |
| PATCH | `/User/{userId}/deactivate` | Block user (soft deactivate) |
| GET | `/Audit` | Write vs read row counts |

## Auth0 Setup

1. Create an API resource in Auth0 and set its identifier as `Auth0:Audience`
2. Set `Auth0:Authority` to your Auth0 domain (e.g. `https://your-tenant.auth0.com/`)
3. Add permissions to the API: `read:inventory`, `read:bol`, `create:bol`, `modify:bol`, `manage:users`
4. For user management, create an M2M application and grant it the Auth0 Management API with scopes:
   `read:users`, `create:users`, `update:users`, `read:roles`, `create:role_members`
5. Set credentials in `{API Project Name}/appsettings.Development.json` (gitignored):

```json
{
  "Auth0": {
    "Authority": "https://your-tenant.auth0.com/",
    "Audience": "your-api-audience",
    "ScalarClientId": "your-m2m-client-id",
    "ScalarClientSecret": "your-m2m-client-secret"
  },
  "ConnectionStrings": {
    "InventoryWrite": "Host=localhost;Port=5433;Database=switchyard_inventory;Username=postgres;Password=password",
    "InventoryRead": "Host=localhost;Port=5433;Database=switchyard_inventory_read;Username=postgres;Password=password"
  }
}
```

## Running the System

Postgres must be up before *any* backend service starts — both .NET APIs and the Go backend connect to it on startup and will fail if it isn't running yet.

```bash
# 1. Postgres first (docker-compose.yml at project root) — required by both .NET APIs and Go
docker compose up -d

# 2. .NET APIs (run each in a separate terminal)
dotnet run --project Switchyard.InventoryAPI
dotnet run --project Switchyard.LogisticsAPI

# 3. Go backend
cd Switchyard-Go
Get-Content .env | Where-Object { $_ -notmatch '^\s*#' -and $_ -match '=' } | ForEach-Object { $k,$v = $_ -split '=',2; Set-Item "Env:$($k.Trim())" $v.Trim() }
go run ./cmd/main.go

# 4. UI
cd Switchyard.UI
npm run dev

# Unit Tests
dotnet test
go test ./...

# Coverage (.NET) — coverlet.runsettings is required: Controllers/Repositories and
# their tests share one assembly per API project rather than a separate *.Tests
# project, so coverlet's default test-assembly exclusion would zero out everything
dotnet test --collect:"XPlat Code Coverage" --settings coverlet.runsettings

# Coverage (Go)
go test -cover ./...

# API docs (Scalar UI — while API is running)
# Inventory: https://localhost:7000/scalar/v1
# Logistics: https://localhost:7001/scalar/v1
```

## Wanted Features

### v1.4 Wanted Features - Demo Stable Hardening / pilot-client ready
- [x] Empty Return board state — new Available sub-section for drivers on empty return to originating warehouse; ETA visible for pre-planning next BOL assignment
- [x] Delivered column redesign — BOL-only close-out card; driver and equipment decouple from the BOL at last stop confirmation and route independently to Empty Return / Available / Maintenance; Delivered represents dispatch review, client notification, and final paperwork before archiving
- [x] Deadhead pairing — enforce `DEADHEAD_CUTOFF_MINUTES` window at the board level; pairing must be secured before driver reaches last stop or contract is voided; driver routes to Empty Return on last stop confirmation
- [x] Rolling refresh tokens for Auth0 sessions in place of fixed-expiry client secrets — code complete (`useRefreshTokens`, no fixed-expiry logic remaining); Auth0 dashboard-side settings (Refresh Token Rotation, expiration windows) still need confirming outside of code
- [x] Color contrast audit (WCAG AA) — verify all text/bg combinations across light and dark themes
- [x] ARIA compliance audit — board columns, cards, icon-only buttons, skip-nav; verified with a live Narrator (screen reader) session, not just static review
- [x] Document dispatch board card border language in a design note — border present = status alert (danger/warn/ok variants); no border = clean/good status. Borders carry semantic weight and should not be used decoratively.
- [x] CQRS read replica hardening — separate Postgres read replica instances stood up for both .NET (`switchyard_inventory_read`, `switchyard_logistics_read`) and Go (`switchyard_read`, cross-instance logical replication)

### Backlog
- [ ] Auth0 Refresh Token Rotation — blocked by the free tier (2026-07-10); the tenant's current plan doesn't include Refresh Token Rotation as a dashboard feature. Code side is already done (`useRefreshTokens: true`, `offline_access` scope implied, no fixed-expiry secret logic remains — see `Switchyard.UI/src/main.tsx`), which already solves the "session dies mid-demo on secret expiry" problem standard (non-rotating) refresh tokens were meant to fix. Rotation itself only adds defense-in-depth against a leaked refresh token. Revisit if/when the project moves to a paid Auth0 tier (or another provider) for any reason — validate whether Rotation is still worth enabling at that point, don't assume it automatically is.
- [ ] Read replica health endpoint — expose sync lag and InSync status
- [ ] Extract User Management to a dedicated identity service when the data layer splits
- [ ] Scalar branding — Switchyard logo and name above the API title; currently blocked by Scalar's limited logo support in the .NET package
- [ ] Equipment relocation / home-warehouse reassignment — `home_warehouse_id` is set once at equipment creation and never mutated anywhere (confirmed via `Resolve()`, which only flips status back to `Available`). Dispatchers currently work around this by reporting a fake breakdown, towing the unit to the desired warehouse, and resolving it there — but the resolve flow doesn't actually update `home_warehouse_id`, so the record still claims its original home base. Extend the "Available" check to validate equipment was made available *at* its recorded `home_warehouse_id`, and/or add a real relocation endpoint that updates the field.
