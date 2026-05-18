# ── 03-drone-happy-path.hurl ──────────────────────────────────────────────────
# Tests: full happy-path delivery lifecycle.
#   enduser submits → drone lists jobs → reserves → picks up → delivers.

POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-user-03", "role": "enduser"}

HTTP 200
[Captures]
user_token: jsonpath "$.token"


POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-drone-03", "role": "drone"}

HTTP 200
[Captures]
drone_token: jsonpath "$.token"


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


# Drone lists open jobs — ours is the only ORIGIN_PICKUP open for this drone
GET http://localhost:1323/drone/jobs?type=ORIGIN_PICKUP
Authorization: Bearer {{drone_token}}

HTTP 200
[Captures]
job_id: jsonpath "$.jobs[-1].id"
[Asserts]
jsonpath "$.jobs" count >= 1
jsonpath "$.jobs[-1].type" == "ORIGIN_PICKUP"
jsonpath "$.jobs[-1].status" == "OPEN"
jsonpath "$.jobs[-1].order_id" == {{order_id}}


# Reserve job
POST http://localhost:1323/drone/jobs/{{job_id}}/reserve
Authorization: Bearer {{drone_token}}

HTTP 200
[Asserts]
jsonpath "$.job.status" == "RESERVED"
jsonpath "$.order.status" == "RESERVED"
jsonpath "$.order.id" == {{order_id}}


# Pick up
POST http://localhost:1323/drone/orders/{{order_id}}/pickup
Authorization: Bearer {{drone_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "PICKED_UP"
jsonpath "$.picked_up_at" isString


# Deliver
POST http://localhost:1323/drone/orders/{{order_id}}/delivered
Authorization: Bearer {{drone_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "DELIVERED"
jsonpath "$.delivered_at" isString


# Enduser verifies final state and timeline
GET http://localhost:1323/orders/{{order_id}}
Authorization: Bearer {{user_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "DELIVERED"
jsonpath "$.timeline" count >= 1
