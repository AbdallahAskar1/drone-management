# ── 08-admin.hurl ─────────────────────────────────────────────────────────────
# Tests: all admin endpoints + the "admin-fix does NOT cancel handoff" rule.

POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-admin-08", "role": "admin"}

HTTP 200
[Captures]
admin_token: jsonpath "$.token"


POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-user-08", "role": "enduser"}

HTTP 200
[Captures]
user_token: jsonpath "$.token"


POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-drone-08", "role": "drone"}

HTTP 200
[Captures]
drone_token: jsonpath "$.token"


# Submit an order (used throughout this file)
POST http://localhost:1323/orders
Authorization: Bearer {{user_token}}
Content-Type: application/json
{
  "origin":      {"lat": 30.0444, "lng": 31.2357},
  "destination": {"lat": 30.0626, "lng": 31.2497}
}

HTTP 201
[Captures]
order_id: jsonpath "$.id"
[Asserts]
jsonpath "$.status" == "READY_FOR_PICKUP"


# ── GET /admin/orders ──────────────────────────────────────────────────────────

GET http://localhost:1323/admin/orders
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.orders" isCollection
jsonpath "$.orders" count >= 1


# Filter by status
GET http://localhost:1323/admin/orders?status=READY_FOR_PICKUP
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.orders" count >= 1
jsonpath "$.orders[0].status" == "READY_FOR_PICKUP"


# Filter by specific IDs
GET http://localhost:1323/admin/orders?ids={{order_id}}
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.orders" count == 1
jsonpath "$.orders[0].id" == {{order_id}}


# Pagination
GET http://localhost:1323/admin/orders?limit=1&offset=0
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.orders" count == 1


# ── PATCH /admin/orders/{id} ──────────────────────────────────────────────────

PATCH http://localhost:1323/admin/orders/{{order_id}}
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "origin":      {"lat": 31.0, "lng": 32.0},
  "destination": {"lat": 31.5, "lng": 32.5}
}

HTTP 200
[Asserts]
jsonpath "$.origin_lat" == 31.0
jsonpath "$.origin_lng" == 32.0
jsonpath "$.dest_lat" == 31.5
jsonpath "$.dest_lng" == 32.5


# Patch only destination
PATCH http://localhost:1323/admin/orders/{{order_id}}
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "destination": {"lat": 30.0626, "lng": 31.2497}
}

HTTP 200
[Asserts]
jsonpath "$.dest_lat" == 30.0626


# ── GET /admin/drones ─────────────────────────────────────────────────────────

GET http://localhost:1323/admin/drones
Authorization: Bearer {{admin_token}}

HTTP 200
[Captures]
drone_id: jsonpath "$.drones[-1].id"
[Asserts]
jsonpath "$.drones" isCollection
jsonpath "$.drones" count >= 1


# ── POST /admin/drones/{id}/broken then /fixed ────────────────────────────────

POST http://localhost:1323/admin/drones/{{drone_id}}/broken
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.drone.status" == "BROKEN"
jsonpath "$.drone.id" == {{drone_id}}


POST http://localhost:1323/admin/drones/{{drone_id}}/fixed
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.drone.status" == "AVAILABLE"
jsonpath "$.drone.id" == {{drone_id}}


# ── Verify: admin-fix does NOT cancel an open handoff job ─────────────────────
# Flow: drone reserves → picks up → breaks → handoff job created →
#       admin fixes drone → handoff job is STILL OPEN.

GET http://localhost:1323/drone/jobs?type=ORIGIN_PICKUP
Authorization: Bearer {{drone_token}}

HTTP 200
[Captures]
origin_job_id: jsonpath "$.jobs[-1].id"
[Asserts]
jsonpath "$.jobs[-1].order_id" == {{order_id}}


POST http://localhost:1323/drone/jobs/{{origin_job_id}}/reserve
Authorization: Bearer {{drone_token}}

HTTP 200


POST http://localhost:1323/drone/orders/{{order_id}}/pickup
Authorization: Bearer {{drone_token}}

HTTP 200


# Drone breaks while carrying
POST http://localhost:1323/drone/self/broken
Authorization: Bearer {{drone_token}}

HTTP 200
[Asserts]
jsonpath "$.drone.status" == "BROKEN"


# Handoff job created
GET http://localhost:1323/drone/jobs?type=HANDOFF_PICKUP
Authorization: Bearer {{drone_token}}

HTTP 200
[Captures]
handoff_job_id: jsonpath "$.jobs[0].id"
[Asserts]
jsonpath "$.jobs" count >= 1
jsonpath "$.jobs[0].order_id" == {{order_id}}
jsonpath "$.jobs[0].status" == "OPEN"


# Admin fixes the broken drone
POST http://localhost:1323/admin/drones/{{drone_id}}/fixed
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.drone.status" == "AVAILABLE"


# Critical: handoff job must STILL be OPEN after the fix
GET http://localhost:1323/drone/jobs?type=HANDOFF_PICKUP
Authorization: Bearer {{drone_token}}

HTTP 200
[Asserts]
jsonpath "$.jobs[0].id" == {{handoff_job_id}}
jsonpath "$.jobs[0].status" == "OPEN"
