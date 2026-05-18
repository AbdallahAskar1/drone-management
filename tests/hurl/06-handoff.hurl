# ── 06-handoff.hurl ───────────────────────────────────────────────────────────
# Tests: broken-drone handoff — the headline scenario.
#   drone1 picks up → breaks mid-delivery → HANDOFF_REQUIRED →
#   drone2 claims handoff job → picks up → delivers.
#
# The "admin-fix does NOT cancel handoff" invariant is verified in 08-admin.hurl.

POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-user-06", "role": "enduser"}

HTTP 200
[Captures]
user_token: jsonpath "$.token"


POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-drone-06a", "role": "drone"}

HTTP 200
[Captures]
drone1_token: jsonpath "$.token"


POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-drone-06b", "role": "drone"}

HTTP 200
[Captures]
drone2_token: jsonpath "$.token"


# Submit order
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


# Drone1 reserves the origin-pickup job
GET http://localhost:1323/drone/jobs?type=ORIGIN_PICKUP
Authorization: Bearer {{drone1_token}}

HTTP 200
[Captures]
job_id: jsonpath "$.jobs[-1].id"
[Asserts]
jsonpath "$.jobs[-1].order_id" == {{order_id}}


POST http://localhost:1323/drone/jobs/{{job_id}}/reserve
Authorization: Bearer {{drone1_token}}

HTTP 200
[Asserts]
jsonpath "$.order.status" == "RESERVED"


# Drone1 picks up
POST http://localhost:1323/drone/orders/{{order_id}}/pickup
Authorization: Bearer {{drone1_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "PICKED_UP"


# Drone1 breaks mid-delivery → handoff triggered
POST http://localhost:1323/drone/self/broken
Authorization: Bearer {{drone1_token}}

HTTP 200
[Asserts]
jsonpath "$.drone.status" == "BROKEN"


# Order must be HANDOFF_REQUIRED
GET http://localhost:1323/orders/{{order_id}}
Authorization: Bearer {{user_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "HANDOFF_REQUIRED"
jsonpath "$.handoff_required_at" isString


# A new HANDOFF_PICKUP job must be OPEN
GET http://localhost:1323/drone/jobs?type=HANDOFF_PICKUP
Authorization: Bearer {{drone2_token}}

HTTP 200
[Captures]
handoff_job_id: jsonpath "$.jobs[0].id"
[Asserts]
jsonpath "$.jobs" count >= 1
jsonpath "$.jobs[0].status" == "OPEN"
jsonpath "$.jobs[0].order_id" == {{order_id}}


# Drone2 reserves the handoff job
POST http://localhost:1323/drone/jobs/{{handoff_job_id}}/reserve
Authorization: Bearer {{drone2_token}}

HTTP 200
[Asserts]
jsonpath "$.order.status" == "RESERVED"


# Drone2 picks up
POST http://localhost:1323/drone/orders/{{order_id}}/pickup
Authorization: Bearer {{drone2_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "PICKED_UP"


# Drone2 delivers
POST http://localhost:1323/drone/orders/{{order_id}}/delivered
Authorization: Bearer {{drone2_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "DELIVERED"
jsonpath "$.delivered_at" isString


GET http://localhost:1323/orders/{{order_id}}
Authorization: Bearer {{user_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "DELIVERED"
