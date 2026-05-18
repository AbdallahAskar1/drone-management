# ── 09-rbac.hurl ──────────────────────────────────────────────────────────────
# Tests: role-based access control enforcement.
#   Wrong role → 403. No token → 401.

POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-user-09", "role": "enduser"}

HTTP 200
[Captures]
user_token: jsonpath "$.token"


POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-drone-09", "role": "drone"}

HTTP 200
[Captures]
drone_token: jsonpath "$.token"


POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-admin-09", "role": "admin"}

HTTP 200
[Captures]
admin_token: jsonpath "$.token"


# ── No token → 401 ────────────────────────────────────────────────────────────

GET http://localhost:1323/orders/1

HTTP 401


GET http://localhost:1323/drone/jobs

HTTP 401


GET http://localhost:1323/admin/orders

HTTP 401


# ── Drone token on enduser routes → 403 ───────────────────────────────────────

POST http://localhost:1323/orders
Authorization: Bearer {{drone_token}}
Content-Type: application/json
{
  "origin": {"lat": 0, "lng": 0},
  "destination": {"lat": 1, "lng": 1}
}

HTTP 403


# ── Enduser token on drone routes → 403 ───────────────────────────────────────

GET http://localhost:1323/drone/jobs
Authorization: Bearer {{user_token}}

HTTP 403


POST http://localhost:1323/drone/self/heartbeat
Authorization: Bearer {{user_token}}
Content-Type: application/json
{"lat": 0, "lng": 0}

HTTP 403


POST http://localhost:1323/drone/self/broken
Authorization: Bearer {{user_token}}

HTTP 403


# ── Enduser / drone token on admin routes → 403 ───────────────────────────────

GET http://localhost:1323/admin/orders
Authorization: Bearer {{user_token}}

HTTP 403


GET http://localhost:1323/admin/drones
Authorization: Bearer {{drone_token}}

HTTP 403


POST http://localhost:1323/admin/drones/1/broken
Authorization: Bearer {{user_token}}

HTTP 403


POST http://localhost:1323/admin/drones/1/fixed
Authorization: Bearer {{drone_token}}

HTTP 403


# ── Admin token on enduser/drone routes → 403 ─────────────────────────────────

POST http://localhost:1323/orders
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "origin": {"lat": 0, "lng": 0},
  "destination": {"lat": 1, "lng": 1}
}

HTTP 403


GET http://localhost:1323/drone/jobs
Authorization: Bearer {{admin_token}}

HTTP 403
