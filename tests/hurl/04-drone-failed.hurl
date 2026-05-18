# ── 04-drone-failed.hurl ──────────────────────────────────────────────────────
# Tests: drone marks a delivery as failed.

POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-user-04", "role": "enduser"}

HTTP 200
[Captures]
user_token: jsonpath "$.token"


POST http://localhost:1323/auth/token
Content-Type: application/json
{"name": "hurl-drone-04", "role": "drone"}

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
[Asserts]
jsonpath "$.status" == "PICKED_UP"


# Mark failed
POST http://localhost:1323/drone/orders/{{order_id}}/failed
Authorization: Bearer {{drone_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "FAILED"
jsonpath "$.failed_at" isString


GET http://localhost:1323/orders/{{order_id}}
Authorization: Bearer {{user_token}}

HTTP 200
[Asserts]
jsonpath "$.status" == "FAILED"
