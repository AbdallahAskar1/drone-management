# Drone Delivery Management

Go backend for coordinating drone-based deliveries. JWT-authenticated REST API with role-based access (admin / enduser / drone), a complete order lifecycle, and first-class handling of the broken-drone handoff scenario.

## Stack

- **Go 1.25** · Echo v5 · GORM · PostgreSQL (pgx driver)
- **Auth**: self-signed HS256 JWT (`golang-jwt/jwt/v5`)
- **Tests**: Go unit tests + Hurl API integration tests

## Quick Start

```bash
# 1. Copy env template (tweak JWT_SECRET for non-dev use)
cp .env.example .env

# 2. Build image and start Postgres + app
make up

# 3. Verify health
curl http://localhost:1323/healthz
# → {"status":"ok"}
```

> **Docker API version**: if you get a "client version too old" error, the Makefile
> already sets `DOCKER_API_VERSION=1.44` automatically. No manual fix needed.

## Running Tests

```bash
# Unit tests (no DB required)
make test

# Hurl API tests (requires running stack)
make hurl-reset   # wipes DB and restarts containers
make hurl         # runs all 9 flows sequentially, 90 requests total

# Individual flows
make hurl-auth
make hurl-enduser
make hurl-drone-happy
make hurl-drone-failed
make hurl-drone-broken
make hurl-handoff
make hurl-heartbeat
make hurl-admin
make hurl-rbac
```

## curl Walkthrough

The full happy path: mint tokens → enduser submits → drone picks up → drone heartbeats → drone delivers.

```bash
BASE=http://localhost:1323

# 1. Mint tokens
USER_TOK=$(curl -s -X POST $BASE/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"name":"alice","role":"enduser"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

DRONE_TOK=$(curl -s -X POST $BASE/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"name":"drone-1","role":"drone"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# 2. Enduser submits an order
ORDER=$(curl -s -X POST $BASE/orders \
  -H "Authorization: Bearer $USER_TOK" \
  -H 'Content-Type: application/json' \
  -d '{"origin":{"lat":30.0444,"lng":31.2357},"destination":{"lat":30.0626,"lng":31.2497}}')
echo $ORDER   # status: READY_FOR_PICKUP
ORDER_ID=$(echo $ORDER | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)

# 3. Drone lists open jobs and reserves one
JOB_ID=$(curl -s "$BASE/drone/jobs?type=ORIGIN_PICKUP" \
  -H "Authorization: Bearer $DRONE_TOK" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)

curl -s -X POST $BASE/drone/jobs/$JOB_ID/reserve \
  -H "Authorization: Bearer $DRONE_TOK"   # job: RESERVED, order: RESERVED

# 4. Drone picks up
curl -s -X POST $BASE/drone/orders/$ORDER_ID/pickup \
  -H "Authorization: Bearer $DRONE_TOK"   # order: PICKED_UP

# 5. Drone sends heartbeat (updates location + ETA on the order)
curl -s -X POST $BASE/drone/self/heartbeat \
  -H "Authorization: Bearer $DRONE_TOK" \
  -H 'Content-Type: application/json' \
  -d '{"lat":30.0530,"lng":31.2420}'

# 6. Drone delivers
curl -s -X POST $BASE/drone/orders/$ORDER_ID/delivered \
  -H "Authorization: Bearer $DRONE_TOK"   # order: DELIVERED

# 7. Enduser checks final state + timeline
curl -s $BASE/orders/$ORDER_ID -H "Authorization: Bearer $USER_TOK"
```

### Broken-drone handoff

```bash
ADMIN_TOK=$(curl -s -X POST $BASE/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"name":"ops","role":"admin"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

DRONE2_TOK=$(curl -s -X POST $BASE/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"name":"drone-2","role":"drone"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# Drone-1 picks up (steps 1-4 above), then breaks mid-delivery:
curl -s -X POST $BASE/drone/self/broken -H "Authorization: Bearer $DRONE_TOK"
# → order: HANDOFF_REQUIRED, new HANDOFF_PICKUP job created

# Admin fixes drone-1 — handoff job is deliberately NOT cancelled:
DRONE1_ID=$(curl -s $BASE/admin/drones \
  -H "Authorization: Bearer $ADMIN_TOK" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
curl -s -X POST $BASE/admin/drones/$DRONE1_ID/fixed -H "Authorization: Bearer $ADMIN_TOK"

# Drone-2 claims the handoff job:
HO_JOB=$(curl -s "$BASE/drone/jobs?type=HANDOFF_PICKUP" \
  -H "Authorization: Bearer $DRONE2_TOK" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
curl -s -X POST $BASE/drone/jobs/$HO_JOB/reserve -H "Authorization: Bearer $DRONE2_TOK"
curl -s -X POST $BASE/drone/orders/$ORDER_ID/pickup  -H "Authorization: Bearer $DRONE2_TOK"
curl -s -X POST $BASE/drone/orders/$ORDER_ID/delivered -H "Authorization: Bearer $DRONE2_TOK"
```

## API Reference

### Public
| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/token` | Mint JWT — body `{name, role}` |
| GET | `/healthz` | Health check |

### Enduser (`role=enduser`)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/orders` | Submit order — body `{origin:{lat,lng}, destination:{lat,lng}}` |
| GET | `/orders/{id}` | Get own order + event timeline |
| POST | `/orders/{id}/withdraw` | Withdraw (only before PICKED_UP) |

### Drone (`role=drone`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/drone/jobs` | List OPEN jobs — optional `?type=ORIGIN_PICKUP\|HANDOFF_PICKUP` |
| POST | `/drone/jobs/{id}/reserve` | Reserve a job |
| POST | `/drone/orders/{id}/pickup` | Collect goods |
| POST | `/drone/orders/{id}/delivered` | Mark delivered |
| POST | `/drone/orders/{id}/failed` | Mark failed |
| POST | `/drone/self/broken` | Self-report broken (triggers handoff if carrying) |
| POST | `/drone/self/heartbeat` | Update location — body `{lat, lng}` |
| GET | `/drone/self/order` | Get currently assigned order |

### Admin (`role=admin`)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/admin/orders` | List orders — `?ids=&status=&limit=&offset=` |
| PATCH | `/admin/orders/{id}` | Patch origin/destination (before PICKED_UP) |
| GET | `/admin/drones` | List all drones |
| POST | `/admin/drones/{id}/broken` | Mark drone broken |
| POST | `/admin/drones/{id}/fixed` | Mark drone available (handoff job stays OPEN) |

## Architecture

```
cmd/server          → wire config, db, router, server
internal/config     → env loading
internal/database   → GORM open + AutoMigrate + WithTx helper
internal/domain     → typed enums, state machines, custom errors (no GORM)
internal/repo       → GORM models + persistence; optimistic-lock updates
internal/service    → business rules, RBAC guards, state transitions
internal/handler    → Echo handlers, request/response DTOs, validation
internal/middleware → JWT auth, role guard, request-ID, recovery
internal/utils      → JWT signer, haversine/ETA, clock interface
tests/              → Go integration tests (httptest + Postgres)
tests/hurl/         → Hurl API flow tests (01–09)
```

## State Machines

**Order**: `READY_FOR_PICKUP → RESERVED → PICKED_UP → DELIVERED | FAILED`
- `PICKED_UP → HANDOFF_REQUIRED` when the carrying drone breaks
- `READY_FOR_PICKUP | CREATED → WITHDRAWN` on enduser request

**Drone**: `AVAILABLE ↔ BUSY` on reserve/complete · `* → BROKEN` · `BROKEN → AVAILABLE` (admin only)

**Key invariant**: Admin marking a drone `AVAILABLE` after it breaks mid-delivery does **not** cancel the pending `HANDOFF_PICKUP` job — another drone can still claim it.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `1323` | HTTP listen port |
| `DATABASE_URL` | — | Postgres DSN |
| `JWT_SECRET` | — | HMAC signing secret |
| `JWT_TTL` | `24h` | Token lifetime |
| `AVG_SPEED_MS` | `10` | Average drone speed (m/s) for ETA |
| `LOG_LEVEL` | `info` | `debug\|info\|warn\|error` |
