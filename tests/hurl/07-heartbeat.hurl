# ── 07-heartbeat.hurl ─────────────────────────────────────────────────────────
# Tests: POST /drone/self/heartbeat (updates location + ETA while carrying),
#        GET /drone/self/order.

POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-user-07", "role": "enduser"}

HTTP 200
[Captures]
user_token: jsonpath "$.token"


POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-drone-07", "role": "drone"}

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


GET http://localhost:1323/drone/jobs?type=ORIGIN_PICKUP
Authorization: Bearer {{drone_token}}

HTTP 200
[Captures]
job_id: jsonpath "$.jobs[-1].id"
[Asserts]
jsonpath "$.jobs[-1].order_id" == {{order_id}}


POST http://localhost:1323/drone/jobs/{{job_id}}/reserve
Authorization: Bearer {{drone_token}}

HTTP 200


POST http://localhost:1323/drone/orders/{{order_id}}/pickup
Authorization: Bearer {{drone_token}}

HTTP 200


# Heartbeat — drone is carrying the order; location + ETA should propagate
POST http://localhost:1323/drone/self/heartbeat
Authorization: Bearer {{drone_token}}
Content-Type: application/json
{
  "lat": 30.0500,
  "lng": 31.2400
}

HTTP 200
[Asserts]
jsonpath "$.drone.last_lat" == 30.05
jsonpath "$.drone.last_lng" == 31.24
jsonpath "$.order.id" == {{order_id}}
jsonpath "$.order.current_lat" == 30.05
jsonpath "$.order.current_lng" == 31.24
jsonpath "$.order.eta_seconds" isInteger


# GET /drone/self/order
GET http://localhost:1323/drone/self/order
Authorization: Bearer {{drone_token}}

HTTP 200
[Asserts]
jsonpath "$.order.id" == {{order_id}}
jsonpath "$.order.status" == "PICKED_UP"
jsonpath "$.drone.current_order_id" == {{order_id}}


# Second heartbeat closer to destination — ETA remains an integer
POST http://localhost:1323/drone/self/heartbeat
Authorization: Bearer {{drone_token}}
Content-Type: application/json
{
  "lat": 30.0620,
  "lng": 31.2490
}

HTTP 200
[Asserts]
jsonpath "$.drone.last_lat" == 30.062
jsonpath "$.order.eta_seconds" isInteger
